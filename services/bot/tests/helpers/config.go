package helpers

import (
	"language-exchange-bot/internal/config"
)

// GetTestConfig возвращает конфигурацию для тестов
func GetTestConfig() *config.Config {
	return &config.Config{
		// Тестовый токен - НЕ ИСПОЛЬЗУЙТЕ В PRODUCTION!
		TelegramToken:  "test_token_12345",
		DatabaseURL:    "postgres://test:test@localhost:5432/test_db",
		Debug:          true,
		EnableTelegram: true,
		EnableDiscord:  false,
		AdminChatIDs:   []int64{123456789, 987654321},
		AdminUsernames: []string{"testadmin", "testadmin2"},
	}
}

// GetTestAdminUserConfig возвращает данные тестового администратора
func GetTestAdminUserConfig() (int64, string) {
	return 123456789, "testadmin"
}

// GetTestRegularUserConfig возвращает данные тестового обычного пользователя
func GetTestRegularUserConfig() (int64, string) {
	return 555666777, "testuser"
}
