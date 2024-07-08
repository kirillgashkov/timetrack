package api

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/api/request"
	"github.com/kirillgashkov/timetrack/internal/api/response"
	"github.com/kirillgashkov/timetrack/internal/config"
	"github.com/kirillgashkov/timetrack/internal/user"
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

	response.MustWriteJSON(w, userToAPI(u), http.StatusOK)
}

func (si *serverInterface) GetUsers(w http.ResponseWriter, r *http.Request, params timetrackapi.GetUsersParams) {
	filter := &user.Filter{}
	if params.Filter != nil {
		for _, f := range *params.Filter {
			parts := strings.SplitN(f, "=", 2)
			if len(parts) != 2 {
				response.MustWriteError(w, "invalid filter", http.StatusUnprocessableEntity)
				return
			}
			k, v := parts[0], parts[1]

			switch k {
			case "passport_number":
				filter.PassportNumber = &v
			case "surname":
				filter.Surname = &v
			case "name":
				filter.Name = &v
			case "patronymic":
				filter.Patronymic = &v
			case "address":
				filter.Address = &v
			default:
				response.MustWriteError(w, "invalid filter", http.StatusUnprocessableEntity)
			}
		}
	}
	offset := 0
	if params.Offset != nil {
		if *params.Offset < 0 {
			response.MustWriteError(w, "invalid offset", http.StatusUnprocessableEntity)
			return
		}
		offset = *params.Offset
	}
	limit := 50
	if params.Limit != nil {
		if *params.Limit < 1 || *params.Limit > 100 {
			response.MustWriteError(w, "invalid limit", http.StatusUnprocessableEntity)
			return
		}
		limit = *params.Limit
	}

	users, err := si.user.List(r.Context(), filter, offset, limit)
	if err != nil {
		slog.Error("failed to get users", "error", err)
		response.MustWriteInternalServerError(w)
		return
	}

	apiUsers := make([]*timetrackapi.User, 0, len(users))
	for _, u := range users {
		apiUsers = append(apiUsers, userToAPI(&u))
	}
	response.MustWriteJSON(w, apiUsers, http.StatusOK)
}

func (si *serverInterface) GetUsersCurrent(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}

func (si *serverInterface) PatchUsersPassportNumber(w http.ResponseWriter, r *http.Request, passportNumber string) {
	var userUpdate *timetrackapi.UserUpdate
	if err := request.ReadJSON(r, &userUpdate); err != nil {
		response.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	u, err := si.user.UpdateByPassportNumber(r.Context(), passportNumber, &user.Update{
		Surname:    userUpdate.Surname,
		Name:       userUpdate.Name,
		Patronymic: userUpdate.Patronymic,
		Address:    userUpdate.Address,
	})
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			response.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to update user", "error", err)
		response.MustWriteInternalServerError(w)
		return
	}

	response.MustWriteJSON(w, userToAPI(u), http.StatusOK)
}

func (si *serverInterface) DeleteUsersPassportNumber(w http.ResponseWriter, r *http.Request, passportNumber string) {
	u, err := si.user.DeleteByPassportNumber(r.Context(), passportNumber)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			response.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to delete user", "error", err)
		response.MustWriteInternalServerError(w)
		return
	}

	response.MustWriteJSON(w, userToAPI(u), http.StatusOK)
}

func (si *serverInterface) GetUsersPassportNumber(w http.ResponseWriter, r *http.Request, passportNumber string) {
	u, err := si.user.GetByPassportNumber(r.Context(), passportNumber)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			response.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get user", "error", err)
		response.MustWriteInternalServerError(w)
		return
	}

	response.MustWriteJSON(w, userToAPI(u), http.StatusOK)
}

func userToAPI(u *user.User) *timetrackapi.User {
	return &timetrackapi.User{
		PassportNumber: u.PassportNumber,
		Surname:        u.Surname,
		Name:           u.Name,
		Patronymic:     u.Patronymic,
		Address:        u.Address,
	}
}
