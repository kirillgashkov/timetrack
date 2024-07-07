package api

import (
	"net"
	"net/http"
	"time"

	"github.com/kirillgashkov/assignment-timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/request"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/response"
	"github.com/kirillgashkov/assignment-timetrack/internal/config"
)

func NewServer(cfg *config.Config) http.Server {
	return http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           newHandler(cfg),
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}
}

func newHandler(cfg *config.Config) http.Handler {
	si := newServerInterface(cfg)
	mux := http.NewServeMux()
	return timetrackapi.HandlerFromMux(si, mux)
}

type serverInterface struct {
	dsn string
}

func newServerInterface(cfg *config.Config) *serverInterface {
	return &serverInterface{dsn: ""}
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
