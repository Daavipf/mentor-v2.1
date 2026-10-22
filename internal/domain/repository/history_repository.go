package repository

import (
	"context"
	"mentor-api/internal/domain/entity"
)

// HistoryRepository defines the interface for history persistence.
type HistoryRepository interface {
	Create(ctx context.Context, history *entity.History) error
	GetByID(ctx context.Context, id string) (*entity.History, error)
	GetByExamID(ctx context.Context, examID string) (*entity.History, error)
	ListByUserID(ctx context.Context, userID string) ([]entity.History, error)
	Update(ctx context.Context, history *entity.History) error
}

