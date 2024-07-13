package auth

import (
	"errors"
	"net/http"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// PostAuth handles "POST /auth".
func (h *Handler) PostAuth(w http.ResponseWriter, r *http.Request) {
	req, err := parseAndValidatePasswordGrant(r)
	if err != nil {
		var ve apiutil.ValidationError
		if errors.As(err, &ve) {
			apiutil.MustWriteUnprocessableEntity(w, ve)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to parse and validate request", err)
		return
	}

	g := &PasswordGrant{Username: req.Username, Password: req.Password}
	t, err := h.service.Authorize(r.Context(), g)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			apiutil.MustWriteError(w, "invalid credentials", http.StatusBadRequest)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to authorize", err)
		return
	}

	resp := &timetrackapi.Token{AccessToken: t.AccessToken, TokenType: timetrackapi.Bearer}
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

func parseAndValidatePasswordGrant(r *http.Request) (*timetrackapi.PasswordGrant, error) {
	req, err := parsePasswordGrant(r)
	if err != nil {
		return nil, err
	}
	if err = validatePasswordGrant(req); err != nil {
		return nil, err
	}
	return req, nil
}

func parsePasswordGrant(r *http.Request) (*timetrackapi.PasswordGrant, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errors.Join(apiutil.ValidationError{"bad form"}, err)
	}

	grantType := timetrackapi.Password
	if r.FormValue("grant_type") != string(grantType) {
		return nil, apiutil.ValidationError{"unsupported grant type"}
	}

	return &timetrackapi.PasswordGrant{
		GrantType: grantType,
		Username:  r.FormValue("username"),
		Password:  r.FormValue("password"),
	}, nil
}

func validatePasswordGrant(req *timetrackapi.PasswordGrant) error {
	m := make([]string, 0)
	if req.Username == "" {
		m = append(m, "missing username")
	}
	if req.Password == "" {
		m = append(m, "missing password")
	}

	if len(m) > 0 {
		return apiutil.ValidationError(m)
	}
	return nil
}
