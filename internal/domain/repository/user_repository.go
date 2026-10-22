package repository

import (
	"context"
	"mentor-api/internal/domain/entity"
)

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}

