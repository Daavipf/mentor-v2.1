package exam

import (
	"context"
	"fmt"
	"mentor-api/internal/domain/entity"
)

func (r *examRepository) Create(ctx context.Context, exam *entity.Exam) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryExam := `
		INSERT INTO exams (id, user_id, date, title, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, queryExam,
		exam.ID,
		exam.UserID,
		exam.Date,
		exam.Title,
		exam.CreatedAt,
		exam.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert exam: %w", err)
	}

	for _, area := range exam.Areas {
		queryArea := `
			INSERT INTO exam_areas (id, exams_id, area)
			VALUES ($1, $2, $3)
		`
		if _, err := tx.ExecContext(ctx, queryArea, area.ID, exam.ID, area.Area); err != nil {
			return fmt.Errorf("failed to insert exam area: %w", err)
		}
	}

	for _, qe := range exam.Questions {
		queryQE := `
			INSERT INTO questions_on_exams (
				question_id, exam_id, got_right, selected_alternative_id, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := tx.ExecContext(ctx, queryQE,
			qe.QuestionID,
			exam.ID,
			qe.GotRight,
			qe.SelectedAlternativeID,
			qe.CreatedAt,
			qe.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert question on exam: %w", err)
		}
	}

	return tx.Commit()
}
