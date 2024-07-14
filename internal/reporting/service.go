package reporting

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kirillgashkov/timetrack/internal/task"
)

type ReportTask struct {
	Task     task.Task
	Duration time.Duration
}

type reportTaskRow struct {
	TaskID          int           `db:"task_id"`
	TaskDescription string        `db:"task_description"`
	Duration        time.Duration `db:"duration"`
}

type Service interface {
	Report(ctx context.Context, userID int, from, to time.Time) ([]ReportTask, error)
}

type ServiceImpl struct {
	db *pgxpool.Pool
}

func NewServiceImpl(db *pgxpool.Pool) *ServiceImpl {
	return &ServiceImpl{db: db}
}

func (s *ServiceImpl) Report(ctx context.Context, userID int, from, to time.Time) ([]ReportTask, error) {
	reportTaskRows, err := s.queryReportTasks(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	reportTasks := make([]ReportTask, 0, len(reportTaskRows))
	for _, rtr := range reportTaskRows {
		reportTasks = append(reportTasks, ReportTask{
			Task: task.Task{
				ID:          rtr.TaskID,
				Description: rtr.TaskDescription,
			},
			Duration: rtr.Duration,
		})
	}
	return reportTasks, nil
}

func (s *ServiceImpl) queryReportTasks(ctx context.Context, userID int, from, to time.Time) ([]reportTaskRow, error) {
	q := `
		SELECT tasks.id AS task_id,
			   tasks.description AS task_description,
			   SUM(LEAST(COALESCE(works.stopped_at, $3), $3) - GREATEST(works.started_at, $2)) AS duration
		FROM works
		JOIN tasks ON works.task_id = tasks.id
		WHERE user_id = $1
		  AND (
			  (works.started_at >= $2 AND works.started_at <= $3)
			  OR (works.stopped_at >= $2 AND works.stopped_at <= $3)
		  )
		GROUP BY tasks.id
		ORDER BY duration DESC, task_id
	`
	rows, err := s.db.Query(ctx, q, userID, from, to)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select report"), err)
	}
	defer rows.Close()

	reportTasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[reportTaskRow])
	if err != nil {
		return nil, errors.Join(errors.New("failed to collect report tasks"), err)
	}

	return reportTasks, nil
}
