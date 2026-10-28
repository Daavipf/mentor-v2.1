package history

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *historyRepository) Create(ctx context.Context, h *entity.History) error {
	query := `
		INSERT INTO histories (id, user_id, exam_id, score, created_at, updated_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		h.ID,
		h.UserID,
		h.ExamID,
		h.Score,
		h.CreatedAt,
		h.UpdatedAt,
		h.FinishedAt,
	)
	return err
}

