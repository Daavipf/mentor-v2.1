package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	JWTExpiresIn time.Duration
	BcryptCost   int
}

func Load() *Config {
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mentor?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "super-secret-jwt-key-mentor-v2")
	jwtExpiresStr := getEnv("JWT_EXPIRES_IN_HOURS", "24")
	bcryptCostStr := getEnv("BCRYPT_COST", "10")

	jwtExpiresHours, err := strconv.Atoi(jwtExpiresStr)
	if err != nil || jwtExpiresHours <= 0 {
		jwtExpiresHours = 24
	}

	bcryptCost, err := strconv.Atoi(bcryptCostStr)
	if err != nil || bcryptCost <= 0 {
		bcryptCost = 10
	}

	return &Config{
		Port:         port,
		DatabaseURL:  dbURL,
		JWTSecret:    jwtSecret,
		JWTExpiresIn: time.Duration(jwtExpiresHours) * time.Hour,
		BcryptCost:   bcryptCost,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

