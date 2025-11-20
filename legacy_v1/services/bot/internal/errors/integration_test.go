package errors_test

import (
	"testing"

	errorsPkg "language-exchange-bot/internal/errors"
)

// TestErrorHandlingTelegramAPI тестирует обработку ошибок Telegram API.
func TestErrorHandlingTelegramAPI(t *testing.T) {
	t.Parallel()

	adminNotifier := errorsPkg.NewAdminNotifier([]int64{123456789, 987654321}, nil)
	errorHandler := errorsPkg.NewErrorHandler(adminNotifier)

	telegramErr := errorsPkg.ErrTelegramAPIRateLimit

	handledErr := errorHandler.HandleTelegramError(telegramErr, 12345, 67890, "SendMessage")
	if handledErr == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errorsPkg.IsTelegramAPIError(handledErr) {
		t.Error("Expected TelegramAPI error type")
	}

	t.Logf("TelegramAPI error handled: %v", handledErr)
}

// TestErrorHandlingDatabase тестирует обработку ошибок базы данных.
func TestErrorHandlingDatabase(t *testing.T) {
	t.Parallel()

	adminNotifier := errorsPkg.NewAdminNotifier([]int64{123456789, 987654321}, nil)
	errorHandler := errorsPkg.NewErrorHandler(adminNotifier)

	dbErr := errorsPkg.ErrDatabaseConnectionFailed

	handledErr := errorHandler.HandleDatabaseError(dbErr, 12345, 67890, "GetUser")
	if handledErr == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errorsPkg.IsDatabaseError(handledErr) {
		t.Error("Expected Database error type")
	}

	t.Logf("Database error handled: %v", handledErr)
}

// TestErrorHandlingValidation тестирует обработку ошибок валидации.
func TestErrorHandlingValidation(t *testing.T) {
	t.Parallel()

	adminNotifier := errorsPkg.NewAdminNotifier([]int64{123456789, 987654321}, nil)
	errorHandler := errorsPkg.NewErrorHandler(adminNotifier)

	validationErr := errorsPkg.ErrInvalidUserInput

	handledErr := errorHandler.HandleValidationError(validationErr, 12345, 67890, "ValidateInput")
	if handledErr == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errorsPkg.IsValidationError(handledErr) {
		t.Error("Expected Validation error type")
	}

	t.Logf("Validation error handled: %v", handledErr)
}

// TestErrorHandlingCache тестирует обработку ошибок кэша.
func TestErrorHandlingCache(t *testing.T) {
	t.Parallel()

	adminNotifier := errorsPkg.NewAdminNotifier([]int64{123456789, 987654321}, nil)
	errorHandler := errorsPkg.NewErrorHandler(adminNotifier)

	cacheErr := errorsPkg.ErrRedisConnectionFailed

	handledErr := errorHandler.HandleCacheError(cacheErr, 12345, 67890, "GetCachedData")
	if handledErr == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errorsPkg.IsCacheError(handledErr) {
		t.Error("Expected Cache error type")
	}

	t.Logf("Cache error handled: %v", handledErr)
}

// TestRequestContextIntegration тестирует интеграцию RequestContext.
func TestRequestContextIntegration(t *testing.T) {
	t.Parallel()

	// Создаем контекст запроса
	ctx := errorsPkg.NewRequestContext(12345, 67890, "TestOperation")

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

// TestAdminNotifierIntegration тестирует интеграцию уведомлений администраторов.
func TestAdminNotifierIntegration(t *testing.T) {
	t.Parallel()

	adminNotifier := errorsPkg.NewAdminNotifier([]int64{123456789, 987654321}, nil)

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
