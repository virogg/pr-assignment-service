package e2e

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/virogg/pr-assignment-service/internal/application/services/pr_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/stats_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/team_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/user_service"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/repository/postgres"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/repository/postgres/test_helpers"
	httpTransport "github.com/virogg/pr-assignment-service/internal/infrastructure/transport/http"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	trm "github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/stretchr/testify/require"
)

type TestServer struct {
	Server  *http.Server
	BaseURL string
	DB      *test_helpers.TestDatabase
	cleanup func()
}

func SetupTestServer(t *testing.T) *TestServer {
	t.Helper()

	testDB := test_helpers.SetupTestDatabase(t)

	opts := &slog.HandlerOptions{Level: slog.LevelError}
	log := slog.New(slog.NewTextHandler(os.Stdout, opts))

	getter := trmpgx.DefaultCtxGetter
	trManager := trm.Must(trmpgx.NewDefaultFactory(testDB.Pool))

	teamRepo := postgres.NewTeamPostgresRepository(testDB.Pool, getter)
	userRepo := postgres.NewUserPostgresRepository(testDB.Pool, getter)
	prRepo := postgres.NewPRPostgresRepository(testDB.Pool, getter)
	statsRepo := postgres.NewStatsPostgresRepository(testDB.Pool, getter)

	teamSvc := team_service.New(userRepo, teamRepo, trManager, log)
	userSvc := user_service.New(userRepo, teamRepo, prRepo, trManager, log)
	prSvc := pr_service.New(userRepo, teamRepo, prRepo, trManager, log)
	statsSvc := stats_service.New(statsRepo, log)

	router := httpTransport.NewRouter(prSvc, statsSvc, teamSvc, userSvc, log)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	//nolint:gosec // `Potential Slowloris Attack` is not a problem here: server is internal
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", slog.Any("error", err))
		}
	}()

	baseURL := fmt.Sprintf("http://%s", addr)
	for i := 0; i < 50; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			break
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(10 * time.Millisecond)
	}

	ts := &TestServer{
		Server:  server,
		BaseURL: baseURL,
		DB:      testDB,
		cleanup: func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
			testDB.Cleanup(t)
		},
	}

	return ts
}

func (ts *TestServer) Cleanup(t *testing.T) {
	t.Helper()
	if ts.cleanup != nil {
		ts.cleanup()
	}
}

func (ts *TestServer) ResetDatabase(t *testing.T) {
	t.Helper()
	ts.DB.TruncateTables(t)
}
