package question

import (
	"context"
	"database/sql"
	"errors"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
)

func (r *questionRepository) GetAlternativeByID(ctx context.Context, alternativeID string) (*entity.Alternative, error) {
	query := `
		SELECT id, question_id, text, is_correct, image_path
		FROM alternatives
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, alternativeID)

	var alt entity.Alternative
	err := row.Scan(&alt.ID, &alt.QuestionID, &alt.Text, &alt.IsCorrect, &alt.ImagePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &alt, nil
}

