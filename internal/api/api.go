package api

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kirillgashkov/assignment-timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/request"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/response"
	"github.com/kirillgashkov/assignment-timetrack/internal/config"
)

func NewServer(cfg *config.Config, db *pgxpool.Pool) (*http.Server, error) {
	h, err := newHandler(db)
	if err != nil {
		return nil, errors.Join(errors.New("failed to create handler"), err)
	}

	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           h,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}, nil
}

func newHandler(db *pgxpool.Pool) (http.Handler, error) {
	si := &serverInterface{db: db}
	mux := http.NewServeMux()
	return timetrackapi.HandlerFromMux(si, mux), nil
}

type serverInterface struct {
	db *pgxpool.Pool
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

	var one int
	if err := si.db.QueryRow(r.Context(), "SELECT 1").Scan(&one); err != nil {
		response.MustWriteInternalServerError(w)
		return
	}

	response.MustWriteJSON(w, timetrackapi.User{"one": one}, http.StatusCreated)
}

func (si *serverInterface) GetUsersCurrent(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
