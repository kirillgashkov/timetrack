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

	authMiddleware := auth.NewMiddleware(authService)
	mux := newServeMux(si, authMiddleware)

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
func newServeMux(si timetrackapi.ServerInterface, authMiddleware *auth.Middleware) *http.ServeMux {
	wrapper := timetrackapi.ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: make([]timetrackapi.MiddlewareFunc, 0),
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			slog.Error("oapi-codegen error", "error", err)
			apiutil.MustWriteInternalServerError(w)
		},
	}

	authenticated := func(f func(http.ResponseWriter, *http.Request)) http.Handler {
		return authMiddleware.Authenticated(http.HandlerFunc(f))
	}

	m := http.NewServeMux()
	m.HandleFunc("POST /auth", wrapper.PostAuth)
	m.HandleFunc("GET /health", wrapper.GetHealth)
	m.Handle("GET /tasks/", authenticated(wrapper.GetTasks))
	m.Handle("POST /tasks/", authenticated(wrapper.PostTasks))
	m.Handle("DELETE /tasks/{id}", authenticated(wrapper.DeleteTasksId))
	m.Handle("GET /tasks/{id}", authenticated(wrapper.GetTasksId))
	m.Handle("PATCH /tasks/{id}", authenticated(wrapper.PatchTasksId))
	m.Handle("POST /tasks/{id}/start", authenticated(wrapper.PostTasksIdStart))
	m.Handle("POST /tasks/{id}/stop", authenticated(wrapper.PostTasksIdStop))
	m.Handle("GET /users/", authenticated(wrapper.GetUsers))
	m.HandleFunc("POST /users/", wrapper.PostUsers)
	m.Handle("GET /users/current", authenticated(wrapper.GetUsersCurrent))
	m.Handle("POST /users/{id}/report", authenticated(wrapper.PostUsersIdReport))
	m.Handle("DELETE /users/{id}", authenticated(wrapper.DeleteUsersId))
	m.Handle("GET /users/{id}", authenticated(wrapper.GetUsersId))
	m.Handle("PATCH /users/{id}", authenticated(wrapper.PatchUsersId))

	return m
}
