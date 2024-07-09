package tracking

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	Service *Service
}

// PostTasksIdStart handles "POST /tasks/{id}/start".
//
//nolint:revive
func (h *Handler) PostTasksIdStart(w http.ResponseWriter, r *http.Request, id int) {
	userID := 1 // TODO: use real user ID

	err := h.Service.StartTask(r.Context(), TaskID(id), UserID(userID))
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
	userID := 1 // TODO: use real user ID

	err := h.Service.StopTask(r.Context(), TaskID(id), UserID(userID))
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
