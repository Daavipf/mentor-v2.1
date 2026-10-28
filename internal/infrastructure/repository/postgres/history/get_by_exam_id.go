package history

import (
	"context"
	"database/sql"
	"errors"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
)

func (r *historyRepository) GetByExamID(ctx context.Context, examID string) (*entity.History, error) {
	query := `
		SELECT id, user_id, exam_id, score, created_at, updated_at, finished_at
		FROM histories
		WHERE exam_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, examID)

	var h entity.History
	err := row.Scan(
		&h.ID,
		&h.UserID,
		&h.ExamID,
		&h.Score,
		&h.CreatedAt,
		&h.UpdatedAt,
		&h.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &h, nil
}

