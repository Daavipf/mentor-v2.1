package analytics

import (
	"context"
	"fmt"

	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

func (r *analyticsRepository) GetPerformanceByDiscipline(
	ctx context.Context,
	userID string,
	filter repository.DisciplineFilter,
) ([]entity.DisciplinePerformance, error) {
	query := `
		SELECT
			COALESCE(q.discipline, 'Geral') AS discipline,
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
	`

	args := []interface{}{userID}

	if filter.StartDate != nil {
		args = append(args, *filter.StartDate)
		query += fmt.Sprintf(" AND e.date >= $%d", len(args))
	}

	if filter.EndDate != nil {
		args = append(args, *filter.EndDate)
		query += fmt.Sprintf(" AND e.date <= $%d", len(args))
	}

	if filter.Area != "" {
		args = append(args, filter.Area)
		query += fmt.Sprintf(" AND q.area = $%d", len(args))
	}

	query += `
		GROUP BY COALESCE(q.discipline, 'Geral'), q.area
		ORDER BY total_questions DESC, accuracy_rate DESC, discipline ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]entity.DisciplinePerformance, 0)
	for rows.Next() {
		var dp entity.DisciplinePerformance
		if err := rows.Scan(
			&dp.Discipline,
			&dp.Area,
			&dp.TotalQuestions,
			&dp.CorrectAnswers,
			&dp.AccuracyRate,
			&dp.AverageQuestionTime,
		); err != nil {
			return nil, err
		}
		results = append(results, dp)
	}

	return results, nil
}
