package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AzureConnectionString string
	FrontendUrl           string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("No .env file found!")
	}

	return Config{
		AzureConnectionString: getEnv("AZURE_CONNECTION_STRING", ""),
		FrontendUrl:           getEnv("FRONTEND_URL", ""),
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
