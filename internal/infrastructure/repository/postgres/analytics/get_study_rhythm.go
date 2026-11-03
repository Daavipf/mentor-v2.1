package analytics

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *analyticsRepository) GetStudyRhythm(ctx context.Context, userID string, days int) ([]entity.DailyStudyRhythm, error) {
	if days <= 0 {
		days = 30
	}

	query := `
		SELECT
			TO_CHAR(DATE_TRUNC('day', COALESCE(qoe.updated_at, e.date)), 'YYYY-MM-DD') AS study_day,
			COUNT(*) AS total_questions,
			COUNT(*) FILTER (WHERE qoe.got_right = true) AS correct_answers
		FROM questions_on_exams qoe
		JOIN exams e ON e.id = qoe.exam_id
		WHERE e.user_id = $1
		  AND (qoe.selected_alternative_id IS NOT NULL OR qoe.got_right IS NOT NULL)
		  AND COALESCE(qoe.updated_at, e.date) >= NOW() - ($2 * INTERVAL '1 day')
		GROUP BY study_day
		ORDER BY study_day ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]entity.DailyStudyRhythm, 0)
	for rows.Next() {
		var sr entity.DailyStudyRhythm
		if err := rows.Scan(
			&sr.Date,
			&sr.TotalQuestions,
			&sr.CorrectAnswers,
		); err != nil {
			return nil, err
		}
		results = append(results, sr)
	}

	return results, nil
}

