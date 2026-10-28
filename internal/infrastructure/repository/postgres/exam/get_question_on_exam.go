package exam

import (
	"context"
	"database/sql"
	"errors"
	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
)

func (r *examRepository) GetQuestionOnExam(ctx context.Context, examID, questionID string) (*entity.QuestionsOnExam, error) {
	query := `
		SELECT question_id, exam_id, got_right, selected_alternative_id, created_at, updated_at
		FROM questions_on_exams
		WHERE exam_id = $1 AND question_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, examID, questionID)

	var qe entity.QuestionsOnExam
	err := row.Scan(
		&qe.QuestionID,
		&qe.ExamID,
		&qe.GotRight,
		&qe.SelectedAlternativeID,
		&qe.CreatedAt,
		&qe.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrQuestionNotInExam
		}
		return nil, err
	}

	return &qe, nil
}
