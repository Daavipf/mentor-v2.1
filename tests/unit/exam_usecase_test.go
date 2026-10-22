package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mentor-api/internal/domain"
	"mentor-api/internal/usecase/exam"
	"mentor-api/internal/usecase/question"
	"mentor-api/tests/mocks"
)

func setupExamUseCase() (exam.UseCase, question.Service) {
	qRepo := mocks.NewMockQuestionRepository()
	eRepo := mocks.NewMockExamRepository(qRepo)
	hRepo := mocks.NewMockHistoryRepository()
	qService := question.NewQuestionService(qRepo)

	// Seed questions
	_ = qService.SeedDefaultQuestions(context.Background())

	examUC := exam.NewExamUseCase(eRepo, hRepo, qService)
	return examUC, qService
}

func TestExam_CreateAndRetrieve(t *testing.T) {
	examUC, _ := setupExamUseCase()
	ctx := context.Background()
	userID := "user-1"

	created, err := examUC.CreateExam(ctx, userID, exam.CreateExamInput{
		Title:         "Simulado Geral 2023",
		Areas:         []string{"Matemática", "História"},
		QuestionCount: 5,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, userID, created.UserID)
	assert.Equal(t, "Simulado Geral 2023", created.Title)
	assert.NotEmpty(t, created.Questions)
	assert.False(t, created.IsFinished)

	// Retrieve by owner
	retrieved, err := examUC.GetExamByID(ctx, userID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)

	// Check that is_correct is NOT leaked while exam is not finished
	for _, q := range retrieved.Questions {
		for _, alt := range q.Alternatives {
			assert.Nil(t, alt.IsCorrect, "is_correct must remain hidden before exam is finished")
		}
	}
}

func TestExam_UserIsolation_ForbiddenAccess(t *testing.T) {
	examUC, _ := setupExamUseCase()
	ctx := context.Background()
	user1 := "user-1"
	user2 := "user-2"

	// User 1 creates exam
	created, err := examUC.CreateExam(ctx, user1, exam.CreateExamInput{
		Title:         "Exame do Usuário 1",
		QuestionCount: 3,
	})
	require.NoError(t, err)

	// User 2 attempts to get User 1's exam -> MUST BE FORBIDDEN
	_, err = examUC.GetExamByID(ctx, user2, created.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	// User 2 attempts to answer User 1's exam -> MUST BE FORBIDDEN
	err = examUC.AnswerQuestion(ctx, user2, created.ID, exam.AnswerQuestionInput{
		QuestionID:            created.Questions[0].QuestionID,
		SelectedAlternativeID: created.Questions[0].Alternatives[0].ID,
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	// User 2 attempts to finish User 1's exam -> MUST BE FORBIDDEN
	_, err = examUC.FinishExam(ctx, user2, created.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	// User 2 attempts to delete User 1's exam -> MUST BE FORBIDDEN
	err = examUC.DeleteExam(ctx, user2, created.ID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	// User 2 listing exams should NOT include User 1's exam
	user2Exams, err := examUC.ListUserExams(ctx, user2)
	require.NoError(t, err)
	assert.Empty(t, user2Exams)
}

func TestExam_AnswerAndFinish(t *testing.T) {
	examUC, qService := setupExamUseCase()
	ctx := context.Background()
	userID := "user-1"

	created, err := examUC.CreateExam(ctx, userID, exam.CreateExamInput{
		Title:         "Simulado Completo",
		QuestionCount: 3,
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.Questions)

	// Answer questions
	for _, q := range created.Questions {
		// Look up full question from internal service to find correct answer
		fullQ, err := qService.GetQuestionByID(ctx, q.QuestionID)
		require.NoError(t, err)

		var selectedAltID string
		for _, alt := range fullQ.Alternatives {
			if alt.IsCorrect {
				selectedAltID = alt.ID
				break
			}
		}

		err = examUC.AnswerQuestion(ctx, userID, created.ID, exam.AnswerQuestionInput{
			QuestionID:            q.QuestionID,
			SelectedAlternativeID: selectedAltID,
		})
		require.NoError(t, err)
	}

	// Finish exam
	finishResp, err := examUC.FinishExam(ctx, userID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, finishResp.ExamID)
	assert.Equal(t, len(created.Questions), finishResp.TotalQuestions)
	assert.Equal(t, len(created.Questions), finishResp.CorrectAnswers)
	assert.Equal(t, 100, finishResp.Score)

	// Cannot answer after finished
	err = examUC.AnswerQuestion(ctx, userID, created.ID, exam.AnswerQuestionInput{
		QuestionID:            created.Questions[0].QuestionID,
		SelectedAlternativeID: created.Questions[0].Alternatives[0].ID,
	})
	assert.ErrorIs(t, err, domain.ErrExamAlreadyFinished)

	// After finished, GetExamByID reveals is_correct and got_right
	finishedExam, err := examUC.GetExamByID(ctx, userID, created.ID)
	require.NoError(t, err)
	assert.True(t, finishedExam.IsFinished)
	assert.NotNil(t, finishedExam.Score)
	assert.Equal(t, 100, *finishedExam.Score)

	for _, q := range finishedExam.Questions {
		assert.NotNil(t, q.GotRight)
		assert.True(t, *q.GotRight)
		for _, alt := range q.Alternatives {
			assert.NotNil(t, alt.IsCorrect, "is_correct should be revealed after exam is finished")
		}
	}
}

func TestExam_Delete_OwnerOnly(t *testing.T) {
	examUC, _ := setupExamUseCase()
	ctx := context.Background()
	userID := "user-1"

	created, err := examUC.CreateExam(ctx, userID, exam.CreateExamInput{
		Title:         "Exame para Deletar",
		QuestionCount: 2,
	})
	require.NoError(t, err)

	err = examUC.DeleteExam(ctx, userID, created.ID)
	require.NoError(t, err)

	_, err = examUC.GetExamByID(ctx, userID, created.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

