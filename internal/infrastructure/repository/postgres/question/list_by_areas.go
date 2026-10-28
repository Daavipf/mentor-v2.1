package question

import (
	"context"
	"fmt"
	"strings"

	"mentor-api/internal/domain/entity"
)

func (r *questionRepository) ListByAreas(ctx context.Context, areas []string, limit int) ([]entity.Question, error) {
	if len(areas) == 0 {
		return r.ListRandom(ctx, limit)
	}

	placeholders := make([]string, len(areas))
	args := make([]interface{}, len(areas)+1)
	for i, a := range areas {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = a
	}
	args[len(areas)] = limit

	query := fmt.Sprintf(`
		SELECT id, context, alternatives_introduction, area, discipline, 
		       year, language, index, title, created_at, updated_at
		FROM questions
		WHERE area IN (%s)
		ORDER BY RANDOM()
		LIMIT $%d
	`, strings.Join(placeholders, ", "), len(areas)+1)

	return r.queryQuestions(ctx, query, args...)
}

