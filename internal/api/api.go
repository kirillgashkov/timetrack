package api

import (
	"net"
	"net/http"
	"time"

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
	mux := http.NewServeMux()
	return mux
}
