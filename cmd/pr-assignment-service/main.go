package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/virogg/pr-assignment-service/internal/application"
	"github.com/virogg/pr-assignment-service/pkg/config"
	"github.com/virogg/pr-assignment-service/pkg/logger"
)

func main() {
	cfg := config.MustLoadConfig()
	log := logger.New(cfg.LogLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := application.New(ctx, cfg, log)
	if err != nil {
		panic(err)
	}
	defer app.Close()

	log.Info("Starting PR service")

	if err := app.Run(ctx); err != nil {
		log.Error("Application error", slog.Any("error", err))
		panic(err)
	}

	log.Info("PR service stopped successfully")
}
