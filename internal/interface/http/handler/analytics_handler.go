package handler

import (
	"net/http"
	"strconv"
	"time"

	"mentor-api/internal/domain"
	"mentor-api/internal/interface/http/middleware"
	"mentor-api/internal/usecase/analytics"
)

type AnalyticsHandler struct {
	analyticsUseCase analytics.UseCase
}

func NewAnalyticsHandler(analyticsUseCase analytics.UseCase) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsUseCase: analyticsUseCase,
	}
}

// Overview godoc
// @Summary      Visão geral do desempenho
// @Description  Retorna o resumo geral de métricas e taxa de acertos do usuário autenticado
// @Tags         Analytics
// @Produce      json
// @Success      200  {object}  analytics.OverviewResponse  "Visão geral retornada com sucesso"
// @Failure      401  {object}  map[string]string         "Não autorizado / Token JWT ausente ou inválido"
// @Failure      500  {object}  map[string]string         "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /analytics/overview [get]
func (h *AnalyticsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	res, err := h.analyticsUseCase.GetOverview(r.Context(), userID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

// ByArea godoc
// @Summary      Desempenho por grande área do conhecimento
// @Description  Retorna o desempenho consolidado agrupado por área do conhecimento (ex: Ciências da Natureza, Humanas, etc.)
// @Tags         Analytics
// @Produce      json
// @Success      200  {array}   analytics.ByAreaResponse  "Desempenho por área retornado com sucesso"
// @Failure      401  {object}  map[string]string                "Não autorizado / Token JWT ausente ou inválido"
// @Failure      500  {object}  map[string]string                "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /analytics/by-area [get]
func (h *AnalyticsHandler) ByArea(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	res, err := h.analyticsUseCase.GetPerformanceByArea(r.Context(), userID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

// ByDiscipline godoc
// @Summary      Desempenho por disciplina
// @Description  Retorna as métricas detalhadas filtradas por disciplina, com opções de filtro por área e intervalo de datas
// @Tags         Analytics
// @Produce      json
// @Param        area        query     string  false  "Filtra por área do conhecimento"
// @Param        start_date  query     string  false  "Data inicial (Formatos: YYYY-MM-DD ou RFC3339)"  example("2026-01-01")
// @Param        end_date    query     string  false  "Data final (Formatos: YYYY-MM-DD ou RFC3339)"    example("2026-12-31")
// @Success      200         {array}   analytics.ByDisciplineResponse  "Desempenho por disciplina retornado com sucesso"
// @Failure      401         {object}  map[string]string                      "Não autorizado / Token JWT ausente ou inválido"
// @Failure      500         {object}  map[string]string                      "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /analytics/by-discipline [get]
func (h *AnalyticsHandler) ByDiscipline(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	filter := analytics.ByDisciplineFilterInput{
		Area: r.URL.Query().Get("area"),
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if t, err := parseDate(startDateStr); err == nil {
			filter.StartDate = &t
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if t, err := parseDate(endDateStr); err == nil {
			filter.EndDate = &t
		}
	}

	res, err := h.analyticsUseCase.GetByDiscipline(r.Context(), userID, filter)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

// TopicsRanking godoc
// @Summary      Ranking de tópicos com mais erros/acertos
// @Description  Retorna a classificação de assuntos com base na taxa de desempenho, exigindo um número mínimo de questões respondidas
// @Tags         Analytics
// @Produce      json
// @Param        min_questions  query     int    false  "Quantidade mínima de questões para incluir o tópico (Padrão: 10)"  default(10)
// @Success      200            {array}   analytics.TopicsRankingResponse  "Ranking de tópicos retornado com sucesso"
// @Failure      401            {object}  map[string]string             "Não autorizado / Token JWT ausente ou inválido"
// @Failure      500            {object}  map[string]string             "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /analytics/topics-ranking [get]
func (h *AnalyticsHandler) TopicsRanking(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	minQuestions := 10
	if minStr := r.URL.Query().Get("min_questions"); minStr != "" {
		if val, err := strconv.Atoi(minStr); err == nil && val > 0 {
			minQuestions = val
		}
	}

	res, err := h.analyticsUseCase.GetTopicsRanking(r.Context(), userID, minQuestions)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

// Evolution godoc
// @Summary      Evolução temporal do usuário
// @Description  Retorna o histórico de evolução de desempenho por períodos (ex: semanal/mensal) retroagindo um número determinado de meses
// @Tags         Analytics
// @Produce      json
// @Param        period  query     string  false  "Periodicidade do agrupamento (weekly, monthly)"  default(weekly)  Enums(weekly, monthly)
// @Param        months  query     int     false  "Quantidade de meses retroativos a analisar (Padrão: 6)"     default(6)
// @Success      200     {array}   analytics.EvolutionResponse  "Dados de evolução retornados com sucesso"
// @Failure      401     {object}  map[string]string               "Não autorizado / Token JWT ausente ou inválido"
// @Failure      500     {object}  map[string]string               "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /analytics/evolution [get]
func (h *AnalyticsHandler) Evolution(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "weekly"
	}

	months := 6
	if monthsStr := r.URL.Query().Get("months"); monthsStr != "" {
		if val, err := strconv.Atoi(monthsStr); err == nil && val > 0 {
			months = val
		}
	}

	res, err := h.analyticsUseCase.GetEvolution(r.Context(), userID, period, months)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

// ExamsHistory godoc
// @Summary      Histórico analítico de simulados
// @Description  Retorna o histórico detalhado de simulados realizados para fins de gráficos e métricas temporais
// @Tags         Analytics
// @Produce      json
// @Success      200  {array}   analytics.ExamsHistoryResponse  "Histórico analítico retornado com sucesso"
// @Failure      401  {object}  map[string]string                "Não autorizado / Token JWT ausente ou inválido"
// @Failure      500  {object}  map[string]string                "Erro interno do servidor"
// @Security     BearerAuth
// @Router       /analytics/exams-history [get]
func (h *AnalyticsHandler) ExamsHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		SendError(w, domain.ErrUnauthorized)
		return
	}

	res, err := h.analyticsUseCase.GetExamsHistory(r.Context(), userID)
	if err != nil {
		SendError(w, err)
		return
	}

	SendSuccess(w, http.StatusOK, res)
}

func parseDate(val string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", val)
}
