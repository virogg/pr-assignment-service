package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/pr-assignment-service/internal/application/services/pr_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/stats_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/team_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/user_service"
)

type Server struct {
	httpServer *http.Server
	port       string
	log        *slog.Logger
}

func NewServer(
	port string,
	prService *pr_service.PRService,
	statsService *stats_service.StatsService,
	teamService *team_service.TeamService,
	userService *user_service.UserService,
	log *slog.Logger,
) *Server {
	router := NewRouter(prService, statsService, teamService, userService, log)

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + port,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		port: port,
		log:  log,
	}
}

func (s *Server) Start(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	serverErrors := make(chan error, 1)

	go func() {
		s.log.Info("Starting server", slog.String("port", s.port))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		s.log.Info("Shutting down courier service server")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown error: %w", err)
		}

		s.log.Info("Server stopped gracefully")
	}

	return nil
}
