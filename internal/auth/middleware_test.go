package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthenticated(t *testing.T) {
	tests := []struct {
		name                    string
		authHeader              *string
		userFromAccessTokenFunc func(string) (*User, error)
		expectedStatusCode      int
		expectedBody            string
	}{
		{
			"missing Authorization header",
			nil,
			nil,
			http.StatusUnprocessableEntity,
			`{"message":"missing Authorization header"}`,
		},
		{
			"invalid Authorization header",
			stringPtr("Basic token"),
			nil,
			http.StatusUnprocessableEntity,
			`{"message":"invalid Authorization header, expected Bearer token"}`,
		},
		{
			"invalid access token",
			stringPtr("Bearer invalid_token"),
			func(token string) (*User, error) {
				return nil, ErrInvalidAccessToken
			},
			http.StatusUnauthorized,
			`{"message":"invalid access token"}`,
		},
		{
			"success",
			stringPtr("Bearer valid_token"),
			func(token string) (*User, error) {
				return &User{ID: 1}, nil
			},
			http.StatusOK,
			"success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &ServiceMock{
				UserFromAccessTokenFunc: tt.userFromAccessTokenFunc,
			}
			middleware := NewMiddleware(mockService)
			handler := middleware.Authenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("success"))
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != nil {
				req.Header.Set("Authorization", *tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("expected body to contain %s, got %s", tt.expectedBody, body)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
