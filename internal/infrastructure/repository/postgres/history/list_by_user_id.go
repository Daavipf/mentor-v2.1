package history

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *historyRepository) ListByUserID(ctx context.Context, userID string) ([]entity.History, error) {
	query := `
		SELECT id, user_id, exam_id, score, created_at, updated_at, finished_at
		FROM histories
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []entity.History
	for rows.Next() {
		var h entity.History
		err := rows.Scan(
			&h.ID,
			&h.UserID,
			&h.ExamID,
			&h.Score,
			&h.CreatedAt,
			&h.UpdatedAt,
			&h.FinishedAt,
		)
		if err != nil {
			return nil, err
		}
		histories = append(histories, h)
	}

	return histories, nil
}

