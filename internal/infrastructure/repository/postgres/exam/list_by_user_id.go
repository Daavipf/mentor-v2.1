package exam

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *examRepository) ListByUserID(ctx context.Context, userID string) ([]entity.Exam, error) {
	query := `
		SELECT id, user_id, date, title, created_at, updated_at
		FROM exams
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exams []entity.Exam
	for rows.Next() {
		var exam entity.Exam
		err := rows.Scan(
			&exam.ID,
			&exam.UserID,
			&exam.Date,
			&exam.Title,
			&exam.CreatedAt,
			&exam.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		exams = append(exams, exam)
	}

	for i := range exams {
		// Load areas for summary
		areaRows, err := r.db.QueryContext(ctx, "SELECT id, exams_id, area FROM exam_areas WHERE exams_id = $1", exams[i].ID)
		if err == nil {
			for areaRows.Next() {
				var a entity.ExamArea
				if err := areaRows.Scan(&a.ID, &a.ExamsID, &a.Area); err == nil {
					exams[i].Areas = append(exams[i].Areas, a)
				}
			}
			areaRows.Close()
		}

		// Load questions on exam (without full relation nesting for listing speed)
		qeRows, err := r.db.QueryContext(ctx, `
			SELECT question_id, exam_id, got_right, selected_alternative_id, created_at, updated_at
			FROM questions_on_exams
			WHERE exam_id = $1
		`, exams[i].ID)
		if err == nil {
			for qeRows.Next() {
				var qe entity.QuestionsOnExam
				if err := qeRows.Scan(
					&qe.QuestionID,
					&qe.ExamID,
					&qe.GotRight,
					&qe.SelectedAlternativeID,
					&qe.CreatedAt,
					&qe.UpdatedAt,
				); err == nil {
					exams[i].Questions = append(exams[i].Questions, qe)
				}
			}
			qeRows.Close()
		}
	}

	return exams, nil
}
