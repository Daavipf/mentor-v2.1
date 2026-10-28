package exam

import (
	"database/sql"

	"mentor-api/internal/domain/repository"
)

type examRepository struct {
	db           *sql.DB
	questionRepo repository.QuestionRepository
}

func NewExamRepository(db *sql.DB, questionRepo repository.QuestionRepository) repository.ExamRepository {
	return &examRepository{
		db:           db,
		questionRepo: questionRepo,
	}
}
