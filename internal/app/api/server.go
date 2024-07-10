package api

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/config"
	"github.com/kirillgashkov/timetrack/internal/auth"
	"github.com/kirillgashkov/timetrack/internal/reporting"
	"github.com/kirillgashkov/timetrack/internal/task"
	"github.com/kirillgashkov/timetrack/internal/tracking"
	"github.com/kirillgashkov/timetrack/internal/user"
)

func NewServer(
	cfg *config.ServerConfig,
	authService *auth.Service,
	reportingService *reporting.Service,
	taskService *task.Service,
	trackingService *tracking.Service,
	userService *user.Service,
) (*http.Server, error) {
	si := NewHandler(authService, reportingService, taskService, trackingService, userService)
	mux := newServeMux(si)

	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}
	return srv, nil
}

// newServeMux creates a new http.ServeMux and registers all handlers of the
// API.
//
// oapi-codegen provides their own timetrackapi.HandlerWithOptions function that
// can be used to create a new handler. However, it does not allow to pass
// per-handler middlewares. This is why we create a new ServeMux and register
// all handlers manually. Probably a better solution would be to switch to other
// oapi-codegen backend that supports per-handler middlewares or other OpenAPI
// library, but it would be too much hassle for now.
func newServeMux(si timetrackapi.ServerInterface) *http.ServeMux {
	wrapper := timetrackapi.ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: make([]timetrackapi.MiddlewareFunc, 0),
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			slog.Error("oapi-codegen error", "error", err)
			apiutil.MustWriteInternalServerError(w)
		},
	}

	m := http.NewServeMux()
	m.HandleFunc("POST /auth", wrapper.PostAuth)
	m.HandleFunc("GET /health", wrapper.GetHealth)
	m.HandleFunc("GET /tasks/", wrapper.GetTasks)
	m.HandleFunc("POST /tasks/", wrapper.PostTasks)
	m.HandleFunc("DELETE /tasks/{id}", wrapper.DeleteTasksId)
	m.HandleFunc("GET /tasks/{id}", wrapper.GetTasksId)
	m.HandleFunc("PATCH /tasks/{id}", wrapper.PatchTasksId)
	m.HandleFunc("POST /tasks/{id}/start", wrapper.PostTasksIdStart)
	m.HandleFunc("POST /tasks/{id}/stop", wrapper.PostTasksIdStop)
	m.HandleFunc("GET /users/", wrapper.GetUsers)
	m.HandleFunc("POST /users/", wrapper.PostUsers)
	m.HandleFunc("GET /users/current", wrapper.GetUsersCurrent)
	m.HandleFunc("POST /users/{id}/report", wrapper.PostUsersIdReport)
	m.HandleFunc("DELETE /users/{passportNumber}", wrapper.DeleteUsersPassportNumber)
	m.HandleFunc("GET /users/{passportNumber}", wrapper.GetUsersPassportNumber)
	m.HandleFunc("PATCH /users/{passportNumber}", wrapper.PatchUsersPassportNumber)

	return m
}
