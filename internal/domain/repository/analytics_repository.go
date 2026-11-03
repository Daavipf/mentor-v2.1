package repository

import (
	"context"
	"time"

	"mentor-api/internal/domain/entity"
)

// DisciplineFilter defines optional query filters for performance by discipline.
type DisciplineFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	Area      string
}

// AnalyticsRepository defines persistence methods for analytics and performance aggregation.
type AnalyticsRepository interface {
	GetOverview(ctx context.Context, userID string) (*entity.OverviewMetrics, error)
	GetPerformanceByArea(ctx context.Context, userID string) ([]entity.AreaPerformance, error)
	GetPerformanceByDiscipline(ctx context.Context, userID string, filter DisciplineFilter) ([]entity.DisciplinePerformance, error)
	GetTopicsRanking(ctx context.Context, userID string, minQuestions int) (*entity.TopicsRanking, error)
	GetEvolution(ctx context.Context, userID string, periodType string, months int) ([]entity.TimePeriodMetric, error)
	GetStudyRhythm(ctx context.Context, userID string, days int) ([]entity.DailyStudyRhythm, error)
	GetExamsHistory(ctx context.Context, userID string) ([]entity.ExamPerformanceHistory, error)
}

