package api

import (
	"net"
	"net/http"
	"time"

	"github.com/kirillgashkov/timetrack/internal/task"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/config"
	"github.com/kirillgashkov/timetrack/internal/user"
)

func NewServer(
	cfg *config.ServerConfig,
	taskService *task.Service,
	userService *user.Service,
) (*http.Server, error) {
	si := NewHandler(taskService, userService)
	mux := http.NewServeMux()
	h := timetrackapi.HandlerFromMux(si, mux)

	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           h,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}
	return srv, nil
}
