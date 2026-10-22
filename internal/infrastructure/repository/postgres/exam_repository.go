package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

type examRepository struct {
	db           *sql.DB
	questionRepo repository.QuestionRepository
}

func NewExamRepository(db *sql.DB, questionRepo repository.QuestionRepository) repository.ExamRepository {
	return &examRepository{
		db:           db,
		questionRepo: questionRepo,
	}
}

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

func (r *examRepository) Update(ctx context.Context, exam *entity.Exam) error {
	query := `
		UPDATE exams
		SET title = $1, date = $2, updated_at = $3
		WHERE id = $4
	`
	res, err := r.db.ExecContext(ctx, query,
		exam.Title,
		exam.Date,
		exam.UpdatedAt,
		exam.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *examRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM exams WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

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

