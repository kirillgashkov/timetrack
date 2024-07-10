package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Middleware struct {
	service *Service
}

type userContextKeyType struct{}

var userContextKey = userContextKeyType{}

func NewMiddleware(service *Service) *Middleware {
	return &Middleware{service: service}
}

func (m *Middleware) Authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			mustWriteUnauthorized(w, "missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			mustWriteUnauthorized(w, "invalid Authorization header, expected Bearer token")
			return
		}
		accessToken := parts[1]

		user, err := m.service.UserFromAccessToken(accessToken)
		if err != nil {
			if errors.Is(err, ErrInvalidAccessToken) {
				mustWriteUnauthorized(w, "invalid access token")
				return
			}
			slog.Error("failed to get user by access token", "error", err)
			apiutil.MustWriteInternalServerError(w)
			return
		}

		ctx := ContextWithUser(r.Context(), user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func mustWriteUnauthorized(w http.ResponseWriter, m string) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	apiutil.MustWriteError(w, m, http.StatusUnauthorized)
}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}
