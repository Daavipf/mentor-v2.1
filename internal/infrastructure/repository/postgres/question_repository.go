package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

type questionRepository struct {
	db *sql.DB
}

func NewQuestionRepository(db *sql.DB) repository.QuestionRepository {
	return &questionRepository{db: db}
}

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

