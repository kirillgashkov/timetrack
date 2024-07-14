package user

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kirillgashkov/timetrack/internal/auth"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// PostUsers handles "POST /users".
func (h *Handler) PostUsers(w http.ResponseWriter, r *http.Request) {
	req, err := parseAndValidateCreateUserRequest(r)
	if err != nil {
		var ve apiutil.ValidationError
		if errors.As(err, &ve) {
			apiutil.MustWriteUnprocessableEntity(w, ve)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to parse and validate request", err)
		return
	}

	u, err := h.service.Create(r.Context(), req.PassportNumber)
	if err != nil {
		if errors.Is(err, ErrInvalidPassportNumber) {
			apiutil.MustWriteUnprocessableEntity(w, apiutil.ValidationError{"invalid passport number"})
			return
		}
		if errors.Is(err, ErrPeopleInfoNotFound) {
			apiutil.MustWriteUnprocessableEntity(
				w, apiutil.ValidationError{"passport number not issued, expired, or revoked"},
			)
			return
		}
		if errors.Is(err, ErrPeopleInfoUnavailable) {
			slog.Error("people info service is unavailable", "error", err)
			apiutil.MustWriteError(w, "service unavailable, try again later", http.StatusServiceUnavailable)
			return
		}
		if errors.Is(err, ErrAlreadyExists) {
			apiutil.MustWriteError(w, "user already exists", http.StatusBadRequest)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to create user", err)
		return
	}

	resp := toUserResponse(u)
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

func parseAndValidateCreateUserRequest(r *http.Request) (*timetrackapi.CreateUserRequest, error) {
	var req *timetrackapi.CreateUserRequest
	if err := apiutil.ReadJSON(r, &req); err != nil {
		return nil, errors.Join(apiutil.ValidationError{"bad JSON"}, err)
	}
	if err := validateCreateUserRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

func validateCreateUserRequest(req *timetrackapi.CreateUserRequest) error {
	if req.PassportNumber == "" {
		return apiutil.ValidationError{"missing passport number"}
	}
	return nil
}

// GetUsers handles "GET /users".
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request, params timetrackapi.GetUsersParams) {
	filter, err := parseListUsersRequestFilter(&params)
	if err != nil {
		var ve apiutil.ValidationError
		if errors.As(err, &ve) {
			apiutil.MustWriteUnprocessableEntity(w, ve)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to parse request", err)
		return
	}

	if err = validateAndNormalizeListUsersRequest(&params); err != nil {
		var ve apiutil.ValidationError
		if errors.As(err, &ve) {
			apiutil.MustWriteUnprocessableEntity(w, ve)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed validate and normalize request", err)
		return
	}

	users, err := h.service.List(r.Context(), filter, *params.Offset, *params.Limit)
	if err != nil {
		apiutil.MustWriteInternalServerError(w, "failed to list users", err)
		return
	}

	resp := make([]*timetrackapi.UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toUserResponse(&u))
	}
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

func parseListUsersRequestFilter(params *timetrackapi.GetUsersParams) (*FilterUser, error) {
	if params.Filter == nil {
		return &FilterUser{}, nil
	}

	e := make([]string, 0)

	filter := &FilterUser{}
	for _, f := range *params.Filter {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) != 2 {
			e = append(e, fmt.Sprintf("invalid filter %q, must be in the form of key=value", f))
		}
		k, v := parts[0], parts[1]

		switch k {
		case "passport-number":
			filter.PassportNumber = &v
		case "surname":
			filter.Surname = &v
		case "name":
			filter.Name = &v
		case "patronymic":
			filter.Patronymic = &sql.NullString{String: v, Valid: true}
		case "patronymic-null":
			patronymic := filter.Patronymic
			if patronymic == nil {
				patronymic = &sql.NullString{}
			}

			switch v {
			case "true":
				patronymic.Valid = false
			case "false":
				patronymic.Valid = true
			default:
				e = append(e, fmt.Sprintf("invalid filter %q, value must be 'true' or 'false'", f))
				continue
			}
			filter.Patronymic = patronymic
		case "address":
			filter.Address = &v
		}
	}

	if len(e) > 0 {
		return nil, apiutil.ValidationError(e)
	}
	return filter, nil
}

func validateAndNormalizeListUsersRequest(params *timetrackapi.GetUsersParams) error {
	if err := validateListUsersRequest(params); err != nil {
		return err
	}
	normalizeListUsersRequest(params)
	return nil
}

func validateListUsersRequest(params *timetrackapi.GetUsersParams) error {
	e := make([]string, 0)

	if params.Offset != nil && *params.Offset < 0 {
		e = append(e, "invalid offset, must be greater than or equal to 0")
	}
	if params.Limit != nil && (*params.Limit < 1 || *params.Limit > 100) {
		e = append(e, "invalid limit, must be between 1 and 100")
	}

	if len(e) > 0 {
		return apiutil.ValidationError(e)
	}
	return nil
}

func normalizeListUsersRequest(params *timetrackapi.GetUsersParams) {
	if params.Offset == nil {
		params.Offset = intPtr(0)
	}
	if params.Limit == nil {
		params.Limit = intPtr(50)
	}
}

// GetUsersCurrent handles "GET /users/current".
func (h *Handler) GetUsersCurrent(w http.ResponseWriter, r *http.Request) {
	currentUser := auth.MustUserFromContext(r.Context())

	u, err := h.service.Get(r.Context(), currentUser.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to get user", err)
		return
	}

	resp := toUserResponse(u)
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
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
		apiutil.MustWriteInternalServerError(w, "failed to get user", err)
		return
	}

	resp := toUserResponse(u)
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

// PatchUsersId handles "PATCH /users/{id}".
//
//nolint:revive
func (h *Handler) PatchUsersId(w http.ResponseWriter, r *http.Request, id int) {
	currentUser := auth.MustUserFromContext(r.Context())
	if currentUser.ID != id {
		apiutil.MustWriteForbidden(w)
		return
	}

	var req *timetrackapi.UpdateUserRequest
	if err := apiutil.ReadJSON(r, &req); err != nil {
		apiutil.MustWriteUnprocessableEntity(w, apiutil.ValidationError{"bad JSON"})
		return
	}

	update := newUpdateFromRequest(req)
	u, err := h.service.Update(r.Context(), id, update)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to update user", err)
		return
	}

	resp := toUserResponse(u)
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

func newUpdateFromRequest(req *timetrackapi.UpdateUserRequest) *UpdateUser {
	var patronymic *sql.NullString
	if req.Patronymic != nil {
		patronymic = &sql.NullString{String: *req.Patronymic, Valid: true}
	}
	if req.PatronymicNull != nil {
		if patronymic == nil {
			patronymic = &sql.NullString{}
		}
		patronymic.Valid = *req.PatronymicNull
	}

	return &UpdateUser{
		PassportNumber: req.PassportNumber,
		Surname:        req.Surname,
		Name:           req.Name,
		Patronymic:     patronymic,
		Address:        req.Address,
	}
}

// DeleteUsersId handles "DELETE /users/{id}".
//
//nolint:revive
func (h *Handler) DeleteUsersId(w http.ResponseWriter, r *http.Request, id int) {
	currentUser := auth.MustUserFromContext(r.Context())
	if currentUser.ID != id {
		apiutil.MustWriteForbidden(w)
		return
	}

	u, err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "user not found", http.StatusNotFound)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to delete user", err)
		return
	}

	resp := toUserResponse(u)
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

// toUserResponse converts User to timetrackapi.UserResponse.
//
// If we had user passwords or other sensitive information we would filter it
// from output models. Passport number is considered sensitive information, but
// it is not filtered because it serves as a username (per assignment
// requirements).
func toUserResponse(u *User) *timetrackapi.UserResponse {
	return &timetrackapi.UserResponse{
		Id:             u.ID,
		PassportNumber: u.PassportNumber,
		Surname:        u.Surname,
		Name:           u.Name,
		Patronymic:     u.Patronymic,
		Address:        u.Address,
	}
}

func intPtr(i int) *int {
	return &i
}
