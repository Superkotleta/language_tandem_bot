// Package config provides configuration management for the application.
package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"language-exchange-bot/internal/localization"

	"github.com/joho/godotenv"
)

// Config represents the application configuration.
type Config struct {
	// Telegram Bot
	TelegramToken string
	// Database
	DatabaseURL          string
	DatabaseMaxOpenConns int // Максимум открытых соединений
	DatabaseMaxIdleConns int // Максимум idle соединений
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
		DatabaseMaxOpenConns:    getDatabaseMaxOpenConns(),
		DatabaseMaxIdleConns:    getDatabaseMaxIdleConns(),
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
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return 0
	}

	return redisDB
}

// getDebug получает флаг отладки.
func getDebug() bool {
	debug, err := strconv.ParseBool(getEnv("DEBUG", "false"))
	if err != nil {
		return false // default value
	}

	return debug
}

// getEnableTelegram получает флаг включения Telegram.
func getEnableTelegram() bool {
	enableTelegram, err := strconv.ParseBool(getEnv("ENABLE_TELEGRAM", "true"))
	if err != nil {
		return true // default value
	}

	return enableTelegram
}

// getEnableDiscord получает флаг включения Discord.
func getEnableDiscord() bool {
	enableDiscord, err := strconv.ParseBool(getEnv("ENABLE_DISCORD", "false"))
	if err != nil {
		return false // default value
	}

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

	// Возвращаем пустой slice вместо nil
	if adminChatIDs == nil {
		adminChatIDs = []int64{}
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
			if username == "" {
				continue
			}

			adminUsernames = append(adminUsernames, username)
		}
	}

	// Возвращаем пустой slice вместо nil
	if adminUsernames == nil {
		adminUsernames = []string{}
	}

	return adminUsernames
}

// getPrimaryInterestScore получает баллы за основные интересы.
func getPrimaryInterestScore() int {
	score, err := strconv.Atoi(getEnv("PRIMARY_INTEREST_SCORE", "3"))
	if err != nil {
		return localization.DefaultPrimaryInterestScore
	}

	return score
}

// getAdditionalInterestScore получает баллы за дополнительные интересы.
func getAdditionalInterestScore() int {
	score, err := strconv.Atoi(getEnv("ADDITIONAL_INTEREST_SCORE", "1"))
	if err != nil {
		return localization.DefaultAdditionalInterestScore
	}

	return score
}

// getMinCompatibilityScore получает минимальный балл совместимости.
func getMinCompatibilityScore() int {
	score, err := strconv.Atoi(getEnv("MIN_COMPATIBILITY_SCORE", "5"))
	if err != nil {
		return localization.DefaultMinCompatibilityScore
	}

	return score
}

// getMaxMatchesPerUser получает максимальное количество совпадений на пользователя.
func getMaxMatchesPerUser() int {
	matches, err := strconv.Atoi(getEnv("MAX_MATCHES_PER_USER", "10"))
	if err != nil {
		return localization.DefaultMaxMatchesPerUser
	}

	return matches
}

// getMinPrimaryInterests получает минимальное количество основных интересов.
func getMinPrimaryInterests() int {
	interests, err := strconv.Atoi(getEnv("MIN_PRIMARY_INTERESTS", "1"))
	if err != nil {
		return localization.DefaultMinPrimaryInterests
	}

	return interests
}

// getMaxPrimaryInterests получает максимальное количество основных интересов.
func getMaxPrimaryInterests() int {
	interests, err := strconv.Atoi(getEnv("MAX_PRIMARY_INTERESTS", "5"))
	if err != nil {
		return localization.DefaultMaxPrimaryInterests
	}

	return interests
}

// getPrimaryPercentage получает процент основных интересов.
func getPrimaryPercentage() float64 {
	percentage, err := strconv.ParseFloat(getEnv("PRIMARY_PERCENTAGE", "0.3"), 64)
	if err != nil {
		return localization.DefaultPrimaryPercentage
	}

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
	// Пропускаем загрузку .env файлов в тестах
	if strings.HasSuffix(os.Args[0], ".test") || os.Getenv("GO_TEST") == "1" {
		return
	}

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

		// В тестах разрешаем любые пути, включая временные директории
		if !strings.HasSuffix(os.Args[0], ".test") && os.Getenv("GO_TEST") != "1" {
			// Проверяем, что путь не содержит опасные символы
			if strings.Contains(cleanPath, "..") || strings.Contains(cleanPath, "~") {
				return ""
			}
		}

		if b, err := os.ReadFile(cleanPath); err == nil {
			return strings.TrimSpace(string(b))
		}

		return ""
	}
}

// getDatabaseMaxOpenConns получает максимальное количество открытых соединений.
func getDatabaseMaxOpenConns() int {
	value := getEnv("DATABASE_MAX_OPEN_CONNS", "25")
	if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
		return parsed
	}

	return localization.DefaultDatabaseMaxOpenConns
}

// getDatabaseMaxIdleConns получает максимальное количество idle соединений.
func getDatabaseMaxIdleConns() int {
	value := getEnv("DATABASE_MAX_IDLE_CONNS", "10")
	if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
		return parsed
	}

	return localization.DefaultDatabaseMaxIdleConns
}
