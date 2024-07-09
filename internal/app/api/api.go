package api

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/kirillgashkov/timetrack/internal/app/config"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/user"
)

func NewServer(cfg *config.ServerConfig, userService *user.Service) (*http.Server, error) {
	h, err := newHandler(userService)
	if err != nil {
		return nil, errors.Join(errors.New("failed to create handler"), err)
	}

	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           h,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}, nil
}

func newHandler(userService *user.Service) (http.Handler, error) {
	si := &user.Handler{Service: userService}
	mux := http.NewServeMux()
	return timetrackapi.HandlerFromMux(si, mux), nil
}
