package config

import (
	"log"
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
	// Redis
	RedisURL      string
	RedisPassword string
	RedisDB       int
	// Server
	Port       string
	Debug      bool
	WebhookURL string
	// Bot Platform Settings
	EnableTelegram bool
	EnableDiscord  bool // Для будущего расширения
	// Admin IDs for notifications
	AdminChatIDs   []int64  // IDs чатов администраторов для уведомлений
	AdminUsernames []string // Username'ы администраторов (читаются только из .env)
}

func Load() *Config {
	// Ищем .env файл в папке deploy (относительно корня проекта)
	envPaths := []string{
		"../../deploy/.env", // из services/bot/cmd/bot/
		"../deploy/.env",    // из services/bot/
		"deploy/.env",       // из корня проекта
		".env",              // текущая директория (fallback)
	}

	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Загружен .env файл из: %s", path)
			break
		}
	}

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

	// Redis configuration
	redisURL := getEnv("REDIS_URL", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	debug, _ := strconv.ParseBool(getEnv("DEBUG", "false"))
	enableTelegram, _ := strconv.ParseBool(getEnv("ENABLE_TELEGRAM", "true"))
	enableDiscord, _ := strconv.ParseBool(getEnv("ENABLE_DISCORD", "false"))

	// Парсим Admin Chat IDs для уведомлений
	adminChatIDsStr := getEnv("ADMIN_CHAT_IDS", "")
	var adminChatIDs []int64

	if adminChatIDsStr != "" {
		for _, idStr := range strings.Split(adminChatIDsStr, ",") {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}

			// Парсим числовой ID
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				adminChatIDs = append(adminChatIDs, id)
			} else {
				log.Printf("Ошибка парсинга admin chat ID '%s': %v", idStr, err)
			}
		}
	}

	// Парсим Admin Usernames для проверки прав
	adminUsernamesStr := getEnv("ADMIN_USERNAMES", "")
	var adminUsernames []string

	if adminUsernamesStr != "" {
		for _, username := range strings.Split(adminUsernamesStr, ",") {
			username = strings.TrimSpace(username)
			if username == "" {
				continue
			}

			// Убираем @ если есть
			if strings.HasPrefix(username, "@") {
				username = strings.TrimPrefix(username, "@")
			}
			adminUsernames = append(adminUsernames, username)
		}
	}

	return &Config{
		TelegramToken:  telegramToken,
		DatabaseURL:    databaseURL,
		RedisURL:       redisURL,
		RedisPassword:  redisPassword,
		RedisDB:        redisDB,
		Port:           getEnv("PORT", "8080"),
		Debug:          debug,
		WebhookURL:     getEnv("WEBHOOK_URL", ""),
		EnableTelegram: enableTelegram,
		EnableDiscord:  enableDiscord,
		AdminChatIDs:   adminChatIDs,
		AdminUsernames: adminUsernames,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
