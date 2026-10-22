package question

type CreateAlternativeInput struct {
	Text      *string `json:"text,omitempty"`
	IsCorrect bool    `json:"is_correct"`
	ImagePath *string `json:"image_path,omitempty"`
}

type CreateTopicInput struct {
	Topic string `json:"topic"`
}

type CreateImageInput struct {
	Path string `json:"path"`
}

type CreateQuestionInput struct {
	Context                  *string                  `json:"context,omitempty"`
	AlternativesIntroduction *string                  `json:"alternatives_introduction,omitempty"`
	Area                     string                   `json:"area"`
	Discipline               *string                  `json:"discipline,omitempty"`
	Topics                   []string                 `json:"topics,omitempty"`
	Year                     int                      `json:"year"`
	Images                   []string                 `json:"images,omitempty"`
	Language                 *string                  `json:"language,omitempty"`
	Index                    int                      `json:"index"`
	Title                    *string                  `json:"title,omitempty"`
	Alternatives             []CreateAlternativeInput `json:"alternatives"`
}

