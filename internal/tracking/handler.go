package tracking

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/kirillgashkov/timetrack/internal/auth"

	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// PostTasksIdStart handles "POST /tasks/{id}/start".
//
//nolint:revive
func (h *Handler) PostTasksIdStart(w http.ResponseWriter, r *http.Request, id int) {
	u := auth.MustUserFromContext(r.Context())

	err := h.service.StartTask(r.Context(), TaskID(id), UserID(u.ID))
	if err != nil {
		if errors.Is(err, ErrAlreadyStarted) {
			apiutil.MustWriteError(w, "task already started", http.StatusBadRequest)
			return
		}
		slog.Error("failed to start task", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteNoContent(w)
}

// PostTasksIdStop handles "POST /tasks/{id}/stop".
//
//nolint:revive
func (h *Handler) PostTasksIdStop(w http.ResponseWriter, r *http.Request, id int) {
	user := auth.MustUserFromContext(r.Context())

	err := h.service.StopTask(r.Context(), TaskID(id), UserID(user.ID))
	if err != nil {
		if errors.Is(err, ErrNotStarted) {
			apiutil.MustWriteError(w, "task not started", http.StatusBadRequest)
			return
		}
		slog.Error("failed to stop task", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteNoContent(w)
}
