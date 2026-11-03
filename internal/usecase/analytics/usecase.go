package analytics

import (
	"context"
	"math"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/repository"
)

// UseCase defines application business rules for analytics and user metrics.
type UseCase interface {
	GetOverview(ctx context.Context, userID string) (*OverviewResponse, error)
	GetByDiscipline(ctx context.Context, userID string, filter ByDisciplineFilterInput) (*ByDisciplineResponse, error)
	GetTopicsRanking(ctx context.Context, userID string, minQuestions int) (*TopicsRankingResponse, error)
	GetEvolution(ctx context.Context, userID string, periodType string, months int) (*EvolutionResponse, error)
	GetExamsHistory(ctx context.Context, userID string) (*ExamsHistoryResponse, error)
}

type analyticsUseCase struct {
	analyticsRepo repository.AnalyticsRepository
}

// NewAnalyticsUseCase instantiates the analytics use case implementation.
func NewAnalyticsUseCase(analyticsRepo repository.AnalyticsRepository) UseCase {
	return &analyticsUseCase{
		analyticsRepo: analyticsRepo,
	}
}

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

	return &OverviewResponse{
		TotalQuestionsAnswered: metrics.TotalQuestionsAnswered,
		TotalCorrectAnswers:    metrics.TotalCorrectAnswers,
		OverallAccuracyRate:    metrics.OverallAccuracyRate,
		TotalCompletedExams:    metrics.TotalCompletedExams,
		AverageExamScore:       metrics.AverageExamScore,
		PerformanceByArea:      areasDTO,
		RecentStudyRhythm:      rhythmDTO,
	}, nil
}

func (u *analyticsUseCase) GetByDiscipline(
	ctx context.Context,
	userID string,
	filter ByDisciplineFilterInput,
) (*ByDisciplineResponse, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	repoFilter := repository.DisciplineFilter{
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Area:      filter.Area,
	}

	disciplines, err := u.analyticsRepo.GetPerformanceByDiscipline(ctx, userID, repoFilter)
	if err != nil {
		return nil, err
	}

	list := make([]DisciplinePerformanceDTO, 0, len(disciplines))
	for _, d := range disciplines {
		list = append(list, DisciplinePerformanceDTO{
			Discipline:     d.Discipline,
			Area:           d.Area,
			TotalQuestions: d.TotalQuestions,
			CorrectAnswers: d.CorrectAnswers,
			AccuracyRate:   d.AccuracyRate,
		})
	}

	return &ByDisciplineResponse{
		Disciplines:      list,
		TotalDisciplines: len(list),
	}, nil
}

func (u *analyticsUseCase) GetTopicsRanking(
	ctx context.Context,
	userID string,
	minQuestions int,
) (*TopicsRankingResponse, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	if minQuestions <= 0 {
		minQuestions = 10
	}

	ranking, err := u.analyticsRepo.GetTopicsRanking(ctx, userID, minQuestions)
	if err != nil {
		return nil, err
	}

	strongDTO := make([]TopicPerformanceDTO, 0, len(ranking.StrongTopics))
	for _, s := range ranking.StrongTopics {
		strongDTO = append(strongDTO, TopicPerformanceDTO{
			Topic:          s.Topic,
			TotalQuestions: s.TotalQuestions,
			CorrectAnswers: s.CorrectAnswers,
			AccuracyRate:   s.AccuracyRate,
		})
	}

	weakDTO := make([]TopicPerformanceDTO, 0, len(ranking.WeakTopics))
	for _, w := range ranking.WeakTopics {
		weakDTO = append(weakDTO, TopicPerformanceDTO{
			Topic:          w.Topic,
			TotalQuestions: w.TotalQuestions,
			CorrectAnswers: w.CorrectAnswers,
			AccuracyRate:   w.AccuracyRate,
		})
	}

	return &TopicsRankingResponse{
		MinQuestionsThreshold: minQuestions,
		StrongTopics:          strongDTO,
		WeakTopics:            weakDTO,
	}, nil
}

func (u *analyticsUseCase) GetEvolution(
	ctx context.Context,
	userID string,
	periodType string,
	months int,
) (*EvolutionResponse, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	if months <= 0 {
		months = 6
	}

	evolution, err := u.analyticsRepo.GetEvolution(ctx, userID, periodType, months)
	if err != nil {
		return nil, err
	}

	points := make([]TimePeriodMetricDTO, 0, len(evolution))
	for _, p := range evolution {
		points = append(points, TimePeriodMetricDTO{
			Period:         p.Period,
			PeriodDate:     p.PeriodDate,
			TotalQuestions: p.TotalQuestions,
			CorrectAnswers: p.CorrectAnswers,
			AccuracyRate:   p.AccuracyRate,
		})
	}

	rhythm, err := u.analyticsRepo.GetStudyRhythm(ctx, userID, 30)
	if err != nil {
		rhythm = nil
	}

	rhythmDTO := make([]DailyStudyRhythmDTO, 0, len(rhythm))
	for _, r := range rhythm {
		rhythmDTO = append(rhythmDTO, DailyStudyRhythmDTO{
			Date:           r.Date,
			TotalQuestions: r.TotalQuestions,
			CorrectAnswers: r.CorrectAnswers,
		})
	}

	resolvedPeriod := "weekly"
	if periodType == "monthly" {
		resolvedPeriod = "monthly"
	}

	return &EvolutionResponse{
		PeriodType:  resolvedPeriod,
		Points:      points,
		StudyRhythm: rhythmDTO,
	}, nil
}

func (u *analyticsUseCase) GetExamsHistory(ctx context.Context, userID string) (*ExamsHistoryResponse, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	exams, err := u.analyticsRepo.GetExamsHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	examsDTO := make([]ExamPerformanceHistoryDTO, 0, len(exams))
	totalScoreSum := 0
	for _, e := range exams {
		examsDTO = append(examsDTO, ExamPerformanceHistoryDTO{
			ExamID:         e.ExamID,
			ExamTitle:      e.ExamTitle,
			Date:           e.Date,
			FinishedAt:     e.FinishedAt,
			Score:          e.Score,
			TotalQuestions: e.TotalQuestions,
			CorrectAnswers: e.CorrectAnswers,
		})
		totalScoreSum += e.Score
	}

	totalCount := len(exams)
	overallAverageScore := 0.0
	recentAverageScore := 0.0
	scoreEvolutionPercentage := 0.0

	if totalCount > 0 {
		overallAverageScore = math.Round((float64(totalScoreSum)/float64(totalCount))*100) / 100

		if totalCount == 1 {
			recentAverageScore = float64(exams[0].Score)
			scoreEvolutionPercentage = 0.0
		} else {
			// Compare older exams vs recent exams
			splitIndex := totalCount / 2
			olderSum := 0
			for i := 0; i < splitIndex; i++ {
				olderSum += exams[i].Score
			}
			olderAvg := float64(olderSum) / float64(splitIndex)

			recentCount := totalCount - splitIndex
			recentSum := 0
			for i := splitIndex; i < totalCount; i++ {
				recentSum += exams[i].Score
			}
			recentAverageScore = math.Round((float64(recentSum)/float64(recentCount))*100) / 100

			if olderAvg > 0 {
				scoreEvolutionPercentage = math.Round(((recentAverageScore-olderAvg)/olderAvg)*100*100) / 100
			} else if recentAverageScore > 0 {
				scoreEvolutionPercentage = 100.0
			}
		}
	}

	return &ExamsHistoryResponse{
		Exams:                    examsDTO,
		TotalFinishedExams:       totalCount,
		OverallAverageScore:      overallAverageScore,
		RecentAverageScore:       recentAverageScore,
		ScoreEvolutionPercentage: scoreEvolutionPercentage,
	}, nil
}

