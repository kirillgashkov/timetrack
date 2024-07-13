package reporting

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

// PostUsersIdReport handles "POST /users/{id}/report".
//
//nolint:revive
func (h *Handler) PostUsersIdReport(w http.ResponseWriter, r *http.Request, id int) {
	u := auth.MustUserFromContext(r.Context())
	if u.ID != id {
		apiutil.MustWriteForbidden(w)
		return
	}

	req, err := parseAndValidateReportRequest(r)
	if err != nil {
		var ve apiutil.ValidationError
		if errors.As(err, &ve) {
			apiutil.MustWriteUnprocessableEntity(w, ve)
			return
		}
		apiutil.MustWriteInternalServerError(w, "failed to parse and validate request", err)
		return
	}

	reportTasks, err := h.service.Report(r.Context(), id, req.From, req.To)
	if err != nil {
		apiutil.MustWriteInternalServerError(w, "failed to generate report", err)
		return
	}

	resp := make([]*timetrackapi.ReportTaskResponse, 0, len(reportTasks))
	for _, t := range reportTasks {
		reportTaskResp := &timetrackapi.ReportTaskResponse{
			Task: &timetrackapi.TaskResponse{
				Id:          t.Task.ID,
				Description: t.Task.Description,
			},
			Duration: &timetrackapi.ReportDurationResponse{
				Hours:   int(t.Duration.Hours()),
				Minutes: int(t.Duration.Minutes()) % 60,
			},
		}
		resp = append(resp, reportTaskResp)
	}
	apiutil.MustWriteJSON(w, resp, http.StatusOK)
}

func parseAndValidateReportRequest(r *http.Request) (*timetrackapi.ReportRequest, error) {
	var req *timetrackapi.ReportRequest
	if err := apiutil.ReadJSON(r, &req); err != nil {
		return nil, errors.Join(apiutil.ValidationError{"bad JSON"}, err)
	}
	if err := validateReportRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

func validateReportRequest(req *timetrackapi.ReportRequest) error {
	e := make([]string, 0)

	if req.From.IsZero() {
		e = append(e, "missing from")
	}
	if req.To.IsZero() {
		e = append(e, "missing to")
	}
	if req.From.After(req.To) {
		e = append(e, "from must be before to")
	}

	if len(e) > 0 {
		return apiutil.ValidationError(e)
	}
	return nil
}
