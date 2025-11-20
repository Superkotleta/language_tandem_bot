package logging

import (
	"errors"
	customErrors "language-exchange-bot/internal/errors"
	"time"
)

// ComponentLogger предоставляет специализированное логирование для компонентов.
type ComponentLogger struct {
	*Logger
}

// NewComponentLogger создает логгер для компонента.
func NewComponentLogger(component string) *ComponentLogger {
	return &ComponentLogger{
		Logger: NewLoggerFromEnv(component),
	}
}

// TelegramLogger предоставляет логирование для Telegram бота.
type TelegramLogger struct {
	*ComponentLogger
}

// NewTelegramLogger создает логгер для Telegram бота.
func NewTelegramLogger() *TelegramLogger {
	return &TelegramLogger{
		ComponentLogger: NewComponentLogger("telegram"),
	}
}

// LogMessageReceived логирует получение сообщения.
func (tl *TelegramLogger) LogMessageReceived(chatID, userID int64, text, requestID string) {
	tl.InfoWithContext(
		"Message received",
		requestID,
		userID,
		chatID,
		"HandleMessage",
		map[string]interface{}{
			"text_length": len(text),
			"has_text":    text != "",
		},
	)
}

// LogMessageSent логирует отправку сообщения.
func (tl *TelegramLogger) LogMessageSent(chatID, userID int64, text, requestID string) {
	tl.InfoWithContext(
		"Message sent",
		requestID,
		userID,
		chatID,
		"SendMessage",
		map[string]interface{}{
			"text_length": len(text),
		},
	)
}

// LogCallbackReceived логирует получение callback query.
func (tl *TelegramLogger) LogCallbackReceived(chatID, userID int64, data, requestID string) {
	tl.InfoWithContext(
		"Callback received",
		requestID,
		userID,
		chatID,
		"HandleCallback",
		map[string]interface{}{
			"callback_data": data,
		},
	)
}

// LogCommandExecuted логирует выполнение команды.
func (tl *TelegramLogger) LogCommandExecuted(chatID, userID int64, command, requestID string) {
	tl.InfoWithContext(
		"Command executed",
		requestID,
		userID,
		chatID,
		"ExecuteCommand",
		map[string]interface{}{
			"command": command,
		},
	)
}

// LogError логирует ошибку с контекстом.
func (tl *TelegramLogger) LogError(err error, chatID, userID int64, operation, requestID string) {
	customErr := &customErrors.CustomError{}
	if errors.As(err, &customErr) {
		tl.ErrorWithContext(
			"Error occurred",
			requestID,
			userID,
			chatID,
			operation,
			map[string]interface{}{
				"error_type": customErr.Type.String(),
				"error_msg":  customErr.Message,
				"user_msg":   customErr.UserMessage,
			},
		)
	} else {
		tl.ErrorWithContext(
			"Error occurred",
			requestID,
			userID,
			chatID,
			operation,
			map[string]interface{}{
				"error_msg": err.Error(),
			},
		)
	}
}

// DatabaseLogger предоставляет логирование для базы данных.
type DatabaseLogger struct {
	*ComponentLogger
}

// NewDatabaseLogger создает логгер для базы данных.
func NewDatabaseLogger() *DatabaseLogger {
	return &DatabaseLogger{
		ComponentLogger: NewComponentLogger("database"),
	}
}

// LogQueryExecuted логирует выполнение запроса.
func (dl *DatabaseLogger) LogQueryExecuted(query string, duration time.Duration, requestID string) {
	dl.DebugWithContext(
		"Query executed",
		requestID,
		0,
		0,
		"ExecuteQuery",
		map[string]interface{}{
			"query":    query,
			"duration": duration.String(),
		},
	)
}

// LogConnectionEstablished логирует установление соединения.
func (dl *DatabaseLogger) LogConnectionEstablished(requestID string) {
	dl.InfoWithContext(
		"Database connection established",
		requestID,
		0,
		0,
		"Connect",
	)
}

// LogConnectionFailed логирует ошибку соединения.
func (dl *DatabaseLogger) LogConnectionFailed(err error, requestID string) {
	dl.ErrorWithContext(
		"Database connection failed",
		requestID,
		0,
		0,
		"Connect",
		map[string]interface{}{
			"error": err.Error(),
		},
	)
}

// LogTransactionStarted логирует начало транзакции.
func (dl *DatabaseLogger) LogTransactionStarted(requestID string) {
	dl.DebugWithContext(
		"Transaction started",
		requestID,
		0,
		0,
		"BeginTransaction",
	)
}

// LogTransactionCommitted логирует коммит транзакции.
func (dl *DatabaseLogger) LogTransactionCommitted(requestID string) {
	dl.DebugWithContext(
		"Transaction committed",
		requestID,
		0,
		0,
		"CommitTransaction",
	)
}

// LogTransactionRolledBack логирует откат транзакции.
func (dl *DatabaseLogger) LogTransactionRolledBack(err error, requestID string) {
	dl.WarnWithContext(
		"Transaction rolled back",
		requestID,
		0,
		0,
		"RollbackTransaction",
		map[string]interface{}{
			"error": err.Error(),
		},
	)
}

// CacheLogger предоставляет логирование для кэша.
type CacheLogger struct {
	*ComponentLogger
}

// NewCacheLogger создает логгер для кэша.
func NewCacheLogger() *CacheLogger {
	return &CacheLogger{
		ComponentLogger: NewComponentLogger("cache"),
	}
}

// LogCacheHit логирует попадание в кэш.
func (cl *CacheLogger) LogCacheHit(key string, requestID string) {
	cl.DebugWithContext(
		"Cache hit",
		requestID,
		0,
		0,
		"GetFromCache",
		map[string]interface{}{
			"key": key,
		},
	)
}

// LogCacheMiss логирует промах кэша.
func (cl *CacheLogger) LogCacheMiss(key string, requestID string) {
	cl.DebugWithContext(
		"Cache miss",
		requestID,
		0,
		0,
		"GetFromCache",
		map[string]interface{}{
			"key": key,
		},
	)
}

// LogCacheSet логирует установку значения в кэш.
func (cl *CacheLogger) LogCacheSet(key string, ttl time.Duration, requestID string) {
	cl.DebugWithContext(
		"Cache set",
		requestID,
		0,
		0,
		"SetToCache",
		map[string]interface{}{
			"key": key,
			"ttl": ttl.String(),
		},
	)
}

// LogCacheInvalidated логирует инвалидацию кэша.
func (cl *CacheLogger) LogCacheInvalidated(pattern string, requestID string) {
	cl.InfoWithContext(
		"Cache invalidated",
		requestID,
		0,
		0,
		"InvalidateCache",
		map[string]interface{}{
			"pattern": pattern,
		},
	)
}

// LogCacheError логирует ошибку кэша.
func (cl *CacheLogger) LogCacheError(err error, operation, requestID string) {
	cl.ErrorWithContext(
		"Cache error",
		requestID,
		0,
		0,
		operation,
		map[string]interface{}{
			"error": err.Error(),
		},
	)
}

// ValidationLogger предоставляет логирование для валидации.
type ValidationLogger struct {
	*ComponentLogger
}

// NewValidationLogger создает логгер для валидации.
func NewValidationLogger() *ValidationLogger {
	return &ValidationLogger{
		ComponentLogger: NewComponentLogger("validation"),
	}
}

// LogValidationPassed логирует успешную валидацию.
func (vl *ValidationLogger) LogValidationPassed(operation string, requestID string) {
	vl.DebugWithContext(
		"Validation passed",
		requestID,
		0,
		0,
		operation,
	)
}

// LogValidationFailed логирует неудачную валидацию.
func (vl *ValidationLogger) LogValidationFailed(operation string, errors map[string][]string, requestID string) {
	vl.WarnWithContext(
		"Validation failed",
		requestID,
		0,
		0,
		operation,
		map[string]interface{}{
			"validation_errors": errors,
		},
	)
}

// LogValidationError логирует ошибку валидации.
func (vl *ValidationLogger) LogValidationError(err error, operation, requestID string) {
	vl.ErrorWithContext(
		"Validation error",
		requestID,
		0,
		0,
		operation,
		map[string]interface{}{
			"error": err.Error(),
		},
	)
}
