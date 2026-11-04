package mocks

import (
	"context"
	"sync"
	"time"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
)

type MockExamRepository struct {
	mu           sync.RWMutex
	exams        map[string]*entity.Exam
	questionRepo repository.QuestionRepository
}

func NewMockExamRepository(questionRepo repository.QuestionRepository) repository.ExamRepository {
	return &MockExamRepository{
		exams:        make(map[string]*entity.Exam),
		questionRepo: questionRepo,
	}
}

func (m *MockExamRepository) Create(ctx context.Context, exam *entity.Exam) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	eCopy := *exam
	m.exams[exam.ID] = &eCopy
	return nil
}

func (m *MockExamRepository) GetByID(ctx context.Context, id string) (*entity.Exam, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exam, exists := m.exams[id]
	if !exists {
		return nil, domain.ErrNotFound
	}

	eCopy := *exam
	return &eCopy, nil
}

func (m *MockExamRepository) ListByUserID(ctx context.Context, userID string) ([]entity.Exam, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []entity.Exam
	for _, exam := range m.exams {
		if exam.UserID == userID {
			result = append(result, *exam)
		}
	}
	return result, nil
}

func (m *MockExamRepository) Update(ctx context.Context, exam *entity.Exam) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.exams[exam.ID]; !exists {
		return domain.ErrNotFound
	}
	eCopy := *exam
	m.exams[exam.ID] = &eCopy
	return nil
}

func (m *MockExamRepository) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.exams[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.exams, id)
	return nil
}

func (m *MockExamRepository) AnswerQuestion(ctx context.Context, examID, questionID, alternativeID string, startedAt, finishedAt time.Time, gotRight bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	exam, exists := m.exams[examID]
	if !exists {
		return domain.ErrNotFound
	}

	found := false
	for i := range exam.Questions {
		if exam.Questions[i].QuestionID == questionID {
			exam.Questions[i].SelectedAlternativeID = &alternativeID
			exam.Questions[i].GotRight = &gotRight
			exam.Questions[i].UpdatedAt = time.Now()
			exam.Questions[i].StartedAt = &startedAt
			exam.Questions[i].FinishedAt = &finishedAt
			found = true
			break
		}
	}

	if !found {
		return domain.ErrQuestionNotInExam
	}

	return nil
}

func (m *MockExamRepository) GetQuestionOnExam(ctx context.Context, examID, questionID string) (*entity.QuestionsOnExam, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exam, exists := m.exams[examID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	for _, qe := range exam.Questions {
		if qe.QuestionID == questionID {
			qeCopy := qe
			return &qeCopy, nil
		}
	}

	return nil, domain.ErrQuestionNotInExam
}
