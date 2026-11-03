package main

import (
	"log"

	"mentor-api/internal/infrastructure/config"
	"mentor-api/internal/infrastructure/database"
)

func main() {
	cfg := config.Load()
	log.Println("[Migrator CLI] Connecting to database for migrations...")

	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[Migrator CLI] Database connection failed: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("[Migrator CLI] Migration execution failed: %v", err)
	}

	log.Println("[Migrator CLI] All database migrations applied successfully.")
}

