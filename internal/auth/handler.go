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
	req, err := parseAndValidateRequest(r)
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

func parseAndValidateRequest(r *http.Request) (*request, error) {
	req, err := parseRequest(r)
	if err != nil {
		return nil, err
	}
	if err = req.validate(); err != nil {
		return nil, err
	}
	return req, nil
}

type request struct {
	GrantType string
	Username  string
	Password  string
}

func parseRequest(r *http.Request) (*request, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errors.Join(apiutil.ValidationError([]string{"bad form"}), err)
	}
	grantType := r.FormValue("grant_type")
	username := r.FormValue("username")
	password := r.FormValue("password")

	return &request{GrantType: grantType, Username: username, Password: password}, nil
}

func (r *request) validate() error {
	m := make([]string, 0)
	if r.GrantType != string(timetrackapi.Password) {
		m = append(m, "invalid grant type")
	}
	if r.Username == "" {
		m = append(m, "missing username")
	}
	if r.Password == "" {
		m = append(m, "missing password")
	}

	if len(m) > 0 {
		return apiutil.ValidationError(m)
	}
	return nil
}
