package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	DatabaseURL   string
	WebhookURL    string
	Port          string
	Debug         bool
}

func Load() *Config {
	// Загружаем .env файл (для разработки)
	godotenv.Load()

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	return &Config{
		TelegramToken: getEnv("TELEGRAM_TOKEN", ""),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost/languagebot?sslmode=disable"),
		WebhookURL:    getEnv("WEBHOOK_URL", ""), // Пустая строка = polling mode
		Port:          getEnv("PORT", "8080"),
		Debug:         debug,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
