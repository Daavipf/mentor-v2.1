package question

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *questionRepository) ListRandom(ctx context.Context, limit int) ([]entity.Question, error) {
	query := `
		SELECT id, context, alternatives_introduction, area, discipline, 
		       year, language, index, title, created_at, updated_at
		FROM questions
		ORDER BY RANDOM()
		LIMIT $1
	`
	return r.queryQuestions(ctx, query, limit)
}

