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

	// Парсим Admin Chat IDs - поддержка как числовых ID, так и username'ов (только для динамического чтения)
	adminChatIDsStr := getEnv("ADMIN_CHAT_IDS", "")
	var adminChatIDs []int64
	var adminUsernames []string

	if adminChatIDsStr != "" {
		for _, idStr := range strings.Split(adminChatIDsStr, ",") {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}

			// Если начинается с @, убираем @ для чистоты хранения
			if strings.HasPrefix(idStr, "@") {
				username := strings.TrimPrefix(idStr, "@")
				adminUsernames = append(adminUsernames, username)
			} else {
				// Парсим числовой ID
				if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
					adminChatIDs = append(adminChatIDs, id)
				} else {
					log.Printf("Ошибка парсинга admin chat ID '%s': %v", idStr, err)
				}
			}
		}
	}

	return &Config{
		TelegramToken:  telegramToken,
		DatabaseURL:    databaseURL,
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
