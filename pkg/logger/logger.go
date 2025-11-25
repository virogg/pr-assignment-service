package logger

import (
	"log/slog"
	"os"
)

func New(logLvl string) *slog.Logger {
	var log *slog.Logger

	opts := &slog.HandlerOptions{AddSource: true}

	switch logLvl {
	case "local":
		opts.Level = slog.LevelDebug
		log = slog.New(slog.NewTextHandler(os.Stdout, opts))
	case "dev":
		opts.Level = slog.LevelDebug
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	case "prod":
		opts.Level = slog.LevelInfo
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	default: // default value is set to "dev" in config
	}

	return log
}
