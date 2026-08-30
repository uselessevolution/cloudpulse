package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	AIProvider  string
	OpenAIModel string
}

func Load() (Config, error) {
	port := getEnv("PORT", "8080")

	databaseURL := os.Getenv("DATABASE_URL")
	aiProvider :=
		getEnv(
			"AI_PROVIDER",
			"mock",
		)

	openAIModel :=
		getEnv(
			"OPENAI_MODEL",
			"gpt-5.2",
		)
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
		AIProvider:  aiProvider,
		OpenAIModel: openAIModel,
	}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
