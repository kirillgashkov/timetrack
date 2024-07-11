package user

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kirillgashkov/timetrack/internal/auth"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// PostUsers handles "POST /users".
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

	u, err := h.service.Create(r.Context(), userCreate.PassportNumber)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			apiutil.MustWriteError(w, "user already exists", http.StatusBadRequest)
			return
		}
		slog.Error("failed to create user", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, toUserAPI(u), http.StatusOK)
}

// GetUsers handles "GET /users".
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

	users, err := h.service.List(r.Context(), filter, offset, limit)
	if err != nil {
		slog.Error("failed to get users", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	usersAPI := make([]*timetrackapi.User, 0, len(users))
	for _, u := range users {
		usersAPI = append(usersAPI, toUserAPI(&u))
	}
	apiutil.MustWriteJSON(w, usersAPI, http.StatusOK)
}

// GetUsersCurrent handles "GET /users/current".
func (h *Handler) GetUsersCurrent(w http.ResponseWriter, r *http.Request) {
	authUser := auth.MustUserFromContext(r.Context())

	u, err := h.service.Get(r.Context(), authUser.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get user", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, toUserAPI(u), http.StatusOK)
}

// GetUsersId handles "GET /users/{id}".
//
//nolint:revive
func (h *Handler) GetUsersId(w http.ResponseWriter, r *http.Request, id int) {
	u, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get user", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, toUserAPI(u), http.StatusOK)
}

// PatchUsersId handles "PATCH /users/{id}".
//
//nolint:revive
func (h *Handler) PatchUsersId(w http.ResponseWriter, r *http.Request, id int) {
	authenticatedUser := auth.MustUserFromContext(r.Context())
	if authenticatedUser.ID != id {
		apiutil.MustWriteForbidden(w)
		return
	}

	var userUpdate *timetrackapi.UserUpdate
	if err := apiutil.ReadJSON(r, &userUpdate); err != nil {
		apiutil.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	u, err := h.service.Update(r.Context(), id, &Update{
		PassportNumber: userUpdate.PassportNumber,
		Surname:        userUpdate.Surname,
		Name:           userUpdate.Name,
		Patronymic:     userUpdate.Patronymic,
		Address:        userUpdate.Address,
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

	apiutil.MustWriteJSON(w, toUserAPI(u), http.StatusOK)
}

// DeleteUsersId handles "DELETE /users/{id}".
//
//nolint:revive
func (h *Handler) DeleteUsersId(w http.ResponseWriter, r *http.Request, id int) {
	authenticatedUser := auth.MustUserFromContext(r.Context())
	if authenticatedUser.ID != id {
		apiutil.MustWriteForbidden(w)
		return
	}

	u, err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to delete user", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, toUserAPI(u), http.StatusOK)
}

// toUserAPI converts a user domain object to a user API object.
//
// If we had user passwords or other sensitive information we would filter it
// from output models. Passport number is considered sensitive information, but
// it is not filtered because it serves as a username (per app requirements).
func toUserAPI(u *User) *timetrackapi.User {
	return &timetrackapi.User{
		Id:             u.ID,
		PassportNumber: u.PassportNumber,
		Surname:        u.Surname,
		Name:           u.Name,
		Patronymic:     u.Patronymic,
		Address:        u.Address,
	}
}
