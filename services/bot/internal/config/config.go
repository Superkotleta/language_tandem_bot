package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Telegram Bot
	TelegramToken string
	// Database
	DatabaseURL string
	// Server
	Port       string
	Debug      bool
	WebhookURL string
	// Bot Platform Settings
	EnableTelegram bool
	EnableDiscord  bool // Для будущего расширения
}

func Load() *Config {
	_ = godotenv.Load()

	getFromFile := func(path string) string {
		if path == "" {
			return ""
		}
		if b, err := os.ReadFile(path); err == nil {
			return strings.TrimSpace(string(b))
		}
		return ""
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		telegramToken = getFromFile(os.Getenv("TELEGRAM_TOKEN_FILE"))
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = getFromFile(os.Getenv("DATABASE_URL_FILE"))
	}

	debug, _ := strconv.ParseBool(getEnv("DEBUG", "false"))
	enableTelegram, _ := strconv.ParseBool(getEnv("ENABLE_TELEGRAM", "true"))
	enableDiscord, _ := strconv.ParseBool(getEnv("ENABLE_DISCORD", "false"))

	return &Config{
		TelegramToken:  telegramToken,
		DatabaseURL:    databaseURL,
		Port:           getEnv("PORT", "8080"),
		Debug:          debug,
		WebhookURL:     getEnv("WEBHOOK_URL", ""),
		EnableTelegram: enableTelegram,
		EnableDiscord:  enableDiscord,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
