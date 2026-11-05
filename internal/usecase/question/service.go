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

/*func (s *questionService) SeedDefaultQuestions(ctx context.Context) error {
	sampleQuestions := []CreateQuestionInput{
		{
			Title:                    ptr("Revolução Francesa"),
			Context:                  ptr("A Revolução Francesa iniciou-se em 1789 com a queda da Bastilha."),
			AlternativesIntroduction: ptr("Qual das seguintes alternativas representa um dos ideais fundamentais do movimento?"),
			Area:                     "Ciências Humanas",
			Discipline:               ptr("História"),
			Topics:                   []string{"Revolução Francesa", "Idade Contemporânea"},
			Year:                     2023,
			Index:                    1,
			Alternatives: []CreateAlternativeInput{
				{Text: ptr("Liberdade, Igualdade e Fraternidade"), IsCorrect: true},
				{Text: ptr("Ordem e Progresso"), IsCorrect: false},
				{Text: ptr("Paz, Pão e Terra"), IsCorrect: false},
				{Text: ptr("Deus, Pátria e Família"), IsCorrect: false},
			},
		},
		{
			Title:                    ptr("Equação do Segundo Grau"),
			Context:                  ptr("Considere a equação x^2 - 5x + 6 = 0."),
			AlternativesIntroduction: ptr("As raízes reais dessa equação são:"),
			Area:                     "Matemática",
			Discipline:               ptr("Álgebra"),
			Topics:                   []string{"Equações", "Polinômios"},
			Year:                     2023,
			Index:                    2,
			Alternatives: []CreateAlternativeInput{
				{Text: ptr("2 e 3"), IsCorrect: true},
				{Text: ptr("1 e 6"), IsCorrect: false},
				{Text: ptr("-2 e -3"), IsCorrect: false},
				{Text: ptr("0 e 5"), IsCorrect: false},
			},
		},
		{
			Title:                    ptr("Leis de Newton"),
			Context:                  ptr("A Segunda Lei de Newton estabelece a relação entre força resultante, massa e aceleração."),
			AlternativesIntroduction: ptr("A fórmula correspondente é:"),
			Area:                     "Ciências da Natureza",
			Discipline:               ptr("Física"),
			Topics:                   []string{"Dinâmica", "Mecânica Clássica"},
			Year:                     2023,
			Index:                    3,
			Alternatives: []CreateAlternativeInput{
				{Text: ptr("F = m * a"), IsCorrect: true},
				{Text: ptr("E = m * c^2"), IsCorrect: false},
				{Text: ptr("V = d / t"), IsCorrect: false},
				{Text: ptr("P = m * g * h"), IsCorrect: false},
			},
		},
	}

	for _, q := range sampleQuestions {
		if _, err := s.CreateQuestion(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

func ptr(s string) *string {
	return &s
}
*/
