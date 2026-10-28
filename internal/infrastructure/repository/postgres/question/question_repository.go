package question

import (
	"database/sql"

	"mentor-api/internal/domain/repository"
)

type questionRepository struct {
	db *sql.DB
}

func NewQuestionRepository(db *sql.DB) repository.QuestionRepository {
	return &questionRepository{db: db}
}

