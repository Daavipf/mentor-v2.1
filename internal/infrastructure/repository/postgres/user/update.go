package user

import (
	"context"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
)

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $1, password = $2, name = $3, updated_at = $4
		WHERE id = $5
	`
	res, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Password,
		user.Name,
		user.UpdatedAt,
		user.ID,
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

