package integration

import (
	"context"
	"net/http"
	"time"

	"mentor-api/internal/infrastructure/security"
	httpInterface "mentor-api/internal/interface/http"
	"mentor-api/internal/interface/http/handler"
	"mentor-api/internal/usecase/analytics"
	"mentor-api/internal/usecase/auth"
	"mentor-api/internal/usecase/exam"
	"mentor-api/internal/usecase/history"
	"mentor-api/internal/usecase/question"
	"mentor-api/tests/mocks"
)

type TestApp struct {
	Router        http.Handler
	TokenManager  security.TokenManager
	UserRepo      *mocks.MockUserRepository
	QuestionRepo  *mocks.MockQuestionRepository
	ExamRepo      *mocks.MockExamRepository
	HistoryRepo   *mocks.MockHistoryRepository
	AnalyticsRepo *mocks.MockAnalyticsRepository
}

func SetupTestApp() *TestApp {
	jwtSecret := "test-jwt-secret-key-mentor-v2-integration"
	tokenMgr := security.NewJWTManager(jwtSecret)
	hasher := security.NewBcryptHasher(4)

	userRepo := mocks.NewMockUserRepository().(*mocks.MockUserRepository)
	questionRepo := mocks.NewMockQuestionRepository().(*mocks.MockQuestionRepository)
	examRepo := mocks.NewMockExamRepository(questionRepo).(*mocks.MockExamRepository)
	historyRepo := mocks.NewMockHistoryRepository().(*mocks.MockHistoryRepository)
	analyticsRepo := mocks.NewMockAnalyticsRepository(examRepo, historyRepo, questionRepo)

	questionService := question.NewQuestionService(questionRepo)
	_ = questionService.SeedDefaultQuestions(context.Background())

	authUC := auth.NewAuthUseCase(userRepo, hasher, tokenMgr, 24*time.Hour)
	examUC := exam.NewExamUseCase(examRepo, historyRepo, questionService)
	historyUC := history.NewHistoryUseCase(historyRepo, examRepo)
	analyticsUC := analytics.NewAnalyticsUseCase(analyticsRepo)

	authHandler := handler.NewAuthHandler(authUC)
	examHandler := handler.NewExamHandler(examUC)
	historyHandler := handler.NewHistoryHandler(historyUC)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsUC)

	router := httpInterface.NewRouter(httpInterface.RouterConfig{
		AuthHandler:      authHandler,
		ExamHandler:      examHandler,
		HistoryHandler:   historyHandler,
		AnalyticsHandler: analyticsHandler,
		TokenManager:     tokenMgr,
	})

	return &TestApp{
		Router:        router,
		TokenManager:  tokenMgr,
		UserRepo:      userRepo,
		QuestionRepo:  questionRepo,
		ExamRepo:      examRepo,
		HistoryRepo:   historyRepo,
		AnalyticsRepo: analyticsRepo,
	}
}
