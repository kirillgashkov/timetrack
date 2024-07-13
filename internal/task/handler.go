package task

import (
	"errors"
	"net/http"

	"github.com/kirillgashkov/timetrack/internal/auth"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// PostTasks handles "POST /tasks/".
func (h *Handler) PostTasks(w http.ResponseWriter, r *http.Request) {
	// For simplicity, we don't use user in the task domain, but we do use them
	// in other domains.
	_ = auth.MustUserFromContext(r.Context())

	var taskCreate *timetrackapi.TaskCreate
	if err := apiutil.ReadJSON(r, &taskCreate); err != nil {
		apiutil.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	u, err := h.service.Create(r.Context(), &CreateTask{Description: taskCreate.Description})
	if err != nil {
		apiutil.MustWriteInternalServerError(w, "failed to create task", err)
		return
	}

	apiutil.MustWriteJSON(w, toTaskAPI(u), http.StatusOK)
}

// GetTasks handles "GET /tasks/".
func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request, params timetrackapi.GetTasksParams) {
	// For simplicity, we don't use user in the task domain, but we do use them
	// in other domains.
	_ = auth.MustUserFromContext(r.Context())

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

	tasks, err := h.service.List(r.Context(), offset, limit)
	if err != nil {
		apiutil.MustWriteInternalServerError(w, "failed to get tasks", err)
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
	// For simplicity, we don't use user in the task domain, but we do use them
	// in other domains.
	_ = auth.MustUserFromContext(r.Context())

	t, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "task not found", http.StatusNotFound)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to get task", err)
		return
	}

	apiutil.MustWriteJSON(w, toTaskAPI(t), http.StatusOK)
}

// PatchTasksId handles "PATCH /tasks/{id}".
//
//nolint:revive
func (h *Handler) PatchTasksId(w http.ResponseWriter, r *http.Request, id int) {
	// For simplicity, we don't use user in the task domain, but we do use them
	// in other domains.
	_ = auth.MustUserFromContext(r.Context())

	var taskUpdateAPI *timetrackapi.TaskUpdate
	if err := apiutil.ReadJSON(r, &taskUpdateAPI); err != nil {
		apiutil.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}

	t, err := h.service.Update(r.Context(), id, &UpdateTask{
		Description: taskUpdateAPI.Description,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "task not found", http.StatusNotFound)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to update task", err)
		return
	}

	apiutil.MustWriteJSON(w, toTaskAPI(t), http.StatusOK)
}

// DeleteTasksId handles "DELETE /tasks/{id}".
//
//nolint:revive
func (h *Handler) DeleteTasksId(w http.ResponseWriter, r *http.Request, id int) {
	// For simplicity, we don't use user in the task domain, but we do use them
	// in other domains.
	_ = auth.MustUserFromContext(r.Context())

	t, err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "task not found", http.StatusNotFound)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to delete task", err)
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
