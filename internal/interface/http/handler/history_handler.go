package handler

import (
	"net/http"

	"mentor-api/internal/domain"
	"mentor-api/internal/interface/http/middleware"
	"mentor-api/internal/usecase/history"

	"github.com/go-chi/chi/v5"
)

type HistoryHandler struct {
	historyUseCase history.UseCase
}

func NewHistoryHandler(historyUseCase history.UseCase) *HistoryHandler {
	return &HistoryHandler{
		historyUseCase: historyUseCase,
	}
}

// List godoc
// @Summary      Lista os históricos de simulados do usuário
// @Description  Retorna a lista simplificada de todos os históricos de simulados finalizados pelo usuário autenticado
// @Tags         Histories
// @Produce      json
// @Success      200  {array}   history.HistoryExamSummary  "Históricos retornados com sucesso"
// @Failure      401  {object}  map[string]string             "Não autorizado / Token JWT ausente ou inválido"
// @Failure      500  {object}  map[string]string             "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /histories [get]
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

// GetByID godoc
// @Summary      Busca um histórico por ID
// @Description  Retorna os detalhes completos de um histórico de simulado específico pertencente ao usuário autenticado
// @Tags         Histories
// @Produce      json
// @Param        id   path      string                  true  "ID do Histórico (UUID)"
// @Success      200  {object}  history.HistoryResponse   "Histórico encontrado com sucesso"
// @Failure      400  {object}  map[string]string       "ID do histórico inválido ou ausente"
// @Failure      401  {object}  map[string]string       "Não autorizado / Token JWT ausente ou inválido"
// @Failure      404  {object}  map[string]string       "Histórico não encontrado"
// @Failure      500  {object}  map[string]string       "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /histories/{id} [get]
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
