package exam

import (
	"context"
	"time"

	"mentor-api/internal/domain"
	"mentor-api/internal/domain/entity"

	"uuid"
)

func (u *examUseCase) SubmitExam(ctx context.Context, examID, userID string, answers []AnswerQuestionInput) (*FinishExamResponse, error) {
	exam, err := u.examRepo.GetByID(ctx, examID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if exam.UserID != userID {
		return nil, domain.ErrForbidden
	}

	existingHistory, err := u.historyRepo.GetByExamID(ctx, examID)
	if err == nil && existingHistory != nil && existingHistory.FinishedAt != nil {
		return nil, domain.ErrExamAlreadyFinished
	}

	correctCount := 0
	for _, answer := range answers {
		qOnExam, err := u.examRepo.GetQuestionOnExam(ctx, examID, answer.QuestionID)
		if err != nil || qOnExam == nil {
			return nil, domain.ErrQuestionNotInExam
		}

		isCorrect, err := u.questionService.ValidateAnswer(ctx, answer.QuestionID, answer.SelectedAlternativeID)
		if err != nil {
			return nil, err
		}

		u.examRepo.AnswerQuestion(ctx, examID, answer.QuestionID, answer.SelectedAlternativeID, *answer.StartedAt, *answer.FinishedAt, isCorrect)
		if isCorrect {
			correctCount++
		}
	}

	totalQuestions := len(exam.Questions)

	score := 0
	if totalQuestions > 0 {
		score = (correctCount * 100) / totalQuestions
	}

	now := time.Now()
	history := &entity.History{
		ID:         uuid.New().String(),
		UserID:     userID,
		ExamID:     examID,
		Score:      &score,
		CreatedAt:  now,
		UpdatedAt:  now,
		FinishedAt: &now,
	}

	if err := u.historyRepo.Create(ctx, history); err != nil {
		return nil, err
	}

	return &FinishExamResponse{
		ExamID:         examID,
		TotalQuestions: totalQuestions,
		CorrectAnswers: correctCount,
		Score:          score,
		FinishedAt:     now,
	}, nil
}
