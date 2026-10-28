package question

import (
	"context"
	"database/sql"
	"errors"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
)

func (r *questionRepository) GetByID(ctx context.Context, id string) (*entity.Question, error) {
	query := `
		SELECT id, context, alternatives_introduction, area, discipline, 
		       year, language, index, title, created_at, updated_at
		FROM questions
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var q entity.Question
	err := row.Scan(
		&q.ID,
		&q.Context,
		&q.AlternativesIntroduction,
		&q.Area,
		&q.Discipline,
		&q.Year,
		&q.Language,
		&q.Index,
		&q.Title,
		&q.CreatedAt,
		&q.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if err := r.loadQuestionRelations(ctx, &q); err != nil {
		return nil, err
	}

	return &q, nil
}

