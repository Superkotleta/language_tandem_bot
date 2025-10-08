package telegram

import (
	"errors"
	"testing"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"
	errorsPkg "language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

// TestTelegramHandler_HandleUpdate_Message tests message handling functionality.
// This smoke test verifies that the handler properly processes incoming messages
// and handles error cases when the service is not initialized.
func TestTelegramHandler_HandleUpdate_Message(t *testing.T) {
	// Создаем rate limiter для тестов
	rateLimiter := NewRateLimiter(DefaultRateLimitConfig())
	defer rateLimiter.Stop() // Останавливаем после теста

	// Создаем минимальный handler без service (для smoke теста)
	handler := &TelegramHandler{
		rateLimiter: rateLimiter,
	}

	// Создаем тестовое сообщение
	message := &tgbotapi.Message{
		MessageID: 1,
		From: &tgbotapi.User{
			ID:           123,
			UserName:     "testuser",
			FirstName:    "Test",
			LanguageCode: "en",
		},
		Chat: &tgbotapi.Chat{
			ID: 123,
		},
		Text: "Hello",
	}

	update := tgbotapi.Update{
		UpdateID: 1,
		Message:  message,
	}

	// Тестируем, что handler возвращает ошибку при отсутствии service
	err := handler.HandleUpdate(update)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not initialized")
}

// TestTelegramHandler_HandleUpdate_CallbackQuery - smoke тест обработки callback'ов
func TestTelegramHandler_HandleUpdate_CallbackQuery(t *testing.T) {
	// Создаем rate limiter для тестов
	rateLimiter := NewRateLimiter(DefaultRateLimitConfig())
	defer rateLimiter.Stop()

	// Создаем минимальный handler
	handler := &TelegramHandler{
		rateLimiter: rateLimiter,
	}

	// Создаем тестовый callback
	callback := &tgbotapi.CallbackQuery{
		ID: "test_callback",
		From: &tgbotapi.User{
			ID:           123,
			UserName:     "testuser",
			FirstName:    "Test",
			LanguageCode: "en",
		},
		Message: &tgbotapi.Message{
			MessageID: 1,
			Chat: &tgbotapi.Chat{
				ID: 123,
			},
		},
		Data: "test_data",
	}

	update := tgbotapi.Update{
		UpdateID:      1,
		CallbackQuery: callback,
	}

	// Тестируем, что handler возвращает ошибку при отсутствии service
	err := handler.HandleUpdate(update)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not initialized")
}

// TestTelegramHandler_isAdmin - тест проверки прав администратора
func TestTelegramHandler_isAdmin(t *testing.T) {
	tests := []struct {
		name       string
		adminIDs   []int64
		adminNames []string
		userID     int64
		username   string
		expected   bool
	}{
		{
			name:       "Admin by Chat ID",
			adminIDs:   []int64{123, 456},
			adminNames: []string{},
			userID:     123,
			username:   "user",
			expected:   true,
		},
		{
			name:       "Admin by Username",
			adminIDs:   []int64{},
			adminNames: []string{"admin", "moderator"},
			userID:     999,
			username:   "admin",
			expected:   true,
		},
		{
			name:       "Not Admin",
			adminIDs:   []int64{123},
			adminNames: []string{"admin"},
			userID:     999,
			username:   "user",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &TelegramHandler{
				adminChatIDs:   tt.adminIDs,
				adminUsernames: tt.adminNames,
			}

			result := handler.isAdmin(tt.userID, tt.username)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTelegramHandler_GetRateLimiterStats - тест получения статистики rate limiter
func TestTelegramHandler_GetRateLimiterStats(t *testing.T) {
	t.Run("With rate limiter", func(t *testing.T) {
		mockRateLimiter := &RateLimiter{}
		handler := &TelegramHandler{
			rateLimiter: mockRateLimiter,
		}

		stats := handler.GetRateLimiterStats()
		assert.NotNil(t, stats)
	})

	t.Run("Without rate limiter", func(t *testing.T) {
		handler := &TelegramHandler{
			rateLimiter: nil,
		}

		stats := handler.GetRateLimiterStats()
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "error")
		assert.Equal(t, "rate limiter not initialized", stats["error"])
	})
}

// TestTelegramHandler_Stop - тест остановки handler
func TestTelegramHandler_Stop(t *testing.T) {
	t.Run("With rate limiter", func(t *testing.T) {
		rateLimiter := NewRateLimiter(DefaultRateLimitConfig())
		handler := &TelegramHandler{
			rateLimiter: rateLimiter,
		}

		// Should not panic
		assert.NotPanics(t, func() {
			handler.Stop()
		})

		// Should not panic on second call
		assert.NotPanics(t, func() {
			handler.Stop()
		})
	})

	t.Run("Without rate limiter", func(t *testing.T) {
		handler := &TelegramHandler{
			rateLimiter: nil,
		}

		// Should not panic
		assert.NotPanics(t, func() {
			handler.Stop()
		})
	})
}

// TestTelegramHandler_GettersSetters - тест геттеров и сеттеров
func TestTelegramHandler_GettersSetters(t *testing.T) {
	handler := &TelegramHandler{}

	// Test service getter/setter
	mockService := &core.BotService{}
	handler.SetService(mockService)
	assert.Equal(t, mockService, handler.GetService())

	// Test bot API getter/setter
	mockBot := &tgbotapi.BotAPI{}
	handler.SetBotAPI(mockBot)
	assert.Equal(t, mockBot, handler.GetBotAPI())
}

// TestTelegramHandler_sendRateLimitMessage - тест отправки сообщения о rate limit
func TestTelegramHandler_sendRateLimitMessage(t *testing.T) {
	mockBot := &tgbotapi.BotAPI{} // Используем реальный BotAPI для простоты
	handler := &TelegramHandler{
		bot: mockBot,
	}

	// Should not panic (в реальном использовании отправит сообщение)
	assert.NotPanics(t, func() {
		handler.sendRateLimitMessage(123, errors.New("rate limit exceeded"))
	})
}

// TestNewTelegramHandler - тест конструктора (пропускаем для реальных объектов)
func TestNewTelegramHandler(t *testing.T) {
	t.Skip("Skipping test with real objects - requires full service initialization")

	// В будущем можно добавить тест с реальными объектами
	// bot := &tgbotapi.BotAPI{}
	// service := &core.BotService{}
	// errorHandler := &errorsPkg.ErrorHandler{}
	//
	// handler := NewTelegramHandler(
	// 	bot,
	// 	service,
	// 	[]int64{123},
	// 	errorHandler,
	// )
	//
	// assert.NotNil(t, handler)
}

// TestNewTelegramHandlerWithAdmins - тест конструктора с админами (пропускаем для реальных объектов)
func TestNewTelegramHandlerWithAdmins(t *testing.T) {
	t.Skip("Skipping test with real objects - requires full service initialization")

	// В будущем можно добавить тест с реальными объектами
	// bot := &tgbotapi.BotAPI{}
	// service := &core.BotService{}
	// errorHandler := &errorsPkg.ErrorHandler{}
	//
	// handler := NewTelegramHandlerWithAdmins(
	// 	bot,
	// 	service,
	// 	[]int64{123},
	// 	[]string{"admin"},
	// 	errorHandler,
	// )
	//
	// assert.NotNil(t, handler)
}

// TestTelegramHandler_HandleCommand_Unknown - smoke тест неизвестной команды
func TestTelegramHandler_HandleCommand_Unknown(t *testing.T) {
	// Создаем rate limiter для тестов
	rateLimiter := NewRateLimiter(DefaultRateLimitConfig())
	defer rateLimiter.Stop()

	// Создаем минимальный handler
	handler := &TelegramHandler{
		rateLimiter: rateLimiter,
	}

	user := &models.User{
		ID:                    1,
		TelegramID:            123,
		InterfaceLanguageCode: "en",
	}

	message := &tgbotapi.Message{
		MessageID: 1,
		From: &tgbotapi.User{
			ID: 123,
		},
		Chat: &tgbotapi.Chat{
			ID: 123,
		},
		Text: "/unknown_command",
		Entities: []tgbotapi.MessageEntity{
			{
				Type:   "bot_command",
				Offset: 0,
				Length: 16,
			},
		},
	}

	// Тестируем, что handler возвращает ошибку при отсутствии service
	err := handler.handleCommand(message, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not initialized")
}

// TestTelegramHandler_HandleState - smoke тест обработки состояний
func TestTelegramHandler_HandleState(t *testing.T) {
	// Создаем rate limiter для тестов
	rateLimiter := NewRateLimiter(DefaultRateLimitConfig())
	defer rateLimiter.Stop()

	// Создаем минимальный handler
	handler := &TelegramHandler{
		rateLimiter: rateLimiter,
	}

	user := &models.User{
		ID:                    1,
		TelegramID:            123,
		State:                 models.StateWaitingLanguage,
		InterfaceLanguageCode: "en",
	}

	message := &tgbotapi.Message{
		MessageID: 1,
		From: &tgbotapi.User{
			ID: 123,
		},
		Chat: &tgbotapi.Chat{
			ID: 123,
		},
		Text: "Test message",
	}

	// Тестируем, что handler возвращает ошибку при отсутствии service
	err := handler.handleState(message, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not initialized")
}

// TestNewTelegramBot_InvalidToken tests bot creation with invalid token
func TestNewTelegramBot_InvalidToken(t *testing.T) {
	db := &database.DB{} // mock DB
	adminChatIDs := []int64{123}

	// Test with invalid token
	bot, err := NewTelegramBot("invalid_token", db, false, adminChatIDs)
	assert.Error(t, err)
	assert.Nil(t, bot)
	assert.Contains(t, err.Error(), "failed to create telegram bot")
}

// TestNewTelegramBotWithUsernames_EmptyUsernames tests bot creation with empty usernames
func TestNewTelegramBotWithUsernames_EmptyUsernames(t *testing.T) {
	db := &database.DB{} // mock DB
	adminUsernames := []string{"", "   ", ""}

	// Test with empty usernames - should create bot but with empty admin list
	bot, err := NewTelegramBotWithUsernames("invalid_token", db, false, adminUsernames)
	assert.Error(t, err) // Will fail due to invalid token, but admin processing should work
	assert.Nil(t, bot)
}

// TestResolveUsernameToChatID tests username resolution
func TestResolveUsernameToChatID(t *testing.T) {
	bot := &TelegramBot{}

	// Test method always returns 0, nil for compatibility
	chatID, err := bot.ResolveUsernameToChatID("testuser")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), chatID)

	// Test with @ prefix
	chatID, err = bot.ResolveUsernameToChatID("@testuser")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), chatID)
}

// TestTelegramBot_GetService tests service getter
func TestTelegramBot_GetService(t *testing.T) {
	// Create bot with mock service
	service := &core.BotService{}
	bot := &TelegramBot{
		service: service,
	}

	// Test service getter
	assert.Equal(t, service, bot.GetService())
}

// TestTelegramBot_GetAdminCount tests admin count getter
func TestTelegramBot_GetAdminCount(t *testing.T) {
	bot := &TelegramBot{
		adminChatIDs: []int64{123, 456},
	}

	assert.Equal(t, 2, bot.GetAdminCount())
}

// TestTelegramBot_SetAdminChatIDs tests setting admin chat IDs
func TestTelegramBot_SetAdminChatIDs(t *testing.T) {
	bot := &TelegramBot{}
	newAdminChatIDs := []int64{789, 101112}

	bot.SetAdminChatIDs(newAdminChatIDs)
	assert.Equal(t, newAdminChatIDs, bot.adminChatIDs)
}

// TestTelegramBot_GetPlatformName tests platform name getter
func TestTelegramBot_GetPlatformName(t *testing.T) {
	bot := &TelegramBot{}

	assert.Equal(t, "telegram", bot.GetPlatformName())
}

// TestTelegramBot_GetBotAPI tests API getter
func TestTelegramBot_GetBotAPI(t *testing.T) {
	api := &tgbotapi.BotAPI{}
	bot := &TelegramBot{
		api: api,
	}

	assert.Equal(t, api, bot.GetBotAPI())
}

// TestTelegramBot_SetErrorHandler tests error handler setter
func TestTelegramBot_SetErrorHandler(t *testing.T) {
	bot := &TelegramBot{}
	errorHandler := &errorsPkg.ErrorHandler{}

	bot.SetErrorHandler(errorHandler)
	assert.Equal(t, errorHandler, bot.errorHandler)
}
