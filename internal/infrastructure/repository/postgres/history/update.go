package history

import (
	"context"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
)

func (r *historyRepository) Update(ctx context.Context, h *entity.History) error {
	query := `
		UPDATE histories
		SET score = $1, updated_at = $2, finished_at = $3
		WHERE id = $4
	`
	res, err := r.db.ExecContext(ctx, query,
		h.Score,
		h.UpdatedAt,
		h.FinishedAt,
		h.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

