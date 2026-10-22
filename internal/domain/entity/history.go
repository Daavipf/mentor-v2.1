package entity

import (
	"time"
)

// History represents the historical attempt and final score of an exam according to History model in schema.prisma.
type History struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	ExamID     string     `json:"exam_id"`
	Score      *int       `json:"score,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Exam       *Exam      `json:"exam,omitempty"`
}

