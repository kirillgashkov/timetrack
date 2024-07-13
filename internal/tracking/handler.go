package tracking

import (
	"errors"
	"net/http"

	"github.com/kirillgashkov/timetrack/internal/auth"

	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// PostTasksIdStart handles "POST /tasks/{id}/start".
//
//nolint:revive
func (h *Handler) PostTasksIdStart(w http.ResponseWriter, r *http.Request, id int) {
	currentUser := auth.MustUserFromContext(r.Context())

	err := h.service.StartTask(r.Context(), TaskID(id), UserID(currentUser.ID))
	if err != nil {
		if errors.Is(err, ErrAlreadyStarted) {
			apiutil.MustWriteError(w, "task already started", http.StatusBadRequest)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to start task", err)
		return
	}

	apiutil.MustWriteNoContent(w)
}

// PostTasksIdStop handles "POST /tasks/{id}/stop".
//
//nolint:revive
func (h *Handler) PostTasksIdStop(w http.ResponseWriter, r *http.Request, id int) {
	currentUser := auth.MustUserFromContext(r.Context())

	err := h.service.StopTask(r.Context(), TaskID(id), UserID(currentUser.ID))
	if err != nil {
		if errors.Is(err, ErrNotStarted) {
			apiutil.MustWriteError(w, "task not started", http.StatusBadRequest)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to stop task", err)
		return
	}

	apiutil.MustWriteNoContent(w)
}
