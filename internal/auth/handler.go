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
	req, err := parseAndValidateAuthRequest(r)
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

	resp := &timetrackapi.TokenResponse{AccessToken: t.AccessToken, TokenType: timetrackapi.Bearer}
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

func parseAndValidateAuthRequest(r *http.Request) (*timetrackapi.AuthRequest, error) {
	req, err := parseAuthRequest(r)
	if err != nil {
		return nil, err
	}
	if err = validateAuthRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

func parseAuthRequest(r *http.Request) (*timetrackapi.AuthRequest, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errors.Join(apiutil.ValidationError{"bad form"}, err)
	}

	grantType := timetrackapi.Password
	if r.FormValue("grant_type") != string(grantType) {
		return nil, apiutil.ValidationError{"unsupported grant type"}
	}

	return &timetrackapi.AuthRequest{
		GrantType: grantType,
		Username:  r.FormValue("username"),
		Password:  r.FormValue("password"),
	}, nil
}

func validateAuthRequest(req *timetrackapi.AuthRequest) error {
	e := make([]string, 0)

	if req.Username == "" {
		e = append(e, "missing username")
	}
	if req.Password == "" {
		e = append(e, "missing password")
	}

	if len(e) > 0 {
		return apiutil.ValidationError(e)
	}
	return nil
}
