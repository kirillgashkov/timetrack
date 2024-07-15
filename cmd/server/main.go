package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kirillgashkov/timetrack/internal/app/api"
	"github.com/kirillgashkov/timetrack/internal/app/config"
	"github.com/kirillgashkov/timetrack/internal/app/database"
	"github.com/kirillgashkov/timetrack/internal/app/logging"
	"github.com/kirillgashkov/timetrack/internal/auth"
	"github.com/kirillgashkov/timetrack/internal/reporting"
	"github.com/kirillgashkov/timetrack/internal/task"
	"github.com/kirillgashkov/timetrack/internal/tracking"
	"github.com/kirillgashkov/timetrack/internal/user"
)

func main() {
	if err := mainErr(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
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

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		return errors.Join(errors.New("failed to create database pool"), err)
	}
	defer db.Close()

	peopleInfoService, err := user.NewPeopleInfoServiceImpl(cfg.PeopleInfoServerURL, &http.Client{
		Timeout: time.Duration(cfg.PeopleInfoServerTimeout) * time.Second,
	})
	if err != nil {
		return errors.Join(errors.New("failed to create people info"), err)
	}

	authService := auth.NewServiceImpl(db)
	reportingService := reporting.NewServiceImpl(db)
	taskService := task.NewServiceImpl(db)
	trackingService := tracking.NewServiceImpl(db)
	userService := user.NewServiceImpl(db, peopleInfoService)

	srv, err := api.NewServer(
		&cfg.Server, authService, reportingService, taskService, trackingService, userService,
	)
	if err != nil {
		return errors.Join(errors.New("failed to create server"), err)
	}

	slog.Info("starting server", "addr", srv.Addr, "mode", cfg.Mode)
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Join(errors.New("failed to listen and serve"), err)
	}

	return nil
}
