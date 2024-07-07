package logging

import (
	"log/slog"
	"os"

	"github.com/kirillgashkov/assignment-timetrack/internal/config"
)

func NewLogger(cfg *config.Config) *slog.Logger {
	var logger *slog.Logger

	switch cfg.Mode {
	case config.ModeDevelopment:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case config.ModeProduction:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		panic("invalid mode")
	}

	return logger
}
