package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AzureConnectionString string
	FrontendUrl           string
	SentryDSN             string
	SentryEnvironment     string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env file: %w", err)
	}

	azureConnectionString, err := requiredEnv("AZURE_CONNECTION_STRING")
	if err != nil {
		return Config{}, err
	}
	frontendURL, err := requiredEnv("FRONTEND_URL")
	if err != nil {
		return Config{}, err
	}

	return Config{
		AzureConnectionString: azureConnectionString,
		FrontendUrl:           frontendURL,
		SentryDSN:             os.Getenv("SENTRY_DSN"),
		SentryEnvironment:     os.Getenv("SENTRY_ENVIRONMENT"),
	}, nil
}


func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}

	return value, nil
}
