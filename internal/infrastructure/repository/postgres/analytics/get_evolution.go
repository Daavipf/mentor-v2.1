package analytics

import (
	"context"
	"strings"

	"mentor-api/internal/domain/entity"
)

func (r *analyticsRepository) GetEvolution(
	ctx context.Context,
	userID string,
	periodType string,
	months int,
) ([]entity.TimePeriodMetric, error) {
	if months <= 0 {
		months = 6
	}

	periodType = strings.ToLower(strings.TrimSpace(periodType))
	if periodType != "monthly" {
		periodType = "weekly"
	}

	query := `
		SELECT
			TO_CHAR(
				DATE_TRUNC(CASE WHEN $2 = 'monthly' THEN 'month' ELSE 'week' END, COALESCE(qoe.updated_at, e.date)),
				CASE WHEN $2 = 'monthly' THEN 'YYYY-MM' ELSE 'YYYY-"W"IW' END
			) AS period,
			DATE_TRUNC(CASE WHEN $2 = 'monthly' THEN 'month' ELSE 'week' END, COALESCE(qoe.updated_at, e.date)) AS period_date,
			COUNT(*) AS total_questions,
			COUNT(*) FILTER (WHERE qoe.got_right = true) AS correct_answers,
			COALESCE(ROUND((COUNT(*) FILTER (WHERE qoe.got_right = true)::numeric / NULLIF(COUNT(*), 0)) * 100, 2), 0) AS accuracy_rate
		FROM questions_on_exams qoe
		JOIN exams e ON e.id = qoe.exam_id
		WHERE e.user_id = $1
		  AND (qoe.selected_alternative_id IS NOT NULL OR qoe.got_right IS NOT NULL)
		  AND COALESCE(qoe.updated_at, e.date) >= NOW() - ($3 * INTERVAL '1 month')
		GROUP BY DATE_TRUNC(CASE WHEN $2 = 'monthly' THEN 'month' ELSE 'week' END, COALESCE(qoe.updated_at, e.date)), period
		ORDER BY period_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, periodType, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]entity.TimePeriodMetric, 0)
	for rows.Next() {
		var m entity.TimePeriodMetric
		if err := rows.Scan(
			&m.Period,
			&m.PeriodDate,
			&m.TotalQuestions,
			&m.CorrectAnswers,
			&m.AccuracyRate,
		); err != nil {
			return nil, err
		}
		results = append(results, m)
	}

	return results, nil
}

