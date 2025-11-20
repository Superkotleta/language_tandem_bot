package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL   string
	TelegramToken string
	LocalesPath   string
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

	localesPath := os.Getenv("LOCALES_PATH")
	if localesPath == "" {
		localesPath = "./locales" // Default value
	}

	return &Config{
		DatabaseURL:   dbURL,
		TelegramToken: tgToken,
		LocalesPath:   localesPath,
	}, nil
}
