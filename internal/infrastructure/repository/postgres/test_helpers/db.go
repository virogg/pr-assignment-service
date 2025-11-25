package test_helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName = "testdb"
	dbUser = "testuser"
	dbPass = "testpass"
)

type TestDatabase struct {
	Container *postgres.PostgresContainer
	Pool      *pgxpool.Pool
	ConnStr   string
}

func SetupTestDatabase(t *testing.T) *TestDatabase {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	if err := runMigrations(pool, t); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return &TestDatabase{
		Container: pgContainer,
		Pool:      pool,
		ConnStr:   connStr,
	}
}

func (td *TestDatabase) Cleanup(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	if td.Pool != nil {
		td.Pool.Close()
	}

	if td.Container != nil {
		if err := td.Container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}
}

func (td *TestDatabase) TruncateTables(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	queries := []string{
		"TRUNCATE TABLE pr_reviewers CASCADE",
		"TRUNCATE TABLE pull_requests CASCADE",
		"TRUNCATE TABLE users CASCADE",
		"TRUNCATE TABLE teams CASCADE",
	}

	for _, query := range queries {
		if _, err := td.Pool.Exec(ctx, query); err != nil {
			t.Fatalf("failed to truncate tables: %v", err)
		}
	}
}

func runMigrations(pool *pgxpool.Pool, t *testing.T) error {
	t.Helper()

	migrationsDir, err := findMigrationsDir(t)
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}

	connConfig := pool.Config().ConnConfig
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		connConfig.User,
		connConfig.Password,
		connConfig.Host,
		connConfig.Port,
		connConfig.Database,
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open sql connection: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func findMigrationsDir(t testing.TB) (string, error) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := wd

	for {
		try := filepath.Join(dir, "migrations")

		info, err := os.Stat(try)
		if err == nil && info.IsDir() {
			return try, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("migrations directory not found starting from %s", wd)
		}
		dir = parent
	}
}
