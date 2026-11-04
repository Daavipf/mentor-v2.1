package analytics

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *analyticsRepository) GetPerformanceByArea(ctx context.Context, userID string) ([]entity.AreaPerformance, error) {
	query := `
		SELECT
			q.area,
			COUNT(*) AS total_questions,
			COUNT(*) FILTER (WHERE qoe.got_right = true) AS correct_answers,
			COALESCE(ROUND((COUNT(*) FILTER (WHERE qoe.got_right = true)::numeric / NULLIF(COUNT(*), 0)) * 100, 2), 0) AS accuracy_rate,
			COALESCE(EXTRACT(EPOCH FROM AVG(qoe.finished_at - qoe.started_at)), 0) AS avg_response_time_seconds
		FROM questions_on_exams qoe
		JOIN exams e ON e.id = qoe.exam_id
		JOIN questions q ON q.id = qoe.question_id
		WHERE e.user_id = $1
		  AND (qoe.selected_alternative_id IS NOT NULL OR qoe.got_right IS NOT NULL)
		GROUP BY q.area
		ORDER BY total_questions DESC, q.area ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]entity.AreaPerformance, 0)
	for rows.Next() {
		var ap entity.AreaPerformance
		if err := rows.Scan(
			&ap.Area,
			&ap.TotalQuestions,
			&ap.CorrectAnswers,
			&ap.AccuracyRate,
			&ap.AverageQuestionTime,
		); err != nil {
			return nil, err
		}
		results = append(results, ap)
	}

	return results, nil
}
