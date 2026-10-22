package entity

import (
	"time"
)

// QuestionTopic represents a topic associated with a question.
type QuestionTopic struct {
	ID         string `json:"id"`
	QuestionID string `json:"question_id"`
	Topic      string `json:"topic"`
}

// QuestionImage represents an image associated with a question.
type QuestionImage struct {
	ID         string `json:"id"`
	QuestionID string `json:"question_id"`
	Path       string `json:"path"`
}

// Alternative represents a possible answer for a question.
type Alternative struct {
	ID         string  `json:"id"`
	QuestionID string  `json:"question_id"`
	Text       *string `json:"text,omitempty"`
	IsCorrect  bool    `json:"is_correct,omitempty"`
	ImagePath  *string `json:"image_path,omitempty"`
}

// Question represents a question entity according to Questions model in schema.prisma.
type Question struct {
	ID                       string          `json:"id"`
	Context                  *string         `json:"context,omitempty"`
	AlternativesIntroduction *string         `json:"alternatives_introduction,omitempty"`
	Area                     string          `json:"area"`
	Discipline               *string         `json:"discipline,omitempty"`
	Topics                   []QuestionTopic `json:"topics,omitempty"`
	Year                     int             `json:"year"`
	Images                   []QuestionImage `json:"images,omitempty"`
	Language                 *string         `json:"language,omitempty"`
	Index                    int             `json:"index"`
	Title                    *string         `json:"title,omitempty"`
	Alternatives             []Alternative   `json:"alternatives,omitempty"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
}

