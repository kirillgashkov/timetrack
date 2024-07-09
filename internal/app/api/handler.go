package api

import (
	"net/http"

	"github.com/kirillgashkov/timetrack/internal/reporting"

	"github.com/kirillgashkov/timetrack/internal/tracking"

	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/task"
	"github.com/kirillgashkov/timetrack/internal/user"
)

type reportingHandler = reporting.Handler
type taskHandler = task.Handler
type trackingHandler = tracking.Handler
type userHandler = user.Handler

type Handler struct {
	*reportingHandler
	*taskHandler
	*trackingHandler
	*userHandler
}

func NewHandler(
	reportingService *reporting.Service,
	taskService *task.Service,
	trackingService *tracking.Service,
	userService *user.Service,
) *Handler {
	return &Handler{
		reportingHandler: reporting.NewHandler(reportingService),
		taskHandler:      task.NewHandler(taskService),
		trackingHandler:  tracking.NewHandler(trackingService),
		userHandler:      user.NewHandler(userService),
	}
}

func (h *Handler) GetHealth(w http.ResponseWriter, _ *http.Request) {
	apiutil.MustWriteJSON(w, timetrackapi.Health{Status: "ok"}, http.StatusOK)
}
