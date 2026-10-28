package exam

import (
	"context"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
)

func (r *examRepository) Update(ctx context.Context, exam *entity.Exam) error {
	query := `
		UPDATE exams
		SET title = $1, date = $2, updated_at = $3
		WHERE id = $4
	`
	res, err := r.db.ExecContext(ctx, query,
		exam.Title,
		exam.Date,
		exam.UpdatedAt,
		exam.ID,
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
