package analytics

import (
	"context"

	"mentor-api/internal/domain/entity"
)

func (r *analyticsRepository) GetTopicsRanking(ctx context.Context, userID string, minQuestions int) (*entity.TopicsRanking, error) {
	if minQuestions <= 0 {
		minQuestions = 10
	}

	ranking := &entity.TopicsRanking{
		StrongTopics: make([]entity.TopicPerformance, 0),
		WeakTopics:   make([]entity.TopicPerformance, 0),
	}

	// 1. Top 5 Strong Topics (highest accuracy rate with at least minQuestions)
	strongQuery := `
		SELECT
			qt.topic,
			COUNT(*) AS total_questions,
			COUNT(*) FILTER (WHERE qoe.got_right = true) AS correct_answers,
			COALESCE(ROUND((COUNT(*) FILTER (WHERE qoe.got_right = true)::numeric / NULLIF(COUNT(*), 0)) * 100, 2), 0) AS accuracy_rate
		FROM questions_on_exams qoe
		JOIN exams e ON e.id = qoe.exam_id
		JOIN question_topics qt ON qt.question_id = qoe.question_id
		WHERE e.user_id = $1
		  AND (qoe.selected_alternative_id IS NOT NULL OR qoe.got_right IS NOT NULL)
		GROUP BY qt.topic
		HAVING COUNT(*) >= $2
		ORDER BY accuracy_rate DESC, total_questions DESC, qt.topic ASC
		LIMIT 5
	`
	rowsStrong, err := r.db.QueryContext(ctx, strongQuery, userID, minQuestions)
	if err != nil {
		return nil, err
	}
	defer rowsStrong.Close()

	for rowsStrong.Next() {
		var tp entity.TopicPerformance
		if err := rowsStrong.Scan(
			&tp.Topic,
			&tp.TotalQuestions,
			&tp.CorrectAnswers,
			&tp.AccuracyRate,
		); err != nil {
			return nil, err
		}
		ranking.StrongTopics = append(ranking.StrongTopics, tp)
	}

	// 2. Top 5 Weak/Critical Topics (lowest accuracy rate)
	weakQuery := `
		SELECT
			qt.topic,
			COUNT(*) AS total_questions,
			COUNT(*) FILTER (WHERE qoe.got_right = true) AS correct_answers,
			COALESCE(ROUND((COUNT(*) FILTER (WHERE qoe.got_right = true)::numeric / NULLIF(COUNT(*), 0)) * 100, 2), 0) AS accuracy_rate
		FROM questions_on_exams qoe
		JOIN exams e ON e.id = qoe.exam_id
		JOIN question_topics qt ON qt.question_id = qoe.question_id
		WHERE e.user_id = $1
		  AND (qoe.selected_alternative_id IS NOT NULL OR qoe.got_right IS NOT NULL)
		GROUP BY qt.topic
		ORDER BY accuracy_rate ASC, total_questions DESC, qt.topic ASC
		LIMIT 5
	`
	rowsWeak, err := r.db.QueryContext(ctx, weakQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rowsWeak.Close()

	for rowsWeak.Next() {
		var tp entity.TopicPerformance
		if err := rowsWeak.Scan(
			&tp.Topic,
			&tp.TotalQuestions,
			&tp.CorrectAnswers,
			&tp.AccuracyRate,
		); err != nil {
			return nil, err
		}
		ranking.WeakTopics = append(ranking.WeakTopics, tp)
	}

	return ranking, nil
}

