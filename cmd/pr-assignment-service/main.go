package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/virogg/pr-assignment-service/internal/application"
	"github.com/virogg/pr-assignment-service/pkg/config"
)

func main() {
	cfg := config.MustLoadConfig()
	log := newLogger(cfg.LogLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := application.NewApp(ctx, cfg, log)
	if err != nil {
		panic(err)
	}
	defer app.Close()

	log.Info("Starting PR service")

	if err := app.Run(ctx); err != nil {
		log.Error("Application error", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("PR service stopped successfully")
}

func newLogger(env string) *slog.Logger {
	var log *slog.Logger

	opts := &slog.HandlerOptions{AddSource: true}

	switch env {
	case "local":
		opts.Level = slog.LevelDebug
		log = slog.New(slog.NewTextHandler(os.Stdout, opts))
	case "dev":
		opts.Level = slog.LevelDebug
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	case "prod":
		opts.Level = slog.LevelInfo
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	default:
		panic("unknown env")
	}

	return log
}
