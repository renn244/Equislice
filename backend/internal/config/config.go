package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AzureConnectionString string
	FrontendUrl           string
	SentryDSN             string
	SentryEnvironment     string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to load .env file: %v", err)
	}

	return Config{
		AzureConnectionString: getEnv("AZURE_CONNECTION_STRING", ""),
		FrontendUrl:           getEnv("FRONTEND_URL", ""),
		SentryDSN:             getEnv("SENTRY_DSN", ""),
		SentryEnvironment:     getEnv("SENTRY_ENVIRONMENT", ""),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" && fallback == "" {
		log.Fatalf("%T does not exist and there is no fallback", key)
		return ""
	}

	if value == "" {
		return fallback
	}

	return value
}
