package main

import (
	"context"
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

	app := application.Must(ctx, cfg, log)
	defer app.Close()

	log.Info("Starting PR service")

	app.MustRun(ctx)

	log.Info("PR service stopped successfully")
}
