package errors_test

import (
	"errors"
	"testing"
	"time"

	errorsPkg "language-exchange-bot/internal/errors"
)

// TestErrorHandlingExample демонстрирует использование новой системы ошибок.
func TestErrorHandlingExample(t *testing.T) {
	t.Parallel()

	// Создаем обработчик ошибок
	adminNotifier := errorsPkg.NewAdminNotifier([]int64{123456789}, nil)
	errorHandler := errorsPkg.NewErrorHandler(adminNotifier)

	// Симулируем ошибку Telegram API
	telegramErr := errorsPkg.ErrTelegramAPIRateLimit
	handledErr := errorHandler.HandleTelegramError(telegramErr, 67890, 12345, "SendMessage")

	// Проверяем, что ошибка обработана правильно
	if handledErr == nil {
		t.Fatal("Expected error, got nil")
	}

	// Проверяем тип ошибки
	if !errorsPkg.IsTelegramAPIError(handledErr) {
		t.Error("Expected TelegramAPI error type")
	}

	// Проверяем, что это CustomError
	customErr := &errorsPkg.CustomError{}
	if errors.As(handledErr, &customErr) {
		validateCustomError(t, customErr)
	} else {
		t.Error("Expected CustomError type")
	}

	t.Logf("Error handled successfully: %v", handledErr)
}

// validateCustomError проверяет поля CustomError.
func validateCustomError(t *testing.T, customErr *errorsPkg.CustomError) {
	t.Helper()

	if customErr.RequestID == "" {
		t.Error("Expected RequestID to be set")
	}

	if customErr.UserMessage == "" {
		t.Error("Expected UserMessage to be set")
	}

	if customErr.Context["user_id"] != int64(12345) {
		t.Error("Expected user_id in context")
	}

	if customErr.Context["chat_id"] != int64(67890) {
		t.Error("Expected chat_id in context")
	}

	if customErr.Context["operation"] != "SendMessage" {
		t.Error("Expected operation in context")
	}
}

// TestRequestContextGeneration тестирует генерацию RequestID.
func TestRequestContextGeneration(t *testing.T) {
	t.Parallel()

	ctx1 := errorsPkg.NewRequestContext(1, 2, "test1")

	time.Sleep(1 * time.Millisecond) // Небольшая задержка для гарантии разных timestamp

	ctx2 := errorsPkg.NewRequestContext(1, 2, "test2")

	if ctx1.RequestID == ctx2.RequestID {
		t.Error("Expected different RequestIDs")
	}

	if ctx1.RequestID == "" {
		t.Error("Expected non-empty RequestID")
	}

	t.Logf("RequestID 1: %s", ctx1.RequestID)
	t.Logf("RequestID 2: %s", ctx2.RequestID)
}

// TestErrorTypes тестирует типы ошибок.
func TestErrorTypes(t *testing.T) {
	t.Parallel()

	ctx := errorsPkg.NewRequestContext(123, 456, "TestOperation")

	// Тестируем разные типы ошибок
	telegramErr := errorsPkg.NewTelegramError("API error", "Проблема с API", ctx)
	databaseErr := errorsPkg.NewDatabaseError("DB error", "Проблема с БД", ctx)
	validationErr := errorsPkg.NewValidationError("Validation error", "Ошибка валидации", ctx)
	cacheErr := errorsPkg.NewCacheError("Cache error", "Проблема с кэшем", ctx)

	// Проверяем типы
	if !errorsPkg.IsTelegramAPIError(telegramErr) {
		t.Error("Expected TelegramAPI error")
	}

	if !errorsPkg.IsDatabaseError(databaseErr) {
		t.Error("Expected Database error")
	}

	if !errorsPkg.IsValidationError(validationErr) {
		t.Error("Expected Validation error")
	}

	if !errorsPkg.IsCacheError(cacheErr) {
		t.Error("Expected Cache error")
	}

	// Проверяем, что все ошибки имеют RequestID
	if telegramErr.RequestID == "" {
		t.Error("Expected RequestID in Telegram error")
	}

	if databaseErr.RequestID == "" {
		t.Error("Expected RequestID in Database error")
	}

	t.Logf("All error types work correctly")
}
