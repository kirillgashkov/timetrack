package task

import (
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
func (h *Handler) GetTasksId(w http.ResponseWriter, r *http.Request, id int) {
	panic("implement me")
}

// PatchTasksId handles "PATCH /tasks/{id}".
func (h *Handler) PatchTasksId(w http.ResponseWriter, r *http.Request, id int) {
	panic("implement me")
}

// DeleteTasksId handles "DELETE /tasks/{id}".
func (h *Handler) DeleteTasksId(w http.ResponseWriter, r *http.Request, id int) {
	panic("implement me")
}

func toTaskAPI(t *Task) *timetrackapi.Task {
	return &timetrackapi.Task{
		Id:          t.ID,
		Description: t.Description,
	}
}
