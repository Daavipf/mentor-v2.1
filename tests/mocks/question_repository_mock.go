package mocks

import (
	"context"
	"math/rand"
	"sync"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

type MockQuestionRepository struct {
	mu           sync.RWMutex
	questions    map[string]*entity.Question
	alternatives map[string]*entity.Alternative
}

func NewMockQuestionRepository() repository.QuestionRepository {
	return &MockQuestionRepository{
		questions:    make(map[string]*entity.Question),
		alternatives: make(map[string]*entity.Alternative),
	}
}

func (m *MockQuestionRepository) Create(ctx context.Context, q *entity.Question) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	qCopy := *q
	m.questions[q.ID] = &qCopy

	for _, alt := range q.Alternatives {
		altCopy := alt
		m.alternatives[alt.ID] = &altCopy
	}

	return nil
}

func (m *MockQuestionRepository) GetByID(ctx context.Context, id string) (*entity.Question, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	q, exists := m.questions[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	qCopy := *q
	return &qCopy, nil
}

func (m *MockQuestionRepository) GetAlternativeByID(ctx context.Context, alternativeID string) (*entity.Alternative, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alt, exists := m.alternatives[alternativeID]
	if !exists {
		return nil, domain.ErrNotFound
	}
	altCopy := *alt
	return &altCopy, nil
}

func (m *MockQuestionRepository) ListByAreas(ctx context.Context, areas []string, limit int) ([]entity.Question, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	areaMap := make(map[string]bool)
	for _, a := range areas {
		areaMap[a] = true
	}

	var matched []entity.Question
	for _, q := range m.questions {
		if areaMap[q.Area] {
			matched = append(matched, *q)
			if limit > 0 && len(matched) >= limit {
				break
			}
		}
	}

	// If no matched by specific area, fallback to all available
	if len(matched) == 0 {
		return m.ListRandom(ctx, limit)
	}

	return matched, nil
}

func (m *MockQuestionRepository) ListRandom(ctx context.Context, limit int) ([]entity.Question, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []entity.Question
	for _, q := range m.questions {
		all = append(all, *q)
	}

	rand.Shuffle(len(all), func(i, j int) {
		all[i], all[j] = all[j], all[i]
	})

	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}

	return all, nil
}

func (m *MockQuestionRepository) ListAll(ctx context.Context, limit, offset int) ([]entity.Question, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []entity.Question
	for _, q := range m.questions {
		all = append(all, *q)
	}

	if offset >= len(all) {
		return []entity.Question{}, nil
	}

	end := offset + limit
	if end > len(all) {
		end = len(all)
	}

	return all[offset:end], nil
}

