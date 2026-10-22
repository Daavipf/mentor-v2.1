package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"mentor-api/internal/domain"
	"mentor-api/internal/interface/http/middleware"
	"mentor-api/internal/usecase/history"
)

type HistoryHandler struct {
	historyUseCase history.UseCase
}

func NewHistoryHandler(historyUseCase history.UseCase) *HistoryHandler {
	return &HistoryHandler{
		historyUseCase: historyUseCase,
	}
}

func (h *HistoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	res, err := h.historyUseCase.ListUserHistories(r.Context(), userID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

func (h *HistoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	historyID := chi.URLParam(r, "id")
	if historyID == "" {
		SendError(w, domain.ErrInvalidInput)
		return
	}

	res, err := h.historyUseCase.GetHistoryByID(r.Context(), userID, historyID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

