package analytics

import (
	"context"

	"mentor-api/internal/domain"
)

// GetPerformanceByArea busca as métricas de desempenho agrupadas por área para o usuário.
func (u *analyticsUseCase) GetPerformanceByArea(ctx context.Context, userID string) (*ByAreaResponse, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	performance, err := u.analyticsRepo.GetPerformanceByArea(ctx, userID)
	if err != nil {
		return nil, err
	}

	areasDTO := make([]AreaPerformanceDTO, 0, len(performance))
	for _, p := range performance {
		areasDTO = append(areasDTO, AreaPerformanceDTO{
			Area:                p.Area,
			TotalQuestions:      p.TotalQuestions,
			CorrectAnswers:      p.CorrectAnswers,
			AccuracyRate:        p.AccuracyRate,
			AverageQuestionTime: p.AverageQuestionTime,
		})
	}

	return &ByAreaResponse{
		Areas: areasDTO,
	}, nil
}
