package config

import (
	"os"
	"testing"

	"language-exchange-bot/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig *config.Config
	}{
		{
			name:    "Default configuration",
			envVars: map[string]string{},
			expectedConfig: &config.Config{
				TelegramToken:  "",
				DatabaseURL:    "",
				RedisURL:       "",
				Port:           "8080",
				Debug:          false,
				WebhookURL:     "",
				EnableTelegram: true,
				EnableDiscord:  false,
				AdminChatIDs:   nil,
				AdminUsernames: nil,
			},
		},
		{
			name: "Custom configuration",
			envVars: map[string]string{
				"TELEGRAM_TOKEN":  "test_token",
				"DATABASE_URL":    "postgres://test",
				"REDIS_URL":       "redis://test",
				"PORT":            "9090",
				"DEBUG":           "true",
				"WEBHOOK_URL":     "https://test.com",
				"ENABLE_TELEGRAM": "true",
				"ENABLE_DISCORD":  "false",
				"ADMIN_CHAT_IDS":  "123,456",
				"ADMIN_USERNAMES": "admin1,admin2",
			},
			expectedConfig: &config.Config{
				TelegramToken:  "test_token",
				DatabaseURL:    "postgres://test",
				RedisURL:       "redis://test",
				Port:           "9090",
				Debug:          true,
				WebhookURL:     "https://test.com",
				EnableTelegram: true,
				EnableDiscord:  false,
				AdminChatIDs:   []int64{123, 456},
				AdminUsernames: []string{"admin1", "admin2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Act
			cfg := config.Load()

			// Assert
			assert.Equal(t, tt.expectedConfig.TelegramToken, cfg.TelegramToken)
			assert.Equal(t, tt.expectedConfig.DatabaseURL, cfg.DatabaseURL)
			assert.Equal(t, tt.expectedConfig.RedisURL, cfg.RedisURL)
			assert.Equal(t, tt.expectedConfig.Port, cfg.Port)
			assert.Equal(t, tt.expectedConfig.Debug, cfg.Debug)
			assert.Equal(t, tt.expectedConfig.WebhookURL, cfg.WebhookURL)
			assert.Equal(t, tt.expectedConfig.EnableTelegram, cfg.EnableTelegram)
			assert.Equal(t, tt.expectedConfig.EnableDiscord, cfg.EnableDiscord)
			assert.Equal(t, tt.expectedConfig.AdminChatIDs, cfg.AdminChatIDs)
			assert.Equal(t, tt.expectedConfig.AdminUsernames, cfg.AdminUsernames)
		})
	}
}

func TestLoadConfig_WithEnvFile(t *testing.T) {
	// This test would require creating a temporary .env file
	// For now, we'll just test that LoadConfig doesn't panic
	assert.NotPanics(t, func() {
		config.Load()
	})
}
