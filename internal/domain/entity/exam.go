package entity

import (
	"time"
)

// ExamArea represents an area associated with an exam.
type ExamArea struct {
	ID      string `json:"id"`
	ExamsID string `json:"exams_id"`
	Area    string `json:"area"`
}

// QuestionsOnExam represents the link between an exam and a question, storing user answer state.
type QuestionsOnExam struct {
	QuestionID            string       `json:"question_id"`
	ExamID                string       `json:"exam_id"`
	GotRight              *bool        `json:"got_right,omitempty"`
	SelectedAlternativeID *string      `json:"selected_alternative_id,omitempty"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
	Question              *Question    `json:"question,omitempty"`
	UserAlternative       *Alternative `json:"user_alternative,omitempty"`
}

// Exam represents an exam/simulado entity according to Exams model in schema.prisma.
type Exam struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Date      time.Time         `json:"date"`
	Title     string            `json:"title"`
	Areas     []ExamArea        `json:"areas,omitempty"`
	Questions []QuestionsOnExam `json:"questions,omitempty"`
	Histories []History         `json:"histories,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

