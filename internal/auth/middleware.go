package auth

import (
	"context"
	"errors"
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
			apiutil.MustWriteUnauthorized(w, "missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			apiutil.MustWriteUnauthorized(w, "invalid Authorization header, expected Bearer token")
			return
		}
		accessToken := parts[1]

		user, err := m.service.UserFromAccessToken(accessToken)
		if err != nil {
			if errors.Is(err, ErrInvalidAccessToken) {
				apiutil.MustWriteUnauthorized(w, "invalid access token")
				return
			}
			apiutil.MustWriteInternalServerError(w, "failed to get user from access token", err)
			return
		}

		ctx := ContextWithUser(r.Context(), user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

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
