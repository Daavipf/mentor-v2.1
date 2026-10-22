package exam

import "time"

type CreateExamInput struct {
	Title         string   `json:"title"`
	Date          *time.Time `json:"date,omitempty"`
	Areas         []string `json:"areas"`
	QuestionCount int      `json:"question_count"`
}

type AnswerQuestionInput struct {
	QuestionID            string `json:"question_id"`
	SelectedAlternativeID string `json:"selected_alternative_id"`
}

type ExamAlternativeDTO struct {
	ID        string  `json:"id"`
	Text      *string `json:"text,omitempty"`
	ImagePath *string `json:"image_path,omitempty"`
	IsCorrect *bool   `json:"is_correct,omitempty"` // Only revealed when exam is finished
}

type ExamQuestionDTO struct {
	QuestionID               string               `json:"question_id"`
	Context                  *string              `json:"context,omitempty"`
	AlternativesIntroduction *string              `json:"alternatives_introduction,omitempty"`
	Area                     string               `json:"area"`
	Discipline               *string              `json:"discipline,omitempty"`
	Year                     int                  `json:"year"`
	Language                 *string              `json:"language,omitempty"`
	Index                    int                  `json:"index"`
	Title                    *string              `json:"title,omitempty"`
	Topics                   []string             `json:"topics,omitempty"`
	Images                   []string             `json:"images,omitempty"`
	Alternatives             []ExamAlternativeDTO `json:"alternatives"`
	SelectedAlternativeID    *string              `json:"selected_alternative_id,omitempty"`
	GotRight                 *bool                `json:"got_right,omitempty"` // Only revealed when exam is finished
}

type ExamSummaryResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Title          string    `json:"title"`
	Date           time.Time `json:"date"`
	Areas          []string  `json:"areas"`
	TotalQuestions int       `json:"total_questions"`
	IsFinished     bool      `json:"is_finished"`
	Score          *int      `json:"score,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ExamResponse struct {
	ID             string            `json:"id"`
	UserID         string            `json:"user_id"`
	Title          string            `json:"title"`
	Date           time.Time         `json:"date"`
	Areas          []string          `json:"areas"`
	Questions      []ExamQuestionDTO `json:"questions"`
	TotalQuestions int               `json:"total_questions"`
	IsFinished     bool              `json:"is_finished"`
	Score          *int              `json:"score,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type FinishExamResponse struct {
	ExamID         string    `json:"exam_id"`
	TotalQuestions int       `json:"total_questions"`
	CorrectAnswers int       `json:"correct_answers"`
	Score          int       `json:"score"`
	FinishedAt     time.Time `json:"finished_at"`
}

