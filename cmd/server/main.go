package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/kirillgashkov/assignment-timetrack/internal/api"

	"github.com/kirillgashkov/assignment-timetrack/internal/config"
	"github.com/kirillgashkov/assignment-timetrack/internal/logging"
)

func main() {
	if err := mainErr(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func mainErr() error {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		return errors.Join(errors.New("failed to create config"), err)
	}

	logger := logging.NewLogger(cfg)
	slog.SetDefault(logger)

	srv, err := api.NewServer(ctx, cfg)
	if err != nil {
		return errors.Join(errors.New("failed to create server"), err)
	}

	slog.Info("starting server", "addr", srv.Addr, "mode", cfg.Mode)
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Join(errors.New("failed to listen and serve"), err)
	}

	return nil
}
