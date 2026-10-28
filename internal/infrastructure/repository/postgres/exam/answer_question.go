package exam

import (
	"context"
	"mentor-api/internal/domain"
	"time"
)

func (r *examRepository) AnswerQuestion(ctx context.Context, examID, questionID, alternativeID string, gotRight bool) error {
	query := `
		UPDATE questions_on_exams
		SET selected_alternative_id = $1, got_right = $2, updated_at = $3
		WHERE exam_id = $4 AND question_id = $5
	`
	res, err := r.db.ExecContext(ctx, query, alternativeID, gotRight, time.Now(), examID, questionID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrQuestionNotInExam
	}

	return nil
}
