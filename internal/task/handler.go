package task

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	Service *Service
}

// PostTasks handles "POST /tasks/".
func (h *Handler) PostTasks(w http.ResponseWriter, r *http.Request) {
	var taskCreate *timetrackapi.TaskCreate
	if err := apiutil.ReadJSON(r, &taskCreate); err != nil {
		apiutil.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	u, err := h.Service.Create(r.Context(), &Create{Description: taskCreate.Description})
	if err != nil {
		slog.Error("failed to create task", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, toTaskAPI(u), http.StatusOK)
}

// GetTasks handles "GET /tasks/".
func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request, params timetrackapi.GetTasksParams) {
	offset := 0
	if params.Offset != nil {
		if *params.Offset < 0 {
			apiutil.MustWriteError(w, "invalid offset", http.StatusUnprocessableEntity)
			return
		}
		offset = *params.Offset
	}
	limit := 50
	if params.Limit != nil {
		if *params.Limit < 1 || *params.Limit > 100 {
			apiutil.MustWriteError(w, "invalid limit", http.StatusUnprocessableEntity)
			return
		}
		limit = *params.Limit
	}

	tasks, err := h.Service.List(r.Context(), offset, limit)
	if err != nil {
		slog.Error("failed to get tasks", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	tasksAPI := make([]*timetrackapi.Task, 0, len(tasks))
	for _, t := range tasks {
		tasksAPI = append(tasksAPI, toTaskAPI(&t))
	}
	apiutil.MustWriteJSON(w, tasksAPI, http.StatusOK)
}

// GetTasksId handles "GET /tasks/{id}".
//
//nolint:revive
func (h *Handler) GetTasksId(w http.ResponseWriter, r *http.Request, id int) {
	t, err := h.Service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "task not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to get task", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, toTaskAPI(t), http.StatusOK)
}

// PatchTasksId handles "PATCH /tasks/{id}".
//
//nolint:revive
func (h *Handler) PatchTasksId(w http.ResponseWriter, r *http.Request, id int) {
	var taskUpdateAPI *timetrackapi.TaskUpdate
	if err := apiutil.ReadJSON(r, &taskUpdateAPI); err != nil {
		apiutil.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	t, err := h.Service.Update(r.Context(), id, &Update{
		Description: taskUpdateAPI.Description,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "task not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to update task", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, toTaskAPI(t), http.StatusOK)
}

// DeleteTasksId handles "DELETE /tasks/{id}".
//
//nolint:revive
func (h *Handler) DeleteTasksId(w http.ResponseWriter, r *http.Request, id int) {
	t, err := h.Service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "task not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to delete task", "error", err)
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, toTaskAPI(t), http.StatusOK)
}

func toTaskAPI(t *Task) *timetrackapi.Task {
	return &timetrackapi.Task{
		Id:          t.ID,
		Description: t.Description,
	}
}
