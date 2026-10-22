package repository

import (
	"context"
	"mentor-api/internal/domain/entity"
)

// QuestionRepository defines the interface for question and alternative persistence.
// As per specifications, questions and alternatives are manipulated only within the system.
type QuestionRepository interface {
	Create(ctx context.Context, question *entity.Question) error
	GetByID(ctx context.Context, id string) (*entity.Question, error)
	GetAlternativeByID(ctx context.Context, alternativeID string) (*entity.Alternative, error)
	ListByAreas(ctx context.Context, areas []string, limit int) ([]entity.Question, error)
	ListRandom(ctx context.Context, limit int) ([]entity.Question, error)
	ListAll(ctx context.Context, limit, offset int) ([]entity.Question, error)
}

