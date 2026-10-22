package mocks

import (
	"context"
	"sync"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

type MockUserRepository struct {
	mu    sync.RWMutex
	users map[string]*entity.User
}

func NewMockUserRepository() repository.UserRepository {
	return &MockUserRepository{
		users: make(map[string]*entity.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.users {
		if u.Email == user.Email {
			return domain.ErrUserEmailConflict
		}
	}

	uCopy := *user
	m.users[user.ID] = &uCopy
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	uCopy := *user
	return &uCopy, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == email {
			uCopy := *user
			return &uCopy, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.ID]; !exists {
		return domain.ErrNotFound
	}
	uCopy := *user
	m.users[user.ID] = &uCopy
	return nil
}

