package errors

import (
	"fmt"
	"testing"
)

// TestErrorHandlingIntegration тестирует интеграцию всей системы обработки ошибок
func TestErrorHandlingIntegration(t *testing.T) {
	// Создаем полную систему обработки ошибок
	adminNotifier := NewAdminNotifier([]int64{123456789, 987654321}, nil)
	errorHandler := NewErrorHandler(adminNotifier)

	// Тестируем разные сценарии ошибок
	t.Run("TelegramAPI Error", func(t *testing.T) {
		telegramErr := fmt.Errorf("telegram API rate limit exceeded")
		handledErr := errorHandler.HandleTelegramError(telegramErr, 12345, 67890, "SendMessage")

		if handledErr == nil {
			t.Fatal("Expected error, got nil")
		}

		if !IsTelegramAPIError(handledErr) {
			t.Error("Expected TelegramAPI error type")
		}

		t.Logf("TelegramAPI error handled: %v", handledErr)
	})

	t.Run("Database Error", func(t *testing.T) {
		dbErr := fmt.Errorf("database connection failed")
		handledErr := errorHandler.HandleDatabaseError(dbErr, 12345, 67890, "GetUser")

		if handledErr == nil {
			t.Fatal("Expected error, got nil")
		}

		if !IsDatabaseError(handledErr) {
			t.Error("Expected Database error type")
		}

		t.Logf("Database error handled: %v", handledErr)
	})

	t.Run("Validation Error", func(t *testing.T) {
		validationErr := fmt.Errorf("invalid user input")
		handledErr := errorHandler.HandleValidationError(validationErr, 12345, 67890, "ValidateInput")

		if handledErr == nil {
			t.Fatal("Expected error, got nil")
		}

		if !IsValidationError(handledErr) {
			t.Error("Expected Validation error type")
		}

		t.Logf("Validation error handled: %v", handledErr)
	})

	t.Run("Cache Error", func(t *testing.T) {
		cacheErr := fmt.Errorf("redis connection failed")
		handledErr := errorHandler.HandleCacheError(cacheErr, 12345, 67890, "GetCachedData")

		if handledErr == nil {
			t.Fatal("Expected error, got nil")
		}

		if !IsCacheError(handledErr) {
			t.Error("Expected Cache error type")
		}

		t.Logf("Cache error handled: %v", handledErr)
	})

	t.Log("All error handling scenarios completed successfully")
}

// TestRequestContextIntegration тестирует интеграцию RequestContext
func TestRequestContextIntegration(t *testing.T) {
	// Создаем контекст запроса
	ctx := NewRequestContext(12345, 67890, "TestOperation")

	// Проверяем, что RequestID генерируется
	if ctx.RequestID == "" {
		t.Error("Expected non-empty RequestID")
	}

	// Проверяем, что контекст содержит правильные данные
	if ctx.UserID != 12345 {
		t.Error("Expected UserID to be 12345")
	}

	if ctx.ChatID != 67890 {
		t.Error("Expected ChatID to be 67890")
	}

	if ctx.Operation != "TestOperation" {
		t.Error("Expected Operation to be 'TestOperation'")
	}

	t.Logf("RequestContext created successfully: %+v", ctx)
}

// TestAdminNotifierIntegration тестирует интеграцию уведомлений администраторов
func TestAdminNotifierIntegration(t *testing.T) {
	adminNotifier := NewAdminNotifier([]int64{123456789, 987654321}, nil)

	// Проверяем, что список администраторов установлен
	chatIDs := adminNotifier.GetAdminChatIDs()
	if len(chatIDs) != 2 {
		t.Errorf("Expected 2 admin chat IDs, got %d", len(chatIDs))
	}

	// Проверяем, что Chat ID правильные
	if chatIDs[0] != 123456789 {
		t.Error("Expected first admin chat ID to be 123456789")
	}

	if chatIDs[1] != 987654321 {
		t.Error("Expected second admin chat ID to be 987654321")
	}

	t.Logf("AdminNotifier configured with %d admin chat IDs", len(chatIDs))
}
