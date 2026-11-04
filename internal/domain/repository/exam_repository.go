package repository

import (
	"context"
	"mentor-api/internal/domain/entity"
	"time"
)

// ExamRepository defines the interface for exam persistence and exam questions management.
type ExamRepository interface {
	Create(ctx context.Context, exam *entity.Exam) error
	GetByID(ctx context.Context, id string) (*entity.Exam, error)
	ListByUserID(ctx context.Context, userID string) ([]entity.Exam, error)
	Update(ctx context.Context, exam *entity.Exam) error
	Delete(ctx context.Context, id string) error
	AnswerQuestion(ctx context.Context, examID, questionID, alternativeID string, startedAt, finishedAt time.Time, gotRight bool) error
	GetQuestionOnExam(ctx context.Context, examID, questionID string) (*entity.QuestionsOnExam, error)
}
