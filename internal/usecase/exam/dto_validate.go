package exam

import "mentor-api/internal/domain"

func (i AnswerQuestionInput) Validate() error {
	if i.QuestionID == "" || i.SelectedAlternativeID == "" {
		return domain.ErrInvalidInput
	}

	return nil
}
