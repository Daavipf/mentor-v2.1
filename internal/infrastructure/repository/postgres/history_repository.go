package postgres

import (
	"context"
	"database/sql"
	"errors"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

type historyRepository struct {
	db *sql.DB
}

func NewHistoryRepository(db *sql.DB) repository.HistoryRepository {
	return &historyRepository{db: db}
}

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

func (r *historyRepository) GetByID(ctx context.Context, id string) (*entity.History, error) {
	query := `
		SELECT id, user_id, exam_id, score, created_at, updated_at, finished_at
		FROM histories
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

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

