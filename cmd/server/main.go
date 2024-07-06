package main

import (
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
	cfg, err := config.New()
	if err != nil {
		return err
	}

	logger := logging.NewLogger(cfg)
	slog.SetDefault(logger)

	srv := api.NewServer(cfg)

	slog.Info("starting server", "addr", srv.Addr, "mode", cfg.Mode)
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
