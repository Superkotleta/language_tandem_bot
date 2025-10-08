package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig_Load_DefaultValues tests configuration loading with default values.
// This test ensures that when no environment variables are set,
// the configuration loads with sensible default values for all fields.
func TestConfig_Load_DefaultValues(t *testing.T) {
	// Clear environment before test to ensure clean state
	clearEnvironment()

	config := Load()

	// Проверяем дефолтные значения
	assert.Equal(t, "", config.TelegramToken)
	assert.Equal(t, "", config.DatabaseURL)
	assert.Equal(t, "localhost:6379", config.RedisURL)
	assert.Equal(t, "", config.RedisPassword)
	assert.Equal(t, 0, config.RedisDB)
	assert.Equal(t, "8080", config.Port)
	assert.False(t, config.Debug)
	assert.Equal(t, "", config.WebhookURL)
	assert.True(t, config.EnableTelegram)
	assert.False(t, config.EnableDiscord)
	assert.Equal(t, "polling", config.TelegramMode)
	assert.Empty(t, config.AdminChatIDs)
	assert.Empty(t, config.AdminUsernames)
	assert.Equal(t, 3, config.PrimaryInterestScore)
	assert.Equal(t, 1, config.AdditionalInterestScore)
	assert.Equal(t, 5, config.MinCompatibilityScore)
	assert.Equal(t, 10, config.MaxMatchesPerUser)
	assert.Equal(t, 1, config.MinPrimaryInterests)
	assert.Equal(t, 5, config.MaxPrimaryInterests)
	assert.Equal(t, 0.3, config.PrimaryPercentage)
}

// TestConfig_Load_FromEnvironment тестирует загрузку конфигурации из environment variables.
func TestConfig_Load_FromEnvironment(t *testing.T) {

	// Очищаем environment перед тестом
	clearEnvironment()

	// Устанавливаем тестовые значения
	envValues := map[string]string{
		"TELEGRAM_TOKEN":            "test_token_123",
		"DATABASE_URL":              "postgres://user:pass@localhost/db",
		"REDIS_URL":                 "redis.example.com:6379",
		"REDIS_PASSWORD":            "redis_pass",
		"REDIS_DB":                  "5",
		"PORT":                      "9090",
		"DEBUG":                     "true",
		"WEBHOOK_URL":               "https://example.com/webhook",
		"ENABLE_TELEGRAM":           "true",
		"ENABLE_DISCORD":            "false",
		"TELEGRAM_MODE":             "webhook",
		"ADMIN_CHAT_IDS":            "123456789,987654321",
		"ADMIN_USERNAMES":           "@admin1,@admin2",
		"PRIMARY_INTEREST_SCORE":    "5",
		"ADDITIONAL_INTEREST_SCORE": "2",
		"MIN_COMPATIBILITY_SCORE":   "8",
		"MAX_MATCHES_PER_USER":      "15",
		"MIN_PRIMARY_INTERESTS":     "2",
		"MAX_PRIMARY_INTERESTS":     "8",
		"PRIMARY_PERCENTAGE":        "0.4",
	}

	for key, value := range envValues {
		if err := os.Setenv(key, value); err != nil {
			t.Logf("Failed to set environment variable %s: %v", key, err)
		}
	}
	defer clearEnvironment()

	config := Load()

	// Проверяем загруженные значения
	assert.Equal(t, "test_token_123", config.TelegramToken)
	assert.Equal(t, "postgres://user:pass@localhost/db", config.DatabaseURL)
	assert.Equal(t, "redis.example.com:6379", config.RedisURL)
	assert.Equal(t, "redis_pass", config.RedisPassword)
	assert.Equal(t, 5, config.RedisDB)
	assert.Equal(t, "9090", config.Port)
	assert.True(t, config.Debug)
	assert.Equal(t, "https://example.com/webhook", config.WebhookURL)
	assert.True(t, config.EnableTelegram)
	assert.False(t, config.EnableDiscord)
	assert.Equal(t, "webhook", config.TelegramMode)
	assert.Equal(t, []int64{123456789, 987654321}, config.AdminChatIDs)
	assert.Equal(t, []string{"admin1", "admin2"}, config.AdminUsernames)
	assert.Equal(t, 5, config.PrimaryInterestScore)
	assert.Equal(t, 2, config.AdditionalInterestScore)
	assert.Equal(t, 8, config.MinCompatibilityScore)
	assert.Equal(t, 15, config.MaxMatchesPerUser)
	assert.Equal(t, 2, config.MinPrimaryInterests)
	assert.Equal(t, 8, config.MaxPrimaryInterests)
	assert.Equal(t, 0.4, config.PrimaryPercentage)
}

// TestConfig_Load_InvalidValues тестирует загрузку конфигурации с невалидными значениями.
func TestConfig_Load_InvalidValues(t *testing.T) {

	// Очищаем environment перед тестом
	clearEnvironment()

	// Устанавливаем невалидные значения
	envValues := map[string]string{
		"REDIS_DB":               "invalid",
		"DEBUG":                  "invalid",
		"ENABLE_TELEGRAM":        "invalid",
		"ENABLE_DISCORD":         "invalid",
		"TELEGRAM_MODE":          "invalid_mode",
		"ADMIN_CHAT_IDS":         "invalid1,invalid2",
		"PRIMARY_INTEREST_SCORE": "invalid",
		"PRIMARY_PERCENTAGE":     "invalid",
	}

	for key, value := range envValues {
		if err := os.Setenv(key, value); err != nil {
			t.Logf("Failed to set environment variable %s: %v", key, err)
		}
	}
	defer clearEnvironment()

	config := Load()

	// Проверяем, что невалидные значения игнорируются и используются дефолты
	assert.Equal(t, 0, config.RedisDB)              // default 0
	assert.False(t, config.Debug)                   // default false
	assert.True(t, config.EnableTelegram)           // default true
	assert.False(t, config.EnableDiscord)           // default false
	assert.Equal(t, "polling", config.TelegramMode) // default polling
	assert.Empty(t, config.AdminChatIDs)            // empty slice
	assert.Equal(t, 3, config.PrimaryInterestScore) // default 3
	assert.Equal(t, 0.3, config.PrimaryPercentage)  // default 0.3
}

// TestConfig_Load_AdminChatIDs тестирует парсинг admin chat IDs.
func TestConfig_Load_AdminChatIDs(t *testing.T) {

	testCases := []struct {
		input    string
		expected []int64
		name     string
	}{
		{"123456789,987654321", []int64{123456789, 987654321}, "valid IDs"},
		{"123, 456, 789", []int64{123, 456, 789}, "IDs with spaces"},
		{"123,,456", []int64{123, 456}, "empty values"},
		{"invalid,123", []int64{123}, "mixed valid/invalid"},
		{"", []int64{}, "empty string"},
		{",,,", []int64{}, "only commas"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clearEnvironment()
			if err := os.Setenv("ADMIN_CHAT_IDS", tc.input); err != nil {
				t.Logf("Failed to set ADMIN_CHAT_IDS: %v", err)
			}
			defer clearEnvironment()

			config := Load()
			assert.Equal(t, tc.expected, config.AdminChatIDs)
		})
	}
}

// TestConfig_Load_AdminUsernames тестирует парсинг admin usernames.
func TestConfig_Load_AdminUsernames(t *testing.T) {

	testCases := []struct {
		input    string
		expected []string
		name     string
	}{
		{"@user1,@user2", []string{"user1", "user2"}, "with @ prefix"},
		{"user1,user2", []string{"user1", "user2"}, "without @ prefix"},
		{"@user1, user2 , @user3", []string{"user1", "user2", "user3"}, "mixed with spaces"},
		{"user1,,user2", []string{"user1", "user2"}, "empty values"},
		{"", []string{}, "empty string"},
		{"@,@", []string{}, "only @ symbols"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clearEnvironment()
			if err := os.Setenv("ADMIN_USERNAMES", tc.input); err != nil {
				t.Logf("Failed to set ADMIN_USERNAMES: %v", err)
			}
			defer clearEnvironment()

			config := Load()
			assert.Equal(t, tc.expected, config.AdminUsernames)
		})
	}
}

// TestConfig_Load_TelegramMode тестирует валидацию режима Telegram.
func TestConfig_Load_TelegramMode(t *testing.T) {

	testCases := []struct {
		input    string
		expected string
		name     string
	}{
		{"polling", "polling", "valid polling"},
		{"webhook", "webhook", "valid webhook"},
		{"POLLING", "polling", "uppercase polling"},
		{"WEBHOOK", "webhook", "uppercase webhook"},
		{"invalid", "polling", "invalid mode defaults to polling"},
		{"", "polling", "empty defaults to polling"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clearEnvironment()
			if err := os.Setenv("TELEGRAM_MODE", tc.input); err != nil {
				t.Logf("Failed to set TELEGRAM_MODE: %v", err)
			}
			defer clearEnvironment()

			config := Load()
			assert.Equal(t, tc.expected, config.TelegramMode)
		})
	}
}

// TestConfig_Load_TokenFromFile тестирует загрузку токена из файла.
func TestConfig_Load_TokenFromFile(t *testing.T) {

	// Очищаем environment перед тестом
	clearEnvironment()

	// Убеждаемся что переменные пустые
	if err := os.Unsetenv("TELEGRAM_TOKEN"); err != nil {
		t.Logf("Failed to unset TELEGRAM_TOKEN: %v", err)
	}
	if err := os.Unsetenv("TELEGRAM_TOKEN_FILE"); err != nil {
		t.Logf("Failed to unset TELEGRAM_TOKEN_FILE: %v", err)
	}

	// Устанавливаем флаг теста для отключения проверки безопасности файлов
	if err := os.Setenv("GO_TEST", "1"); err != nil {
		t.Logf("Failed to set GO_TEST: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("GO_TEST"); err != nil {
			t.Logf("Failed to unset GO_TEST: %v", err)
		}
	}()

	// Создаем временный файл с токеном
	tempDir := t.TempDir()
	tokenFile := filepath.Join(tempDir, "telegram_token.txt")
	tokenContent := "file_token_123"

	err := os.WriteFile(tokenFile, []byte(tokenContent), 0600)
	require.NoError(t, err)

	// Устанавливаем переменную окружения с путем к файлу
	if err := os.Setenv("TELEGRAM_TOKEN_FILE", tokenFile); err != nil {
		t.Logf("Failed to set TELEGRAM_TOKEN_FILE: %v", err)
	}
	defer clearEnvironment()

	config := Load()

	assert.Equal(t, tokenContent, config.TelegramToken)
}

// TestConfig_Load_DatabaseURLFromFile тестирует загрузку database URL из файла.
func TestConfig_Load_DatabaseURLFromFile(t *testing.T) {

	// Очищаем environment перед тестом
	clearEnvironment()

	// Убеждаемся что переменные пустые
	if err := os.Unsetenv("DATABASE_URL"); err != nil {
		t.Logf("Failed to unset DATABASE_URL: %v", err)
	}
	if err := os.Unsetenv("DATABASE_URL_FILE"); err != nil {
		t.Logf("Failed to unset DATABASE_URL_FILE: %v", err)
	}

	// Устанавливаем флаг теста для отключения проверки безопасности файлов
	if err := os.Setenv("GO_TEST", "1"); err != nil {
		t.Logf("Failed to set GO_TEST: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("GO_TEST"); err != nil {
			t.Logf("Failed to unset GO_TEST: %v", err)
		}
	}()

	// Создаем временный файл с URL базы данных
	tempDir := t.TempDir()
	dbFile := filepath.Join(tempDir, "database_url.txt")
	dbContent := "postgres://user:pass@host:5432/db"

	err := os.WriteFile(dbFile, []byte(dbContent), 0600)
	require.NoError(t, err)

	// Устанавливаем переменную окружения с путем к файлу
	if err := os.Setenv("DATABASE_URL_FILE", dbFile); err != nil {
		t.Logf("Failed to set DATABASE_URL_FILE: %v", err)
	}
	defer clearEnvironment()

	config := Load()

	assert.Equal(t, dbContent, config.DatabaseURL)
}

// TestConfig_Load_FileSecurity тестирует безопасность загрузки файлов.
func TestConfig_Load_FileSecurity(t *testing.T) {

	// Очищаем environment перед тестом
	clearEnvironment()

	// Проверяем, что опасные пути отклоняются
	dangerousPaths := []string{
		"../../../etc/passwd",
		"~/secret.txt",
		"../../config/../../../root",
	}

	for _, dangerousPath := range dangerousPaths {
		t.Run("dangerous_"+dangerousPath, func(t *testing.T) {
			os.Setenv("TELEGRAM_TOKEN_FILE", dangerousPath)
			defer clearEnvironment()

			config := Load()
			// Путь должен быть отклонен, токен останется пустым
			assert.Empty(t, config.TelegramToken)
		})
	}
}

// TestConfig_Load_NumericParsing тестирует парсинг числовых значений.
func TestConfig_Load_NumericParsing(t *testing.T) {

	testCases := []struct {
		envKey   string
		envValue string
		expected int
		field    string
	}{
		{"REDIS_DB", "0", 0, "RedisDB"},
		{"REDIS_DB", "5", 5, "RedisDB"},
		{"PRIMARY_INTEREST_SCORE", "10", 10, "PrimaryInterestScore"},
		{"MIN_COMPATIBILITY_SCORE", "3", 3, "MinCompatibilityScore"},
		{"MAX_MATCHES_PER_USER", "20", 20, "MaxMatchesPerUser"},
		{"MIN_PRIMARY_INTERESTS", "0", 0, "MinPrimaryInterests"},
		{"MAX_PRIMARY_INTERESTS", "10", 10, "MaxPrimaryInterests"},
	}

	for _, tc := range testCases {
		t.Run(tc.envKey+"_"+tc.envValue, func(t *testing.T) {
			clearEnvironment()
			os.Setenv(tc.envKey, tc.envValue)
			defer clearEnvironment()

			config := Load()

			switch tc.field {
			case "RedisDB":
				assert.Equal(t, tc.expected, config.RedisDB)
			case "PrimaryInterestScore":
				assert.Equal(t, tc.expected, config.PrimaryInterestScore)
			case "MinCompatibilityScore":
				assert.Equal(t, tc.expected, config.MinCompatibilityScore)
			case "MaxMatchesPerUser":
				assert.Equal(t, tc.expected, config.MaxMatchesPerUser)
			case "MinPrimaryInterests":
				assert.Equal(t, tc.expected, config.MinPrimaryInterests)
			case "MaxPrimaryInterests":
				assert.Equal(t, tc.expected, config.MaxPrimaryInterests)
			}
		})
	}
}

// TestConfig_Load_BooleanParsing тестирует парсинг булевых значений.
func TestConfig_Load_BooleanParsing(t *testing.T) {

	testCases := []struct {
		envKey   string
		envValue string
		expected bool
		field    string
	}{
		{"DEBUG", "true", true, "Debug"},
		{"DEBUG", "false", false, "Debug"},
		{"DEBUG", "1", true, "Debug"},
		{"DEBUG", "0", false, "Debug"},
		{"ENABLE_TELEGRAM", "true", true, "EnableTelegram"},
		{"ENABLE_DISCORD", "true", true, "EnableDiscord"},
	}

	for _, tc := range testCases {
		t.Run(tc.envKey+"_"+tc.envValue, func(t *testing.T) {
			clearEnvironment()
			os.Setenv(tc.envKey, tc.envValue)
			defer clearEnvironment()

			config := Load()

			switch tc.field {
			case "Debug":
				assert.Equal(t, tc.expected, config.Debug)
			case "EnableTelegram":
				assert.Equal(t, tc.expected, config.EnableTelegram)
			case "EnableDiscord":
				assert.Equal(t, tc.expected, config.EnableDiscord)
			}
		})
	}
}

// TestConfig_Load_FloatParsing тестирует парсинг float значений.
func TestConfig_Load_FloatParsing(t *testing.T) {

	testCases := []struct {
		envValue string
		expected float64
	}{
		{"0.0", 0.0},
		{"0.3", 0.3},
		{"0.5", 0.5},
		{"1.0", 1.0},
		{"0.25", 0.25},
		{"invalid", 0.3}, // default value
		{"", 0.3},        // default value
	}

	for _, tc := range testCases {
		t.Run("PRIMARY_PERCENTAGE_"+tc.envValue, func(t *testing.T) {
			clearEnvironment()
			if tc.envValue != "" {
				os.Setenv("PRIMARY_PERCENTAGE", tc.envValue)
			}
			defer clearEnvironment()

			config := Load()
			assert.Equal(t, tc.expected, config.PrimaryPercentage)
		})
	}
}

// clearEnvironment очищает environment variables для тестов.
func clearEnvironment() {
	envKeys := []string{
		"TELEGRAM_TOKEN",
		"TELEGRAM_TOKEN_FILE",
		"DATABASE_URL",
		"DATABASE_URL_FILE",
		"REDIS_URL",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"PORT",
		"DEBUG",
		"WEBHOOK_URL",
		"ENABLE_TELEGRAM",
		"ENABLE_DISCORD",
		"TELEGRAM_MODE",
		"ADMIN_CHAT_IDS",
		"ADMIN_USERNAMES",
		"PRIMARY_INTEREST_SCORE",
		"ADDITIONAL_INTEREST_SCORE",
		"MIN_COMPATIBILITY_SCORE",
		"MAX_MATCHES_PER_USER",
		"MIN_PRIMARY_INTERESTS",
		"MAX_PRIMARY_INTERESTS",
		"PRIMARY_PERCENTAGE",
	}

	for _, key := range envKeys {
		os.Unsetenv(key)
	}
}
