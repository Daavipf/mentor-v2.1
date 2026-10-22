package postgres

import (
	"context"
	"database/sql"
	"errors"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.Name,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, email, password, name, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var user entity.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password, name, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	row := r.db.QueryRowContext(ctx, query, email)

	var user entity.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

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

