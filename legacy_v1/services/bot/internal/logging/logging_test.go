package logging //nolint:testpackage

import (
	"language-exchange-bot/internal/errors"
	"testing"
	"time"
)

const (
	testDebugString = "DEBUG"
	testInfoString  = "INFO"
	testWarnString  = "WARN"
	testErrorString = "ERROR"
)

// TestLogger тестирует базовый логгер.
func TestLogger(t *testing.T) {
	t.Parallel()

	logger := NewLogger(DEBUG, "test")

	t.Run("LogLevels", func(t *testing.T) {
		t.Parallel()
		// Тест всех уровней логирования
		logger.Debug("Debug message")
		logger.Info("Info message")
		logger.Warn("Warning message")
		logger.Error("Error message")
	})

	t.Run("LogWithContext", func(t *testing.T) {
		t.Parallel()
		// Тест логирования с контекстом
		logger.DebugWithContext("Debug with context", "req_123", 12345, 67890, "TestOperation")
		logger.InfoWithContext("Info with context", "req_123", 12345, 67890, "TestOperation")
		logger.WarnWithContext("Warning with context", "req_123", 12345, 67890, "TestOperation")
		logger.ErrorWithContext("Error with context", "req_123", 12345, 67890, "TestOperation")
	})

	t.Run("LogWithFields", func(t *testing.T) {
		// Тест логирования с дополнительными полями
		logger.InfoWithContext(
			"Message with fields",
			"req_123",
			12345,
			67890,
			"TestOperation",
			map[string]interface{}{
				"field1": "value1",
				"field2": 123,
				"field3": true,
			},
		)
	})

	t.Run("LogLevelFiltering", func(t *testing.T) {
		// Тест фильтрации по уровню
		logger.SetLevel(ERROR)

		// Эти сообщения не должны выводиться
		logger.Debug("This should not appear")
		logger.Info("This should not appear")
		logger.Warn("This should not appear")

		// Это сообщение должно выводиться
		logger.Error("This should appear")
	})

	t.Log("Logger tests completed successfully")
}

// TestComponentLoggers тестирует специализированные логгеры.
func TestComponentLoggers(t *testing.T) {
	t.Run("TelegramLogger", func(t *testing.T) {
		telegramLogger := NewTelegramLogger()

		telegramLogger.LogMessageReceived(12345, 67890, "Hello world", "req_123")
		telegramLogger.LogMessageSent(12345, 67890, "Response message", "req_123")
		telegramLogger.LogCallbackReceived(12345, 67890, "callback_data", "req_123")
		telegramLogger.LogCommandExecuted(12345, 67890, "/start", "req_123")

		// Тест логирования ошибки
		testErr := errors.NewTelegramError("Test error", "User message", &errors.RequestContext{
			RequestID: "req_123",
			UserID:    67890,
			ChatID:    12345,
			Operation: "TestOperation",
		})
		telegramLogger.LogError(testErr, 12345, 67890, "TestOperation", "req_123")
	})

	t.Run("DatabaseLogger", func(t *testing.T) {
		databaseLogger := NewDatabaseLogger()

		databaseLogger.LogConnectionEstablished("req_123")
		databaseLogger.LogQueryExecuted("SELECT id, username, email FROM users", time.Millisecond*100, "req_123")
		databaseLogger.LogTransactionStarted("req_123")
		databaseLogger.LogTransactionCommitted("req_123")

		// Тест логирования ошибки
		testErr := errors.NewDatabaseError("Connection failed", "Database error", &errors.RequestContext{
			RequestID: "req_123",
			UserID:    0,
			ChatID:    0,
			Operation: "Connect",
		})
		databaseLogger.LogConnectionFailed(testErr, "req_123")
	})

	t.Run("CacheLogger", func(t *testing.T) {
		cacheLogger := NewCacheLogger()

		cacheLogger.LogCacheHit("user:123", "req_123")
		cacheLogger.LogCacheMiss("user:456", "req_123")
		cacheLogger.LogCacheSet("user:789", time.Minute*5, "req_123")
		cacheLogger.LogCacheInvalidated("user:*", "req_123")

		// Тест логирования ошибки
		testErr := errors.NewCacheError("Redis connection failed", "Cache error", &errors.RequestContext{
			RequestID: "req_123",
			UserID:    0,
			ChatID:    0,
			Operation: "GetFromCache",
		})
		cacheLogger.LogCacheError(testErr, "GetFromCache", "req_123")
	})

	t.Run("ValidationLogger", func(t *testing.T) {
		validationLogger := NewValidationLogger()

		validationLogger.LogValidationPassed("ValidateUser", "req_123")
		validationLogger.LogValidationFailed("ValidateUser", map[string][]string{
			"email": {"Invalid email format"},
			"name":  {"Name is required"},
		}, "req_123")

		// Тест логирования ошибки
		testErr := errors.NewValidationError("Validation failed", "Check your input", &errors.RequestContext{
			RequestID: "req_123",
			UserID:    67890,
			ChatID:    12345,
			Operation: "ValidateUser",
		})
		validationLogger.LogValidationError(testErr, "ValidateUser", "req_123")
	})

	t.Log("Component logger tests completed successfully")
}

// TestLoggingService тестирует сервис логирования.
func TestLoggingService(t *testing.T) {
	// Создаем мок errorHandler
	adminNotifier := errors.NewAdminNotifier([]int64{123456789}, nil)
	errorHandler := errors.NewErrorHandler(adminNotifier)

	// Создаем сервис логирования
	loggingService := NewLoggingService(errorHandler)

	t.Run("LogErrorWithContext", func(t *testing.T) {
		// Тест логирования ошибки с контекстом
		testErr := errors.NewTelegramError("Test error", "User message", &errors.RequestContext{
			RequestID: "req_123",
			UserID:    67890,
			ChatID:    12345,
			Operation: "TestOperation",
		})

		loggingService.LogErrorWithContext(testErr, "req_123", 67890, 12345, "TestOperation", "telegram")
	})

	t.Run("LogRequestStart", func(t *testing.T) {
		loggingService.LogRequestStart("req_123", 67890, 12345, "TestOperation")
	})

	t.Run("LogRequestEnd", func(t *testing.T) {
		loggingService.LogRequestEnd("req_123", 67890, 12345, "TestOperation", true)
		loggingService.LogRequestEnd("req_123", 67890, 12345, "TestOperation", false)
	})

	t.Run("LogPerformance", func(t *testing.T) {
		loggingService.LogPerformance("TestOperation", "100ms", "req_123", map[string]interface{}{
			"queries":    5,
			"cache_hits": 3,
		})
	})

	t.Run("LogSecurityEvent", func(t *testing.T) {
		loggingService.LogSecurityEvent("Invalid access attempt", 67890, 12345, "req_123", map[string]interface{}{
			"ip":         "192.168.1.1",
			"user_agent": "TelegramBot/1.0",
		})
	})

	t.Run("LogAdminAction", func(t *testing.T) {
		loggingService.LogAdminAction("User ban", 123456789, 67890, "req_123", map[string]interface{}{
			"reason":   "Spam",
			"duration": "24h",
		})
	})

	t.Run("SetLogLevel", func(t *testing.T) {
		t.Parallel()
		loggingService.SetLogLevel(ERROR)

		if loggingService.GetLogLevel() != ERROR {
			t.Error("Expected ERROR level")
		}

		loggingService.SetLogLevel(DEBUG)

		if loggingService.GetLogLevel() != DEBUG {
			t.Error("Expected DEBUG level")
		}
	})

	t.Log("Logging service tests completed successfully")
}

// TestLogLevelParsing тестирует парсинг уровней логирования.
func TestLogLevelParsing(t *testing.T) {
	t.Parallel()
	t.Run("ParseLogLevel", func(t *testing.T) {
		t.Parallel()

		if ParseLogLevel(testDebugString) != DEBUG {
			t.Error("Expected DEBUG level")
		}

		if ParseLogLevel(testInfoString) != INFO {
			t.Error("Expected INFO level")
		}

		if ParseLogLevel(testWarnString) != WARN {
			t.Error("Expected WARN level")
		}

		if ParseLogLevel(testErrorString) != ERROR {
			t.Error("Expected ERROR level")
		}

		if ParseLogLevel("INVALID") != INFO {
			t.Error("Expected INFO level for invalid input")
		}
	})

	t.Run("LogLevelString", func(t *testing.T) {
		t.Parallel()

		if DEBUG.String() != testDebugString {
			t.Error("Expected DEBUG string")
		}

		if INFO.String() != testInfoString {
			t.Error("Expected INFO string")
		}

		if WARN.String() != testWarnString {
			t.Error("Expected WARN string")
		}

		if ERROR.String() != testErrorString {
			t.Error("Expected ERROR string")
		}
	})

	t.Log("Log level parsing tests completed successfully")
}
