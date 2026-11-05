package main

import (
	"context"
	"log"
	"time"

	"mentor-api/internal/infrastructure/config"
	"mentor-api/internal/infrastructure/database"
	pgQuestion "mentor-api/internal/infrastructure/repository/postgres/question"
	"mentor-api/internal/usecase/question"
)

func main() {
	cfg := config.Load()
	log.Println("Starting internal question seeder...")

	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	questionRepo := pgQuestion.NewQuestionRepository(db)
	questionService := question.NewQuestionService(questionRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := questionService.SeedDefaultQuestions(ctx, cfg); err != nil {
		log.Fatalf("Failed to seed questions: %v", err)
	}

	log.Println("Questions and alternatives seeded successfully inside the system.")
}
