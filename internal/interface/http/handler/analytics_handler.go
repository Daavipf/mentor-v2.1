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

