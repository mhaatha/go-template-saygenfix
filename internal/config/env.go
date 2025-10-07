package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	DBURL   string

	SessionName   string
	SessionMaxAge string

	GeminiAPIKey string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("../../.env")
	if err != nil {
		return &Config{}, err
	}

	return &Config{
		AppPort: os.Getenv("APP_PORT"),
		DBURL:   os.Getenv("DB_URL"),

		SessionName:   os.Getenv("SESSION_NAME"),
		SessionMaxAge: os.Getenv("SESSION_MAX_AGE"),

		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
	}, nil
}
