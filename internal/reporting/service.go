package reporting

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kirillgashkov/timetrack/internal/task"
)

type Service struct {
	db *pgxpool.Pool
}

type ReportTask struct {
	Task     task.Task
	Duration time.Duration
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Report(ctx context.Context, userID int, from, to time.Time) ([]ReportTask, error) {
	rows, err := s.db.Query(
		ctx,
		`
			SELECT tasks.id AS task_id,
				   tasks.description AS task_description,
				   SUM(LEAST(works.stopped_at, $3) - GREATEST(works.started_at, $2)) AS duration
			FROM works
			JOIN tasks ON works.task_id = tasks.id
			WHERE user_id = $1
			  AND (
				  (works.started_at >= $2 AND works.started_at <= $3)
				  OR (works.stopped_at >= $2 AND works.stopped_at <= $3)
			  )
			GROUP BY tasks.id;
		`,
		userID,
		from,
		to,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select report"), err)
	}
	defer rows.Close()

	type reportTask struct {
		TaskID          int           `db:"task_id"`
		TaskDescription string        `db:"task_description"`
		Duration        time.Duration `db:"duration"`
	}
	reportTasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[reportTask])
	if err != nil {
		return nil, errors.Join(errors.New("failed to collect report tasks"), err)
	}

	report := make([]ReportTask, 0, len(reportTasks))
	for _, rt := range reportTasks {
		report = append(report, ReportTask{
			Task: task.Task{
				ID:          rt.TaskID,
				Description: rt.TaskDescription,
			},
			Duration: rt.Duration,
		})
	}
	return report, nil
}
