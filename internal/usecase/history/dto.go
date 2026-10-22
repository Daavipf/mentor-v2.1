package history

import "time"

type HistoryExamSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type HistoryResponse struct {
	ID         string              `json:"id"`
	UserID     string              `json:"user_id"`
	ExamID     string              `json:"exam_id"`
	Score      *int                `json:"score,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	FinishedAt *time.Time          `json:"finished_at,omitempty"`
	Exam       *HistoryExamSummary `json:"exam,omitempty"`
}

