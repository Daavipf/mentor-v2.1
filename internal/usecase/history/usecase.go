package history

import (
	"context"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/repository"
)

type UseCase interface {
	ListUserHistories(ctx context.Context, userID string) ([]HistoryResponse, error)
	GetHistoryByID(ctx context.Context, userID string, historyID string) (*HistoryResponse, error)
}

type historyUseCase struct {
	historyRepo repository.HistoryRepository
	examRepo    repository.ExamRepository
}

func NewHistoryUseCase(historyRepo repository.HistoryRepository, examRepo repository.ExamRepository) UseCase {
	return &historyUseCase{
		historyRepo: historyRepo,
		examRepo:    examRepo,
	}
}

func (u *historyUseCase) ListUserHistories(ctx context.Context, userID string) ([]HistoryResponse, error) {
	histories, err := u.historyRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]HistoryResponse, 0, len(histories))
	for _, h := range histories {
		var examSummary *HistoryExamSummary
		if exam, err := u.examRepo.GetByID(ctx, h.ExamID); err == nil && exam != nil {
			examSummary = &HistoryExamSummary{
				ID:    exam.ID,
				Title: exam.Title,
			}
		}

		result = append(result, HistoryResponse{
			ID:         h.ID,
			UserID:     h.UserID,
			ExamID:     h.ExamID,
			Score:      h.Score,
			CreatedAt:  h.CreatedAt,
			UpdatedAt:  h.UpdatedAt,
			FinishedAt: h.FinishedAt,
			Exam:       examSummary,
		})
	}

	return result, nil
}

func (u *historyUseCase) GetHistoryByID(ctx context.Context, userID string, historyID string) (*HistoryResponse, error) {
	h, err := u.historyRepo.GetByID(ctx, historyID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// Strict User Ownership Check: Only owner can access their history
	if h.UserID != userID {
		return nil, domain.ErrForbidden
	}

	var examSummary *HistoryExamSummary
	if exam, err := u.examRepo.GetByID(ctx, h.ExamID); err == nil && exam != nil {
		examSummary = &HistoryExamSummary{
			ID:    exam.ID,
			Title: exam.Title,
		}
	}

	return &HistoryResponse{
		ID:         h.ID,
		UserID:     h.UserID,
		ExamID:     h.ExamID,
		Score:      h.Score,
		CreatedAt:  h.CreatedAt,
		UpdatedAt:  h.UpdatedAt,
		FinishedAt: h.FinishedAt,
		Exam:       examSummary,
	}, nil
}

