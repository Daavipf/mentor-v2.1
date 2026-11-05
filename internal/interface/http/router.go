package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"mentor-api/internal/infrastructure/security"
	"mentor-api/internal/interface/http/handler"
	"mentor-api/internal/interface/http/middleware"
)

type RouterConfig struct {
	AuthHandler      *handler.AuthHandler
	ExamHandler      *handler.ExamHandler
	HistoryHandler   *handler.HistoryHandler
	AnalyticsHandler *handler.AnalyticsHandler
	TokenManager     security.TokenManager
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.Logger)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handler.SendSuccess(w, http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", cfg.AuthHandler.Register)
			r.Post("/login", cfg.AuthHandler.Login)
		})

		// Protected routes requiring valid JWT
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.TokenManager))

			// User profile
			r.Get("/auth/me", cfg.AuthHandler.Me)

			// Exams / Simulados (each user restricted to their own)
			r.Route("/exams", func(r chi.Router) {
				r.Post("/", cfg.ExamHandler.Create)
				r.Get("/", cfg.ExamHandler.List)
				r.Get("/{id}", cfg.ExamHandler.GetByID)
				r.Post("/{id}/answers", cfg.ExamHandler.AnswerQuestion)
				r.Post("/{id}/finish", cfg.ExamHandler.Finish)
				r.Post("/{id}/submit", cfg.ExamHandler.SubmitExam)
				r.Delete("/{id}", cfg.ExamHandler.Delete)
			})

			// Exam Histories (each user restricted to their own)
			r.Route("/histories", func(r chi.Router) {
				r.Get("/", cfg.HistoryHandler.List)
				r.Get("/{id}", cfg.HistoryHandler.GetByID)
			})

			// Analytics & Performance Metrics (each user restricted to their own)
			r.Route("/analytics", func(r chi.Router) {
				r.Get("/overview", cfg.AnalyticsHandler.Overview)
				r.Get("/by-discipline", cfg.AnalyticsHandler.ByDiscipline)
				r.Get("/by-area", cfg.AnalyticsHandler.ByArea)
				r.Get("/topics-ranking", cfg.AnalyticsHandler.TopicsRanking)
				r.Get("/evolution", cfg.AnalyticsHandler.Evolution)
				r.Get("/exams-history", cfg.AnalyticsHandler.ExamsHistory)
			})
		})
	})

	return r
}
