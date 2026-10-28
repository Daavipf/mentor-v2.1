package question

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *questionRepository) ListAll(ctx context.Context, limit, offset int) ([]entity.Question, error) {
	query := `
		SELECT id, context, alternatives_introduction, area, discipline, 
		       year, language, index, title, created_at, updated_at
		FROM questions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	return r.queryQuestions(ctx, query, limit, offset)
}

