package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"mentor-api/internal/domain"
	"mentor-api/internal/interface/http/middleware"
	"mentor-api/internal/usecase/exam"
)

type ExamHandler struct {
	examUseCase exam.UseCase
}

func NewExamHandler(examUseCase exam.UseCase) *ExamHandler {
	return &ExamHandler{
		examUseCase: examUseCase,
	}
}

func (h *ExamHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	var input exam.CreateExamInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	res, err := h.examUseCase.CreateExam(r.Context(), userID, input)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusCreated, res)
}

func (h *ExamHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	res, err := h.examUseCase.ListUserExams(r.Context(), userID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

func (h *ExamHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	examID := chi.URLParam(r, "id")
	if examID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	res, err := h.examUseCase.GetExamByID(r.Context(), userID, examID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

func (h *ExamHandler) AnswerQuestion(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	examID := chi.URLParam(r, "id")
	if examID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	var input exam.AnswerQuestionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	err := h.examUseCase.AnswerQuestion(r.Context(), userID, examID, input)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Answer recorded successfully",
	})
}

func (h *ExamHandler) Finish(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	examID := chi.URLParam(r, "id")
	if examID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	res, err := h.examUseCase.FinishExam(r.Context(), userID, examID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

func (h *ExamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	examID := chi.URLParam(r, "id")
	if examID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	err := h.examUseCase.DeleteExam(r.Context(), userID, examID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Exam deleted successfully",
	})
}

