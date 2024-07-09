package user

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
)

type Handler struct {
	Service *Service
}

func (h *Handler) PostUsers(w http.ResponseWriter, r *http.Request) {
	var userCreate *timetrackapi.UserCreate
	if err := apiutil.ReadJSON(r, &userCreate); err != nil {
		apiutil.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}
	if userCreate.PassportNumber == "" {
		apiutil.MustWriteError(w, "missing passport number", http.StatusUnprocessableEntity)
		return
	}

	u, err := h.Service.Create(r.Context(), userCreate.PassportNumber)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			apiutil.MustWriteError(w, "user already exists", http.StatusBadRequest)
			return
		}
		slog.Error("failed to create user", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, userToAPI(u), http.StatusOK)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request, params timetrackapi.GetUsersParams) {
	filter := &Filter{}
	if params.Filter != nil {
		for _, f := range *params.Filter {
			parts := strings.SplitN(f, "=", 2)
			if len(parts) != 2 {
				apiutil.MustWriteError(w, "invalid filter", http.StatusUnprocessableEntity)
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
				apiutil.MustWriteError(w, "invalid filter", http.StatusUnprocessableEntity)
			}
		}
	}
	offset := 0
	if params.Offset != nil {
		if *params.Offset < 0 {
			apiutil.MustWriteError(w, "invalid offset", http.StatusUnprocessableEntity)
			return
		}
		offset = *params.Offset
	}
	limit := 50
	if params.Limit != nil {
		if *params.Limit < 1 || *params.Limit > 100 {
			apiutil.MustWriteError(w, "invalid limit", http.StatusUnprocessableEntity)
			return
		}
		limit = *params.Limit
	}

	users, err := h.Service.List(r.Context(), filter, offset, limit)
	if err != nil {
		slog.Error("failed to get users", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiUsers := make([]*timetrackapi.User, 0, len(users))
	for _, u := range users {
		apiUsers = append(apiUsers, userToAPI(&u))
	}
	apiutil.MustWriteJSON(w, apiUsers, http.StatusOK)
}

func (h *Handler) GetUsersCurrent(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}

func (h *Handler) PatchUsersPassportNumber(w http.ResponseWriter, r *http.Request, passportNumber string) {
	var userUpdate *timetrackapi.UserUpdate
	if err := apiutil.ReadJSON(r, &userUpdate); err != nil {
		apiutil.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	u, err := h.Service.UpdateByPassportNumber(r.Context(), passportNumber, &Update{
		Surname:    userUpdate.Surname,
		Name:       userUpdate.Name,
		Patronymic: userUpdate.Patronymic,
		Address:    userUpdate.Address,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to update user", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, userToAPI(u), http.StatusOK)
}

func (h *Handler) DeleteUsersPassportNumber(w http.ResponseWriter, r *http.Request, passportNumber string) {
	u, err := h.Service.DeleteByPassportNumber(r.Context(), passportNumber)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to delete user", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, userToAPI(u), http.StatusOK)
}

func (h *Handler) GetUsersPassportNumber(w http.ResponseWriter, r *http.Request, passportNumber string) {
	u, err := h.Service.GetByPassportNumber(r.Context(), passportNumber)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get user", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, userToAPI(u), http.StatusOK)
}

func userToAPI(u *User) *timetrackapi.User {
	return &timetrackapi.User{
		Id:             u.ID,
		PassportNumber: u.PassportNumber,
		Surname:        u.Surname,
		Name:           u.Name,
		Patronymic:     u.Patronymic,
		Address:        u.Address,
	}
}
