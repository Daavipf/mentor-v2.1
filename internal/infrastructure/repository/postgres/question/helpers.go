package question

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *questionRepository) queryQuestions(ctx context.Context, query string, args ...interface{}) ([]entity.Question, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []entity.Question
	for rows.Next() {
		var q entity.Question
		err := rows.Scan(
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
			return nil, err
		}
		questions = append(questions, q)
	}

	for i := range questions {
		if err := r.loadQuestionRelations(ctx, &questions[i]); err != nil {
			return nil, err
		}
	}

	return questions, nil
}

func (r *questionRepository) loadQuestionRelations(ctx context.Context, q *entity.Question) error {
	// Topics
	tRows, err := r.db.QueryContext(ctx, "SELECT id, question_id, topic FROM question_topics WHERE question_id = $1", q.ID)
	if err != nil {
		return err
	}
	defer tRows.Close()

	for tRows.Next() {
		var t entity.QuestionTopic
		if err := tRows.Scan(&t.ID, &t.QuestionID, &t.Topic); err != nil {
			return err
		}
		q.Topics = append(q.Topics, t)
	}

	// Images
	iRows, err := r.db.QueryContext(ctx, "SELECT id, question_id, path FROM question_images WHERE question_id = $1", q.ID)
	if err != nil {
		return err
	}
	defer iRows.Close()

	for iRows.Next() {
		var img entity.QuestionImage
		if err := iRows.Scan(&img.ID, &img.QuestionID, &img.Path); err != nil {
			return err
		}
		q.Images = append(q.Images, img)
	}

	// Alternatives
	aRows, err := r.db.QueryContext(ctx, "SELECT id, question_id, text, is_correct, image_path FROM alternatives WHERE question_id = $1", q.ID)
	if err != nil {
		return err
	}
	defer aRows.Close()

	for aRows.Next() {
		var alt entity.Alternative
		if err := aRows.Scan(&alt.ID, &alt.QuestionID, &alt.Text, &alt.IsCorrect, &alt.ImagePath); err != nil {
			return err
		}
		q.Alternatives = append(q.Alternatives, alt)
	}

	return nil
}

