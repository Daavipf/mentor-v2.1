package exam

import (
	"context"
	"database/sql"
	"errors"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
)

func (r *examRepository) GetByID(ctx context.Context, id string) (*entity.Exam, error) {
	query := `
		SELECT id, user_id, date, title, created_at, updated_at
		FROM exams
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var exam entity.Exam
	err := row.Scan(
		&exam.ID,
		&exam.UserID,
		&exam.Date,
		&exam.Title,
		&exam.CreatedAt,
		&exam.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	// Areas
	areaRows, err := r.db.QueryContext(ctx, "SELECT id, exams_id, area FROM exam_areas WHERE exams_id = $1", exam.ID)
	if err != nil {
		return nil, err
	}
	defer areaRows.Close()

	for areaRows.Next() {
		var a entity.ExamArea
		if err := areaRows.Scan(&a.ID, &a.ExamsID, &a.Area); err != nil {
			return nil, err
		}
		exam.Areas = append(exam.Areas, a)
	}

	// Questions on exam
	qeRows, err := r.db.QueryContext(ctx, `
		SELECT question_id, exam_id, got_right, selected_alternative_id, created_at, updated_at
		FROM questions_on_exams
		WHERE exam_id = $1
		ORDER BY created_at ASC
	`, exam.ID)
	if err != nil {
		return nil, err
	}
	defer qeRows.Close()

	for qeRows.Next() {
		var qe entity.QuestionsOnExam
		if err := qeRows.Scan(
			&qe.QuestionID,
			&qe.ExamID,
			&qe.GotRight,
			&qe.SelectedAlternativeID,
			&qe.CreatedAt,
			&qe.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Load related question details
		if q, err := r.questionRepo.GetByID(ctx, qe.QuestionID); err == nil && q != nil {
			qe.Question = q
		}

		exam.Questions = append(exam.Questions, qe)
	}

	return &exam, nil
}
