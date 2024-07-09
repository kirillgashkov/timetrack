package task

import "net/http"

type Handler struct{}

// PostTasks handles "POST /tasks/".
func (h *Handler) PostTasks(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// GetTasks handles "GET /tasks/".
func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
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
