package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Middleware struct {
	service Service
}

func NewMiddleware(service Service) *Middleware {
	return &Middleware{service: service}
}

func (m *Middleware) Authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := ParseAccessToken(r)
		if err != nil {
			var ve apiutil.ValidationError
			if errors.As(err, &ve) {
				apiutil.MustWriteUnprocessableEntity(w, ve)
				return
			}
			apiutil.MustWriteInternalServerError(w, "failed to parse access token", err)
			return
		}

		u, err := m.service.UserFromAccessToken(t)
		if err != nil {
			if errors.Is(err, ErrInvalidAccessToken) {
				apiutil.MustWriteUnauthorized(w, "invalid access token")
				return
			}
			apiutil.MustWriteInternalServerError(w, "failed to get user from access token", err)
			return
		}

		ctx := ContextWithUser(r.Context(), u)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func ParseAccessToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", apiutil.ValidationError{"missing Authorization header"}
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", apiutil.ValidationError{"invalid Authorization header, expected Bearer token"}
	}

	accessToken := parts[1]
	return accessToken, nil
}

type userContextKeyType struct{}

var userContextKey = userContextKeyType{}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func MustUserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		panic("user not found in context")
	}
	return user
}
