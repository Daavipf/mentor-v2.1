package question

import (
	"context"
	"time"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"
	"mentor-api/internal/domain/repository"
	"mentor-api/internal/infrastructure/config"

	"github.com/google/uuid"
)

// Service defines internal methods for manipulating questions and alternatives.
// This service is strictly used internally by use cases and seeders, not exposed to HTTP.
type Service interface {
	CreateQuestion(ctx context.Context, input CreateQuestionInput) (*entity.Question, error)
	GetQuestionByID(ctx context.Context, id string) (*entity.Question, error)
	SelectQuestionsForExam(ctx context.Context, areas []string, count int) ([]entity.Question, error)
	ValidateAnswer(ctx context.Context, questionID string, alternativeID string) (bool, error)
	SeedDefaultQuestions(ctx context.Context, cfg *config.Config) error
}

type questionService struct {
	questionRepo repository.QuestionRepository
}

func NewQuestionService(questionRepo repository.QuestionRepository) Service {
	return &questionService{
		questionRepo: questionRepo,
	}
}

func (s *questionService) CreateQuestion(ctx context.Context, input CreateQuestionInput) (*entity.Question, error) {
	if input.Area == "" || len(input.Alternatives) == 0 {
		return nil, domain.ErrInvalidInput
	}

	questionID := uuid.New().String()
	now := time.Now()

	topics := make([]entity.QuestionTopic, 0, len(input.Topics))
	for _, t := range input.Topics {
		topics = append(topics, entity.QuestionTopic{
			ID:         uuid.New().String(),
			QuestionID: questionID,
			Topic:      t,
		})
	}

	images := make([]entity.QuestionImage, 0, len(input.Images))
	for _, img := range input.Images {
		images = append(images, entity.QuestionImage{
			ID:         uuid.New().String(),
			QuestionID: questionID,
			Path:       img,
		})
	}

	alternatives := make([]entity.Alternative, 0, len(input.Alternatives))
	for _, alt := range input.Alternatives {
		alternatives = append(alternatives, entity.Alternative{
			ID:         uuid.New().String(),
			QuestionID: questionID,
			Text:       alt.Text,
			IsCorrect:  alt.IsCorrect,
			ImagePath:  alt.ImagePath,
		})
	}

	question := &entity.Question{
		ID:                       questionID,
		Context:                  input.Context,
		AlternativesIntroduction: input.AlternativesIntroduction,
		Area:                     input.Area,
		Discipline:               input.Discipline,
		Topics:                   topics,
		Year:                     input.Year,
		Images:                   images,
		Language:                 input.Language,
		Index:                    input.Index,
		Title:                    input.Title,
		Alternatives:             alternatives,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	if err := s.questionRepo.Create(ctx, question); err != nil {
		return nil, err
	}

	return question, nil
}

func (s *questionService) GetQuestionByID(ctx context.Context, id string) (*entity.Question, error) {
	return s.questionRepo.GetByID(ctx, id)
}

func (s *questionService) SelectQuestionsForExam(ctx context.Context, areas []string, count int) ([]entity.Question, error) {
	if count <= 0 {
		count = 10
	}

	if len(areas) > 0 {
		return s.questionRepo.ListByAreas(ctx, areas, count)
	}

	return s.questionRepo.ListRandom(ctx, count)
}

func (s *questionService) ValidateAnswer(ctx context.Context, questionID string, alternativeID string) (bool, error) {
	alt, err := s.questionRepo.GetAlternativeByID(ctx, alternativeID)
	if err != nil {
		return false, domain.ErrAlternativeNotFound
	}

	if alt.QuestionID != questionID {
		return false, domain.ErrAlternativeNotFound
	}

	return alt.IsCorrect, nil
}
