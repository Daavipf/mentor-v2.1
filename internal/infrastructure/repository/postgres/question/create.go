package question

import (
	"context"
	"fmt"

	"mentor-api/internal/domain/entity"
)

func (r *questionRepository) Create(ctx context.Context, q *entity.Question) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryQuestion := `
		INSERT INTO questions (
			id, context, alternatives_introduction, area, discipline, 
			year, language, index, title, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.ExecContext(ctx, queryQuestion,
		q.ID,
		q.Context,
		q.AlternativesIntroduction,
		q.Area,
		q.Discipline,
		q.Year,
		q.Language,
		q.Index,
		q.Title,
		q.CreatedAt,
		q.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert question: %w", err)
	}

	for _, topic := range q.Topics {
		queryTopic := `
			INSERT INTO question_topics (id, question_id, topic)
			VALUES ($1, $2, $3)
		`
		if _, err := tx.ExecContext(ctx, queryTopic, topic.ID, q.ID, topic.Topic); err != nil {
			return fmt.Errorf("failed to insert question topic: %w", err)
		}
	}

	for _, img := range q.Images {
		queryImage := `
			INSERT INTO question_images (id, question_id, path)
			VALUES ($1, $2, $3)
		`
		if _, err := tx.ExecContext(ctx, queryImage, img.ID, q.ID, img.Path); err != nil {
			return fmt.Errorf("failed to insert question image: %w", err)
		}
	}

	for _, alt := range q.Alternatives {
		queryAlt := `
			INSERT INTO alternatives (id, question_id, text, is_correct, image_path)
			VALUES ($1, $2, $3, $4, $5)
		`
		if _, err := tx.ExecContext(ctx, queryAlt, alt.ID, q.ID, alt.Text, alt.IsCorrect, alt.ImagePath); err != nil {
			return fmt.Errorf("failed to insert alternative: %w", err)
		}
	}

	return tx.Commit()
}

