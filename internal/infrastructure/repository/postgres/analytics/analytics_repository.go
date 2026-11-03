package analytics

import (
	"database/sql"

	"mentor-api/internal/domain/repository"
)

type analyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) repository.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

