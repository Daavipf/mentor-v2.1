package analytics

import (
	"context"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/repository"
)

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
