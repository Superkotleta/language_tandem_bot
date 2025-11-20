package config

import (
	"os"
	"fmt"
)

type Config struct {
	DatabaseURL   string
	TelegramToken string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	tgToken := os.Getenv("TELEGRAM_TOKEN")
	if tgToken == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	return &Config{
		DatabaseURL:   dbURL,
		TelegramToken: tgToken,
	}, nil
}


