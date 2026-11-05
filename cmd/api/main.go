package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "mentor-api/cmd/api/docs"
	"mentor-api/internal/infrastructure/config"
	"mentor-api/internal/infrastructure/database"
	pgAnalytics "mentor-api/internal/infrastructure/repository/postgres/analytics"
	pgExam "mentor-api/internal/infrastructure/repository/postgres/exam"
	pgHistory "mentor-api/internal/infrastructure/repository/postgres/history"
	pgQuestion "mentor-api/internal/infrastructure/repository/postgres/question"
	pgUser "mentor-api/internal/infrastructure/repository/postgres/user"
	"mentor-api/internal/infrastructure/security"
	httpInterface "mentor-api/internal/interface/http"
	"mentor-api/internal/interface/http/handler"
	"mentor-api/internal/usecase/analytics"
	"mentor-api/internal/usecase/auth"
	"mentor-api/internal/usecase/exam"
	"mentor-api/internal/usecase/history"
	"mentor-api/internal/usecase/question"
)

// @title Mentor API
// @version 1.0
// @description API para a aplicação Mentor, simulador de provas e simulados padrão Enem. Rotas de autenticação, gerenciamento de simulados, histórico de simulados e análise de desempenho do usuário.
// @basePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Digite "Bearer " seguido do token JWT
func main() {
	cfg := config.Load()
	log.Println("Starting Mentor API...")

	// 1. Connect to PostgreSQL
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v. Running in offline/disconnected mode.", err)
	} else {
		defer db.Close()
		log.Println("Successfully connected to PostgreSQL.")

		// Run database migrations with version control
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("Database migration failed: %v", err)
		}
		log.Println("Database migrations applied successfully.")
	}

	// 2. Initialize Security Services
	hasher := security.NewBcryptHasher(cfg.BcryptCost)
	tokenManager := security.NewJWTManager(cfg.JWTSecret)

	// 3. Initialize Repositories
	userRepo := pgUser.NewUserRepository(db)
	questionRepo := pgQuestion.NewQuestionRepository(db)
	examRepo := pgExam.NewExamRepository(db, questionRepo)
	historyRepo := pgHistory.NewHistoryRepository(db)
	analyticsRepo := pgAnalytics.NewAnalyticsRepository(db)

	// 4. Initialize Use Cases & Internal Services
	questionService := question.NewQuestionService(questionRepo)
	authUseCase := auth.NewAuthUseCase(userRepo, hasher, tokenManager, cfg.JWTExpiresIn)
	examUseCase := exam.NewExamUseCase(examRepo, historyRepo, questionService)
	historyUseCase := history.NewHistoryUseCase(historyRepo, examRepo)
	analyticsUseCase := analytics.NewAnalyticsUseCase(analyticsRepo)

	// 5. Seed default questions internally if database is available
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		existing, _ := questionRepo.ListRandom(ctx, 1)
		if len(existing) == 0 {
			log.Println("Questions database is empty. Seeding default questions internally...")
			if err := questionService.SeedDefaultQuestions(ctx); err != nil {
				log.Printf("Warning: Failed to seed default questions: %v", err)
			} else {
				log.Println("Default questions seeded successfully.")
			}
		}
		cancel()
	}

	// 6. Initialize HTTP Handlers & Router
	authHandler := handler.NewAuthHandler(authUseCase)
	examHandler := handler.NewExamHandler(examUseCase)
	historyHandler := handler.NewHistoryHandler(historyUseCase)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsUseCase)

	router := httpInterface.NewRouter(httpInterface.RouterConfig{
		AuthHandler:      authHandler,
		ExamHandler:      examHandler,
		HistoryHandler:   historyHandler,
		AnalyticsHandler: analyticsHandler,
		TokenManager:     tokenManager,
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Start server in goroutine
	go func() {
		log.Printf("HTTP Server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 8. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped.")
}
