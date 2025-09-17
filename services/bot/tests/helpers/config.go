package helpers

import (
	"language-exchange-bot/internal/config"
)

// GetTestConfig возвращает конфигурацию для тестов
func GetTestConfig() *config.Config {
	return &config.Config{
		TelegramToken:  "test_token_12345",
		DatabaseURL:    "postgres://test:test@localhost:5432/test_db",
		Debug:          true,
		EnableTelegram: true,
		EnableDiscord:  false,
		AdminChatIDs:   []int64{123456789, 987654321},
		AdminUsernames: []string{"testadmin", "testadmin2"},
	}
}

// GetTestAdminUser возвращает данные тестового администратора
func GetTestAdminUser() (int64, string) {
	return 123456789, "testadmin"
}

// GetTestRegularUser возвращает данные тестового обычного пользователя
func GetTestRegularUser() (int64, string) {
	return 555666777, "testuser"
}
