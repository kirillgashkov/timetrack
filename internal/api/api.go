package api

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/kirillgashkov/assignment-timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/request"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/response"
	"github.com/kirillgashkov/assignment-timetrack/internal/config"
)

func NewServer(ctx context.Context, cfg *config.Config) (*http.Server, error) {
	h, err := newHandler(ctx, cfg)
	if err != nil {
		return nil, errors.Join(errors.New("failed to create handler"), err)
	}
	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           h,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}, nil
}

func newHandler(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	si, err := newServerInterface(ctx, cfg)
	if err != nil {
		return nil, errors.Join(errors.New("failed to create server interface"), err)
	}

	mux := http.NewServeMux()

	return timetrackapi.HandlerFromMux(si, mux), nil
}

type serverInterface struct {
	db *sql.DB
}

func newServerInterface(ctx context.Context, cfg *config.Config) (*serverInterface, error) {
	db, err := newDB(ctx, cfg.DSN)
	if err != nil {
		return nil, errors.Join(errors.New("failed to create database pool"), err)
	}
	return &serverInterface{db: db}, nil
}

func newDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Join(errors.New("failed to open database"), err)
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, errors.Join(errors.New("failed to ping database"), err)
	}
	return db, nil
}

func (si *serverInterface) GetHealth(w http.ResponseWriter, _ *http.Request) {
	response.MustWriteJSON(w, timetrackapi.Health{Status: "ok"}, http.StatusOK)
}

func (si *serverInterface) PostUsers(w http.ResponseWriter, r *http.Request) {
	var userCreate *timetrackapi.UserCreate
	if err := request.ReadJSON(r, &userCreate); err != nil {
		response.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	panic("implement me")
}

func (si *serverInterface) GetUsersCurrent(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
