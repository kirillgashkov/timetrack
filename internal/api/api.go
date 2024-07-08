package api

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/kirillgashkov/assignment-timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/request"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/response"
	"github.com/kirillgashkov/assignment-timetrack/internal/config"
	"github.com/kirillgashkov/assignment-timetrack/internal/user"
)

type serverInterface struct {
	user *user.Service
}

func NewServer(cfg *config.ServerConfig, user *user.Service) (*http.Server, error) {
	h, err := newHandler(user)
	if err != nil {
		return nil, errors.Join(errors.New("failed to create handler"), err)
	}

	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           h,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}, nil
}

func newHandler(user *user.Service) (http.Handler, error) {
	si := &serverInterface{user: user}
	mux := http.NewServeMux()
	return timetrackapi.HandlerFromMux(si, mux), nil
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
	if userCreate.PassportNumber == "" {
		response.MustWriteError(w, "missing passport number", http.StatusUnprocessableEntity)
		return
	}

	u, err := si.user.Create(r.Context(), userCreate.PassportNumber)
	if err != nil {
		if errors.Is(err, user.ErrAlreadyExists) {
			response.MustWriteError(w, "user already exists", http.StatusBadRequest)
			return
		}
		slog.Error("failed to create user", "error", err)
		response.MustWriteInternalServerError(w)
		return
	}

	uOut := &timetrackapi.User{
		PassportNumber: u.PassportNumber,
		Surname:        u.Surname,
		Name:           u.Name,
		Patronymic:     u.Patronymic,
		Address:        u.Address,
	}
	response.MustWriteJSON(w, uOut, http.StatusOK)
}

func (si *serverInterface) GetUsersCurrent(http.ResponseWriter, *http.Request) {
	panic("implement me")
}
