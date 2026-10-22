package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"mentor-api/internal/domain"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SendSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func SendError(w http.ResponseWriter, err error) {
	statusCode := MapDomainErrorToHTTPStatus(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   err.Error(),
	})
}

func MapDomainErrorToHTTPStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrAlreadyExists), errors.Is(err, domain.ErrUserEmailConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrExamAlreadyFinished),
		errors.Is(err, domain.ErrQuestionNotInExam),
		errors.Is(err, domain.ErrAlternativeNotFound):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

