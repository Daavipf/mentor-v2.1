package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/usecase/analytics"
	"mentor-api/tests/mocks"
)

func TestAnalytics_GetOverview_Empty(t *testing.T) {
	mockRepo := &mocks.MockAnalyticsRepository{}
	uc := analytics.NewAnalyticsUseCase(mockRepo)

	ctx := context.Background()
	res, err := uc.GetOverview(ctx, "user-empty")
	require.NoError(t, err)

	assert.Equal(t, 0, res.TotalQuestionsAnswered)
	assert.Equal(t, 0, res.TotalCorrectAnswers)
	assert.Equal(t, 0.0, res.OverallAccuracyRate)
	assert.Equal(t, 0, res.TotalCompletedExams)
	assert.Equal(t, 0.0, res.AverageExamScore)
	assert.Empty(t, res.PerformanceByArea)
	assert.Empty(t, res.RecentStudyRhythm)
}

func TestAnalytics_GetOverview_Populated(t *testing.T) {
	mockRepo := &mocks.MockAnalyticsRepository{
		CustomOverview: &entity.OverviewMetrics{
			TotalQuestionsAnswered: 20,
			TotalCorrectAnswers:    15,
			OverallAccuracyRate:    75.0,
			TotalCompletedExams:    2,
			AverageExamScore:       85.5,
			PerformanceByArea: []entity.AreaPerformance{
				{Area: "Matemática", TotalQuestions: 10, CorrectAnswers: 8, AccuracyRate: 80.0},
				{Area: "Ciências Humanas", TotalQuestions: 10, CorrectAnswers: 7, AccuracyRate: 70.0},
			},
			RecentStudyRhythm: []entity.DailyStudyRhythm{
				{Date: "2026-10-25", TotalQuestions: 10, CorrectAnswers: 8},
				{Date: "2026-10-26", TotalQuestions: 10, CorrectAnswers: 7},
			},
		},
	}
	uc := analytics.NewAnalyticsUseCase(mockRepo)

	ctx := context.Background()
	res, err := uc.GetOverview(ctx, "user-123")
	require.NoError(t, err)

	assert.Equal(t, 20, res.TotalQuestionsAnswered)
	assert.Equal(t, 15, res.TotalCorrectAnswers)
	assert.Equal(t, 75.0, res.OverallAccuracyRate)
	assert.Equal(t, 2, res.TotalCompletedExams)
	assert.Equal(t, 85.5, res.AverageExamScore)
	assert.Len(t, res.PerformanceByArea, 2)
	assert.Len(t, res.RecentStudyRhythm, 2)
}

func TestAnalytics_GetOverview_Unauthorized(t *testing.T) {
	mockRepo := &mocks.MockAnalyticsRepository{}
	uc := analytics.NewAnalyticsUseCase(mockRepo)

	_, err := uc.GetOverview(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestAnalytics_GetByDiscipline(t *testing.T) {
	mockRepo := &mocks.MockAnalyticsRepository{
		CustomDisciplines: []entity.DisciplinePerformance{
			{Discipline: "Física", Area: "Ciências da Natureza", TotalQuestions: 15, CorrectAnswers: 12, AccuracyRate: 80.0},
			{Discipline: "Química", Area: "Ciências da Natureza", TotalQuestions: 10, CorrectAnswers: 5, AccuracyRate: 50.0},
		},
	}
	uc := analytics.NewAnalyticsUseCase(mockRepo)

	ctx := context.Background()
	res, err := uc.GetByDiscipline(ctx, "user-123", analytics.ByDisciplineFilterInput{
		Area: "Ciências da Natureza",
	})
	require.NoError(t, err)

	assert.Equal(t, 2, res.TotalDisciplines)
	assert.Equal(t, "Física", res.Disciplines[0].Discipline)
	assert.Equal(t, 80.0, res.Disciplines[0].AccuracyRate)
}

func TestAnalytics_GetTopicsRanking(t *testing.T) {
	mockRepo := &mocks.MockAnalyticsRepository{
		CustomTopics: &entity.TopicsRanking{
			StrongTopics: []entity.TopicPerformance{
				{Topic: "Eletromagnetismo", TotalQuestions: 12, CorrectAnswers: 11, AccuracyRate: 91.67},
				{Topic: "Termodinâmica", TotalQuestions: 10, CorrectAnswers: 8, AccuracyRate: 80.0},
			},
			WeakTopics: []entity.TopicPerformance{
				{Topic: "Óptica", TotalQuestions: 5, CorrectAnswers: 1, AccuracyRate: 20.0},
			},
		},
	}
	uc := analytics.NewAnalyticsUseCase(mockRepo)

	ctx := context.Background()
	res, err := uc.GetTopicsRanking(ctx, "user-123", 10)
	require.NoError(t, err)

	assert.Equal(t, 10, res.MinQuestionsThreshold)
	assert.Len(t, res.StrongTopics, 2)
	assert.Equal(t, "Eletromagnetismo", res.StrongTopics[0].Topic)
	assert.Len(t, res.WeakTopics, 1)
	assert.Equal(t, "Óptica", res.WeakTopics[0].Topic)
}

func TestAnalytics_GetEvolution(t *testing.T) {
	now := time.Now()
	mockRepo := &mocks.MockAnalyticsRepository{
		CustomEvolution: []entity.TimePeriodMetric{
			{Period: "2026-W40", PeriodDate: now.AddDate(0, 0, -14), TotalQuestions: 20, CorrectAnswers: 14, AccuracyRate: 70.0},
			{Period: "2026-W41", PeriodDate: now.AddDate(0, 0, -7), TotalQuestions: 25, CorrectAnswers: 20, AccuracyRate: 80.0},
		},
		CustomRhythm: []entity.DailyStudyRhythm{
			{Date: "2026-10-25", TotalQuestions: 10, CorrectAnswers: 8},
		},
	}
	uc := analytics.NewAnalyticsUseCase(mockRepo)

	ctx := context.Background()
	res, err := uc.GetEvolution(ctx, "user-123", "weekly", 6)
	require.NoError(t, err)

	assert.Equal(t, "weekly", res.PeriodType)
	assert.Len(t, res.Points, 2)
	assert.Equal(t, 70.0, res.Points[0].AccuracyRate)
	assert.Equal(t, 80.0, res.Points[1].AccuracyRate)
	assert.Len(t, res.StudyRhythm, 1)
}

func TestAnalytics_GetExamsHistory_EvolutionCalculation(t *testing.T) {
	now := time.Now()
	// Scenario: 4 exams with scores: 50, 60, 80, 90
	// Older avg (50, 60): 55.0
	// Recent avg (80, 90): 85.0
	// Evolution: ((85.0 - 55.0) / 55.0) * 100 = 54.55%
	mockRepo := &mocks.MockAnalyticsRepository{
		CustomExams: []entity.ExamPerformanceHistory{
			{ExamID: "e1", ExamTitle: "Simulado 1", Date: now.AddDate(0, 0, -30), FinishedAt: &now, Score: 50, TotalQuestions: 10, CorrectAnswers: 5},
			{ExamID: "e2", ExamTitle: "Simulado 2", Date: now.AddDate(0, 0, -20), FinishedAt: &now, Score: 60, TotalQuestions: 10, CorrectAnswers: 6},
			{ExamID: "e3", ExamTitle: "Simulado 3", Date: now.AddDate(0, 0, -10), FinishedAt: &now, Score: 80, TotalQuestions: 10, CorrectAnswers: 8},
			{ExamID: "e4", ExamTitle: "Simulado 4", Date: now, FinishedAt: &now, Score: 90, TotalQuestions: 10, CorrectAnswers: 9},
		},
	}
	uc := analytics.NewAnalyticsUseCase(mockRepo)

	ctx := context.Background()
	res, err := uc.GetExamsHistory(ctx, "user-123")
	require.NoError(t, err)

	assert.Equal(t, 4, res.TotalFinishedExams)
	assert.Equal(t, 70.0, res.OverallAverageScore) // (50+60+80+90) / 4 = 70.0
	assert.Equal(t, 85.0, res.RecentAverageScore)  // (80+90) / 2 = 85.0
	assert.Equal(t, 54.55, res.ScoreEvolutionPercentage)
}

