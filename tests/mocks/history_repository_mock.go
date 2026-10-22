package mocks

import (
	"context"
	"sync"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

type MockHistoryRepository struct {
	mu        sync.RWMutex
	histories map[string]*entity.History
}

func NewMockHistoryRepository() repository.HistoryRepository {
	return &MockHistoryRepository{
		histories: make(map[string]*entity.History),
	}
}

func (m *MockHistoryRepository) Create(ctx context.Context, h *entity.History) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	hCopy := *h
	m.histories[h.ID] = &hCopy
	return nil
}

func (m *MockHistoryRepository) GetByID(ctx context.Context, id string) (*entity.History, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	h, exists := m.histories[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	hCopy := *h
	return &hCopy, nil
}

func (m *MockHistoryRepository) GetByExamID(ctx context.Context, examID string) (*entity.History, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, h := range m.histories {
		if h.ExamID == examID {
			hCopy := *h
			return &hCopy, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockHistoryRepository) ListByUserID(ctx context.Context, userID string) ([]entity.History, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []entity.History
	for _, h := range m.histories {
		if h.UserID == userID {
			result = append(result, *h)
		}
	}
	return result, nil
}

func (m *MockHistoryRepository) Update(ctx context.Context, h *entity.History) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.histories[h.ID]; !exists {
		return domain.ErrNotFound
	}
	hCopy := *h
	m.histories[h.ID] = &hCopy
	return nil
}

