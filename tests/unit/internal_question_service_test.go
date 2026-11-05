package unit

import (
	"context"
	"testing"

	"mentor-api/internal/domain"
	"mentor-api/internal/usecase/question"
	"mentor-api/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string {
	return &s
}

func TestInternalQuestionService_CreateAndRetrieve(t *testing.T) {
	qRepo := mocks.NewMockQuestionRepository()
	service := mocks.NewQuestionServiceMock(qRepo)
	ctx := context.Background()

	input := question.CreateQuestionInput{
		Title:                    strPtr("Origem das Espécies"),
		Context:                  strPtr("A teoria da evolução pela seleção natural foi proposta por Charles Darwin."),
		AlternativesIntroduction: strPtr("Assinale a alternativa correta:"),
		Area:                     "Ciências da Natureza",
		Discipline:               strPtr("Biologia"),
		Topics:                   []string{"Evolução", "Genética"},
		Year:                     2023,
		Index:                    1,
		Images:                   []string{"/images/darwin.png"},
		Alternatives: []question.CreateAlternativeInput{
			{Text: strPtr("Organismos mais adaptados sobrevivem e se reproduzem"), IsCorrect: true},
			{Text: strPtr("A evolução ocorre pelo uso e desuso de órgãos"), IsCorrect: false},
		},
	}

	created, err := service.CreateQuestion(ctx, input)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "Ciências da Natureza", created.Area)
	assert.Len(t, created.Topics, 2)
	assert.Len(t, created.Images, 1)
	assert.Len(t, created.Alternatives, 2)

	// Validate answer
	var correctAltID, wrongAltID string
	for _, a := range created.Alternatives {
		if a.IsCorrect {
			correctAltID = a.ID
		} else {
			wrongAltID = a.ID
		}
	}

	isRight, err := service.ValidateAnswer(ctx, created.ID, correctAltID)
	require.NoError(t, err)
	assert.True(t, isRight)

	isWrong, err := service.ValidateAnswer(ctx, created.ID, wrongAltID)
	require.NoError(t, err)
	assert.False(t, isWrong)

	// Inexistent alternative
	_, err = service.ValidateAnswer(ctx, created.ID, "non-existent-alt-id")
	assert.ErrorIs(t, err, domain.ErrAlternativeNotFound)
}

func TestInternalQuestionService_SeedDefaultQuestions(t *testing.T) {
	qRepo := mocks.NewMockQuestionRepository()
	service := mocks.NewQuestionServiceMock(qRepo)
	ctx := context.Background()

	err := service.SeedDefaultQuestions(ctx, nil)
	require.NoError(t, err)

	questions, err := service.SelectQuestionsForExam(ctx, nil, 10)
	require.NoError(t, err)
	assert.Len(t, questions, 3)
}
