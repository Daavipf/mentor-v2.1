package analytics

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *analyticsRepository) GetExamsHistory(ctx context.Context, userID string) ([]entity.ExamPerformanceHistory, error) {
	query := `
		SELECT
			e.id AS exam_id,
			e.title AS exam_title,
			e.date AS exam_date,
			h.finished_at AS finished_at,
			COALESCE(h.score, 0) AS score,
			COUNT(qoe.question_id) AS total_questions,
			COUNT(qoe.question_id) FILTER (WHERE qoe.got_right = true) AS correct_answers
		FROM histories h
		JOIN exams e ON e.id = h.exam_id
		LEFT JOIN questions_on_exams qoe ON qoe.exam_id = e.id
		WHERE h.user_id = $1
		  AND h.finished_at IS NOT NULL
		GROUP BY e.id, e.title, e.date, h.finished_at, h.score
		ORDER BY h.finished_at ASC, e.date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]entity.ExamPerformanceHistory, 0)
	for rows.Next() {
		var eh entity.ExamPerformanceHistory
		if err := rows.Scan(
			&eh.ExamID,
			&eh.ExamTitle,
			&eh.Date,
			&eh.FinishedAt,
			&eh.Score,
			&eh.TotalQuestions,
			&eh.CorrectAnswers,
		); err != nil {
			return nil, err
		}
		results = append(results, eh)
	}

	return results, nil
}

