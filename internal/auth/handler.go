package auth

import (
	"errors"
	"net/http"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// PostAuth handles "POST /auth".
func (h *Handler) PostAuth(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		apiutil.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}
	grantType := r.FormValue("grant_type")
	username := r.FormValue("username")
	password := r.FormValue("password")

	if grantType != string(timetrackapi.Password) {
		apiutil.MustWriteError(w, "unsupported grant type", http.StatusUnprocessableEntity)
		return
	}
	if username == "" {
		apiutil.MustWriteError(w, "missing username", http.StatusUnprocessableEntity)
		return
	}
	if password == "" {
		apiutil.MustWriteError(w, "missing password", http.StatusUnprocessableEntity)
		return
	}

	token, err := h.service.Authorize(
		r.Context(),
		&PasswordGrant{
			Username: username,
			Password: password,
		},
	)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			apiutil.MustWriteError(w, "invalid credentials", http.StatusBadRequest)
			return
		}
		apiutil.MustWriteInternalServerError(w)
		return
	}

	tokenAPI := &timetrackapi.Token{
		AccessToken: token.AccessToken,
		TokenType:   timetrackapi.Bearer,
	}
	apiutil.MustWriteJSON(w, tokenAPI, http.StatusOK)
}
