package history

import (
	"database/sql"

	"mentor-api/internal/domain/repository"
)

type historyRepository struct {
	db *sql.DB
}

func NewHistoryRepository(db *sql.DB) repository.HistoryRepository {
	return &historyRepository{db: db}
}

