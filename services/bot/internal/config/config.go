// Package config provides configuration management for the application.
package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config represents the application configuration.
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
	// Telegram Bot Mode: "polling" or "webhook"
	TelegramMode string
	// Admin IDs for notifications
	AdminChatIDs   []int64  // IDs чатов администраторов для уведомлений
	AdminUsernames []string // Username'ы администраторов (читаются только из .env)
	// Matching Configuration
	PrimaryInterestScore    int // Баллы за совпадение основных интересов
	AdditionalInterestScore int // Баллы за совпадение дополнительных интересов
	MinCompatibilityScore   int // Минимальный балл совместимости
	MaxMatchesPerUser       int // Максимальное количество совпадений на пользователя
	// Interest Limits
	MinPrimaryInterests int     // Минимум основных интересов
	MaxPrimaryInterests int     // Максимум основных интересов
	PrimaryPercentage   float64 // Процент основных интересов от общего количества
}

// Load loads configuration from environment variables and .env file.
func Load() *Config {
	loadEnvFile()

	getFromFile := createFileReader()

	config := &Config{
		TelegramToken:           getTelegramToken(getFromFile),
		DatabaseURL:             getDatabaseURL(getFromFile),
		RedisURL:                getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword:           getEnv("REDIS_PASSWORD", ""),
		RedisDB:                 getRedisDB(),
		Port:                    getEnv("PORT", "8080"),
		Debug:                   getDebug(),
		WebhookURL:              getEnv("WEBHOOK_URL", ""),
		EnableTelegram:          getEnableTelegram(),
		EnableDiscord:           getEnableDiscord(),
		TelegramMode:            getTelegramMode(),
		AdminChatIDs:            parseAdminChatIDs(),
		AdminUsernames:          parseAdminUsernames(),
		PrimaryInterestScore:    getPrimaryInterestScore(),
		AdditionalInterestScore: getAdditionalInterestScore(),
		MinCompatibilityScore:   getMinCompatibilityScore(),
		MaxMatchesPerUser:       getMaxMatchesPerUser(),
		MinPrimaryInterests:     getMinPrimaryInterests(),
		MaxPrimaryInterests:     getMaxPrimaryInterests(),
		PrimaryPercentage:       getPrimaryPercentage(),
	}

	return config
}

// getTelegramToken получает токен Telegram из переменных окружения или файла.
func getTelegramToken(getFromFile func(string) string) string {
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		telegramToken = getFromFile(os.Getenv("TELEGRAM_TOKEN_FILE"))
	}

	return telegramToken
}

// getDatabaseURL получает URL базы данных из переменных окружения или файла.
func getDatabaseURL(getFromFile func(string) string) string {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = getFromFile(os.Getenv("DATABASE_URL_FILE"))
	}

	return databaseURL
}

// getRedisDB получает номер базы данных Redis.
func getRedisDB() int {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return redisDB
}

// getDebug получает флаг отладки.
func getDebug() bool {
	debug, _ := strconv.ParseBool(getEnv("DEBUG", "false"))

	return debug
}

// getEnableTelegram получает флаг включения Telegram.
func getEnableTelegram() bool {
	enableTelegram, _ := strconv.ParseBool(getEnv("ENABLE_TELEGRAM", "true"))

	return enableTelegram
}

// getEnableDiscord получает флаг включения Discord.
func getEnableDiscord() bool {
	enableDiscord, _ := strconv.ParseBool(getEnv("ENABLE_DISCORD", "false"))

	return enableDiscord
}

// getTelegramMode получает режим работы Telegram бота.
func getTelegramMode() string {
	mode := getEnv("TELEGRAM_MODE", "polling")
	mode = strings.ToLower(mode)

	// Валидация режима
	if mode != "polling" && mode != "webhook" {
		log.Printf("Warning: invalid TELEGRAM_MODE '%s', using 'polling' as default", mode)
		return "polling"
	}

	return mode
}

// parseAdminChatIDs парсит ID чатов администраторов.
func parseAdminChatIDs() []int64 {
	adminChatIDsStr := getEnv("ADMIN_CHAT_IDS", "")

	var adminChatIDs []int64

	if adminChatIDsStr != "" {
		for _, idStr := range strings.Split(adminChatIDsStr, ",") {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}

			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				adminChatIDs = append(adminChatIDs, id)
			} else {
				log.Printf("Ошибка парсинга admin chat ID '%s': %v", idStr, err)
			}
		}
	}

	return adminChatIDs
}

// parseAdminUsernames парсит имена пользователей администраторов.
func parseAdminUsernames() []string {
	adminUsernamesStr := getEnv("ADMIN_USERNAMES", "")

	var adminUsernames []string

	if adminUsernamesStr != "" {
		for _, username := range strings.Split(adminUsernamesStr, ",") {
			username = strings.TrimSpace(username)
			if username == "" {
				continue
			}

			username = strings.TrimPrefix(username, "@")
			adminUsernames = append(adminUsernames, username)
		}
	}

	return adminUsernames
}

// getPrimaryInterestScore получает баллы за основные интересы.
func getPrimaryInterestScore() int {
	score, _ := strconv.Atoi(getEnv("PRIMARY_INTEREST_SCORE", "3"))

	return score
}

// getAdditionalInterestScore получает баллы за дополнительные интересы.
func getAdditionalInterestScore() int {
	score, _ := strconv.Atoi(getEnv("ADDITIONAL_INTEREST_SCORE", "1"))

	return score
}

// getMinCompatibilityScore получает минимальный балл совместимости.
func getMinCompatibilityScore() int {
	score, _ := strconv.Atoi(getEnv("MIN_COMPATIBILITY_SCORE", "5"))

	return score
}

// getMaxMatchesPerUser получает максимальное количество совпадений на пользователя.
func getMaxMatchesPerUser() int {
	matches, _ := strconv.Atoi(getEnv("MAX_MATCHES_PER_USER", "10"))

	return matches
}

// getMinPrimaryInterests получает минимальное количество основных интересов.
func getMinPrimaryInterests() int {
	interests, _ := strconv.Atoi(getEnv("MIN_PRIMARY_INTERESTS", "1"))

	return interests
}

// getMaxPrimaryInterests получает максимальное количество основных интересов.
func getMaxPrimaryInterests() int {
	interests, _ := strconv.Atoi(getEnv("MAX_PRIMARY_INTERESTS", "5"))

	return interests
}

// getPrimaryPercentage получает процент основных интересов.
func getPrimaryPercentage() float64 {
	percentage, _ := strconv.ParseFloat(getEnv("PRIMARY_PERCENTAGE", "0.3"), 64)

	return percentage
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// loadEnvFile загружает .env файл из возможных путей.
func loadEnvFile() {
	envPaths := []string{
		"../../deploy/.env", // из services/bot/cmd/bot/
		"../deploy/.env",    // из services/bot/
		"deploy/.env",       // из корня проекта
		".env",              // текущая директория (fallback)
	}

	for _, path := range envPaths {
		err := godotenv.Load(path)
		if err == nil {
			log.Printf("Загружен .env файл из: %s", path)

			break
		}
	}
}

// createFileReader создает функцию для чтения файлов.
func createFileReader() func(string) string {
	return func(path string) string {
		if path == "" {
			return ""
		}

		// Очищаем путь для безопасности
		cleanPath := filepath.Clean(path)

		// Проверяем, что путь не содержит опасные символы
		if strings.Contains(cleanPath, "..") || strings.Contains(cleanPath, "~") {
			return ""
		}

		if b, err := os.ReadFile(cleanPath); err == nil {
			return strings.TrimSpace(string(b))
		}

		return ""
	}
}
