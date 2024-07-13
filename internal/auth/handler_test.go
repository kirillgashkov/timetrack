package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
)

func TestPostAuth(t *testing.T) {
	tests := []struct {
		name               string
		formData           url.Values
		authorizeFunc      func(ctx context.Context, g *PasswordGrant) (*Token, error)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			"ok",
			url.Values{"grant_type": {string(timetrackapi.Password)}, "username": {"username"}, "password": {"password"}},
			func(context.Context, *PasswordGrant) (*Token, error) {
				return &Token{AccessToken: "valid_token"}, nil
			},
			http.StatusOK,
			`{"access_token":"valid_token","token_type":"Bearer"}`,
		},
		{
			"missing username",
			url.Values{"grant_type": {string(timetrackapi.Password)}, "password": {"password"}},
			nil,
			http.StatusUnprocessableEntity,
			`{"message":"missing username"}`,
		},
		{
			"missing password",
			url.Values{"grant_type": {string(timetrackapi.Password)}, "username": {"username"}},
			nil,
			http.StatusUnprocessableEntity,
			`{"message":"missing password"}`,
		},
		{
			"invalid grant type",
			url.Values{"grant_type": {"invalid"}, "username": {"username"}, "password": {"password"}},
			nil,
			http.StatusUnprocessableEntity,
			`{"message":"invalid grant type"}`,
		},
		{
			"invalid credentials",
			url.Values{"grant_type": {string(timetrackapi.Password)}, "username": {"username"}, "password": {"password"}},
			func(context.Context, *PasswordGrant) (*Token, error) {
				return nil, ErrInvalidCredentials
			},
			http.StatusBadRequest,
			`{"message":"invalid credentials"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &ServiceMock{
				AuthorizeFunc: tt.authorizeFunc,
			}
			handler := NewHandler(mockService)

			formData := tt.formData.Encode()
			req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString(formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			handler.PostAuth(w, req)

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
