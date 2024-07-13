package task

import (
	"errors"
	"net/http"

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
//
// For simplicity, we don't use user in the task domain, but we do use them in
// other domains.
func (h *Handler) PostTasks(w http.ResponseWriter, r *http.Request) {
	var req *timetrackapi.CreateTaskRequest
	if err := apiutil.ReadJSON(r, &req); err != nil {
		apiutil.MustWriteUnprocessableEntity(w, apiutil.ValidationError{"bad JSON"})
		return
	}

	create := &CreateTask{Description: req.Description}
	t, err := h.service.Create(r.Context(), create)
	if err != nil {
		apiutil.MustWriteInternalServerError(w, "failed to create task", err)
		return
	}

	resp := toTaskResponse(t)
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

// GetTasks handles "GET /tasks/".
//
// For simplicity, we don't use user in the task domain, but we do use them in
// other domains.
func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request, params timetrackapi.GetTasksParams) {
	if err := validateAndNormalizeListTasksRequest(&params); err != nil {
		var ve apiutil.ValidationError
		if errors.As(err, &ve) {
			apiutil.MustWriteUnprocessableEntity(w, ve)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to validate and normalize request", err)
		return
	}

	tasks, err := h.service.List(r.Context(), *params.Offset, *params.Limit)
	if err != nil {
		apiutil.MustWriteInternalServerError(w, "failed to list tasks", err)
		return
	}

	resp := make([]*timetrackapi.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, toTaskResponse(&t))
	}
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

func validateAndNormalizeListTasksRequest(params *timetrackapi.GetTasksParams) error {
	if err := validateListTasksRequest(*params); err != nil {
		return err
	}
	normalizeListTasksRequest(params)
	return nil
}

func validateListTasksRequest(params timetrackapi.GetTasksParams) error {
	e := make([]string, 0)

	if params.Offset != nil && *params.Offset < 0 {
		e = append(e, "invalid offset, must be greater than or equal to 0")
	}
	if params.Limit != nil && *params.Limit < 1 || *params.Limit > 100 {
		e = append(e, "invalid limit, must be between 1 and 100")
	}

	if len(e) > 0 {
		return apiutil.ValidationError(e)
	}
	return nil
}

func normalizeListTasksRequest(params *timetrackapi.GetTasksParams) {
	if params.Offset == nil {
		params.Offset = intPtr(0)
	}
	if params.Limit == nil {
		params.Limit = intPtr(50)
	}
}

// GetTasksId handles "GET /tasks/{id}".
//
// For simplicity, we don't use user in the task domain, but we do use them in
// other domains.
//
//nolint:revive
func (h *Handler) GetTasksId(w http.ResponseWriter, r *http.Request, id int) {
	t, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "task not found", http.StatusNotFound)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to get task", err)
		return
	}

	resp := toTaskResponse(t)
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

// PatchTasksId handles "PATCH /tasks/{id}".
//
// For simplicity, we don't use user in the task domain, but we do use them in
// other domains.
//
//nolint:revive
func (h *Handler) PatchTasksId(w http.ResponseWriter, r *http.Request, id int) {
	var req *timetrackapi.UpdateTaskRequest
	if err := apiutil.ReadJSON(r, &req); err != nil {
		apiutil.MustWriteUnprocessableEntity(w, apiutil.ValidationError{"bad JSON"})
		return
	}

	update := &UpdateTask{Description: req.Description}
	t, err := h.service.Update(r.Context(), id, update)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "task not found", http.StatusNotFound)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to update task", err)
		return
	}

	apiutil.MustWriteJSON(w, toTaskResponse(t), http.StatusOK)
}

// DeleteTasksId handles "DELETE /tasks/{id}".
//
// For simplicity, we don't use user in the task domain, but we do use them in
// other domains.
//
//nolint:revive
func (h *Handler) DeleteTasksId(w http.ResponseWriter, r *http.Request, id int) {
	t, err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiutil.MustWriteError(w, "task not found", http.StatusNotFound)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to delete task", err)
		return
	}

	resp := toTaskResponse(t)
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

func toTaskResponse(t *Task) *timetrackapi.TaskResponse {
	return &timetrackapi.TaskResponse{
		Id:          t.ID,
		Description: t.Description,
	}
}

func intPtr(i int) *int {
	return &i
}
