package application

import (
	"context"
	"log/slog"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/virogg/pr-assignment-service/internal/application/services/pr_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/stats_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/team_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/user_service"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/repository/postgres"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/transport/http"
	"github.com/virogg/pr-assignment-service/pkg/config"
	pkgpg "github.com/virogg/pr-assignment-service/pkg/postgres"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	server *http.Server
	pool   *pgxpool.Pool
	log    *slog.Logger
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
	pool, err := pkgpg.NewPool(ctx, cfg.DB.GetDSN(), log)
	if err != nil {
		log.Error("Failed to connect to database", slog.Any("error", err))
		return nil, err
	}

	c := trmpgx.DefaultCtxGetter
	trManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	prRepo := postgres.NewPRPostgresRepository(pool, c)
	statsRepo := postgres.NewStatsPostgresRepository(pool, c)
	teamRepo := postgres.NewTeamPostgresRepository(pool, c)
	userRepo := postgres.NewUserPostgresRepository(pool, c)

	prService := pr_service.New(userRepo, teamRepo, prRepo, trManager, log)
	teamService := team_service.New(userRepo, teamRepo, trManager, log)
	userService := user_service.New(userRepo, teamRepo, prRepo, trManager, log)
	statsService := stats_service.New(statsRepo, log)

	server := http.NewServer(log, cfg.ServerPort, prService, statsService, teamService, userService)

	return &App{
		server: server,
		pool:   pool,
		log:    log,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	err := a.server.Start(ctx)
	if err != nil {
		a.log.Error("Server error", slog.Any("error", err))
	}

	a.log.Info("Shutting down application")
	return nil
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		a.log.Error("failed to run application", slog.Any("error", err))
		panic(err)
	}
}

func (a *App) Close() {
	a.log.Info("Closing database connection pool")
	a.pool.Close()
}
