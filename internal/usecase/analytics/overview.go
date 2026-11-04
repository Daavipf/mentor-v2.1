package analytics

import (
	"context"

	"mentor-api/internal/domain"
)

func (u *analyticsUseCase) GetOverview(ctx context.Context, userID string) (*OverviewResponse, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	metrics, err := u.analyticsRepo.GetOverview(ctx, userID)
	if err != nil {
		return nil, err
	}

	areasDTO := make([]AreaPerformanceDTO, 0, len(metrics.PerformanceByArea))
	for _, a := range metrics.PerformanceByArea {
		areasDTO = append(areasDTO, AreaPerformanceDTO{
			Area:           a.Area,
			TotalQuestions: a.TotalQuestions,
			CorrectAnswers: a.CorrectAnswers,
			AccuracyRate:   a.AccuracyRate,
		})
	}

	rhythmDTO := make([]DailyStudyRhythmDTO, 0, len(metrics.RecentStudyRhythm))
	for _, r := range metrics.RecentStudyRhythm {
		rhythmDTO = append(rhythmDTO, DailyStudyRhythmDTO{
			Date:           r.Date,
			TotalQuestions: r.TotalQuestions,
			CorrectAnswers: r.CorrectAnswers,
		})
	}

	avgQTime := metrics.AverageQuestionTime

	return &OverviewResponse{
		TotalQuestionsAnswered: metrics.TotalQuestionsAnswered,
		TotalCorrectAnswers:    metrics.TotalCorrectAnswers,
		OverallAccuracyRate:    metrics.OverallAccuracyRate,
		TotalCompletedExams:    metrics.TotalCompletedExams,
		AverageExamScore:       metrics.AverageExamScore,
		PerformanceByArea:      areasDTO,
		RecentStudyRhythm:      rhythmDTO,
		AverageQuestionTime:    avgQTime,
	}, nil
}
