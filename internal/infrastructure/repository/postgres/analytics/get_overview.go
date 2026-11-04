package analytics

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *analyticsRepository) GetOverview(ctx context.Context, userID string) (*entity.OverviewMetrics, error) {
	metrics := &entity.OverviewMetrics{
		PerformanceByArea: make([]entity.AreaPerformance, 0),
		RecentStudyRhythm: make([]entity.DailyStudyRhythm, 0),
	}

	// 1. Total questions answered, correct answers and overall accuracy rate
	qoeQuery := `
		SELECT 
			COUNT(*) AS total_answered,
			COUNT(*) FILTER (WHERE qoe.got_right = true) AS total_correct,
			COALESCE(ROUND((COUNT(*) FILTER (WHERE qoe.got_right = true)::numeric / NULLIF(COUNT(*), 0)) * 100, 2), 0) AS accuracy_rate
		FROM questions_on_exams qoe
		JOIN exams e ON e.id = qoe.exam_id
		WHERE e.user_id = $1
		  AND (qoe.selected_alternative_id IS NOT NULL OR qoe.got_right IS NOT NULL)
	`
	err := r.db.QueryRowContext(ctx, qoeQuery, userID).Scan(
		&metrics.TotalQuestionsAnswered,
		&metrics.TotalCorrectAnswers,
		&metrics.OverallAccuracyRate,
	)
	if err != nil {
		return nil, err
	}

	// 2. Total completed exams and average exam score
	histQuery := `
		SELECT
			COUNT(*) FILTER (WHERE h.finished_at IS NOT NULL) AS total_completed_exams,
			COALESCE(ROUND(AVG(h.score) FILTER (WHERE h.finished_at IS NOT NULL AND h.score IS NOT NULL), 2), 0) AS average_score
		FROM histories h
		WHERE h.user_id = $1
	`
	err = r.db.QueryRowContext(ctx, histQuery, userID).Scan(
		&metrics.TotalCompletedExams,
		&metrics.AverageExamScore,
	)
	if err != nil {
		return nil, err
	}

	// 3. Performance by knowledge area
	areas, err := r.GetPerformanceByArea(ctx, userID)
	if err == nil && areas != nil {
		metrics.PerformanceByArea = areas
	}

	// 4. Study rhythm in the last 30 days
	rhythm, err := r.GetStudyRhythm(ctx, userID, 30)
	if err == nil && rhythm != nil {
		metrics.RecentStudyRhythm = rhythm
	}

	// 5. Average question completion time (in seconds)
	avgQuery := `
    SELECT
      COALESCE(
        ROUND(
          AVG(
            EXTRACT(EPOCH FROM (qoe.finished_at - qoe.started_at))
          ) FILTER (
            WHERE qoe.finished_at IS NOT NULL 
              AND qoe.started_at IS NOT NULL
              AND qoe.finished_at >= qoe.started_at
              AND (qoe.selected_alternative_id IS NOT NULL OR qoe.got_right IS NOT NULL)
          ), 
          2
        ), 
        0
      ) AS average_time_in_seconds
    FROM questions_on_exams qoe
    JOIN exams e ON e.id = qoe.exam_id
    WHERE e.user_id = $1
  `

	err = r.db.QueryRowContext(ctx, avgQuery, userID).Scan(&metrics.AverageQuestionTime)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}
