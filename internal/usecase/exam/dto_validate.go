package exam

import "mentor-api/internal/domain"

func (i AnswerQuestionInput) Validate() error {
	if i.StartedAt == nil || i.FinishedAt == nil {
		return domain.ErrInvalidInput
	}

	return nil
}
