package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisURL        string
	ClerkSecretKey  string
	ClerkIssuerURL  string
	OpenAIAPIKey    string
	Environment     string
	RateLimitRPM    int
}

func Load() (*Config, error) {
	// Load .env file if present (ignored in production)
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	databaseURL := getEnv("DATABASE_URL", "")
	redisURL := getEnv("REDIS_URL", "")
	clerkSecretKey := getEnv("CLERK_SECRET_KEY", "")
	clerkIssuerURL := getEnv("CLERK_ISSUER_URL", "https://clerk.ledgerly.com")
	openAIAPIKey := getEnv("OPENAI_API_KEY", "")
	environment := getEnv("ENVIRONMENT", "development")
	rateLimitRPM := getEnvInt("RATE_LIMIT_RPM", 60)

	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if clerkSecretKey == "" && environment == "production" {
		return nil, fmt.Errorf("CLERK_SECRET_KEY is required in production")
	}

	return &Config{
		Port:           port,
		DatabaseURL:    databaseURL,
		RedisURL:       redisURL,
		ClerkSecretKey: clerkSecretKey,
		ClerkIssuerURL: clerkIssuerURL,
		OpenAIAPIKey:   openAIAPIKey,
		Environment:    environment,
		RateLimitRPM:   rateLimitRPM,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		var result int
		_, err := fmt.Sscanf(value, "%d", &result)
		if err != nil {
			return fallback
		}
		return result
	}
	return fallback
}
