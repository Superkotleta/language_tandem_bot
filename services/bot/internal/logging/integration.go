package logging

import (
	"language-exchange-bot/internal/errors"
)

// LoggingService предоставляет централизованное логирование
type LoggingService struct {
	telegramLogger   *TelegramLogger
	databaseLogger   *DatabaseLogger
	cacheLogger      *CacheLogger
	validationLogger *ValidationLogger
	errorHandler     *errors.ErrorHandler
}

// NewLoggingService создает новый сервис логирования
func NewLoggingService(errorHandler *errors.ErrorHandler) *LoggingService {
	return &LoggingService{
		telegramLogger:   NewTelegramLogger(),
		databaseLogger:   NewDatabaseLogger(),
		cacheLogger:      NewCacheLogger(),
		validationLogger: NewValidationLogger(),
		errorHandler:     errorHandler,
	}
}

// Telegram возвращает логгер для Telegram
func (ls *LoggingService) Telegram() *TelegramLogger {
	return ls.telegramLogger
}

// Database возвращает логгер для базы данных
func (ls *LoggingService) Database() *DatabaseLogger {
	return ls.databaseLogger
}

// Cache возвращает логгер для кэша
func (ls *LoggingService) Cache() *CacheLogger {
	return ls.cacheLogger
}

// Validation возвращает логгер для валидации
func (ls *LoggingService) Validation() *ValidationLogger {
	return ls.validationLogger
}

// LogErrorWithContext логирует ошибку с полным контекстом
func (ls *LoggingService) LogErrorWithContext(err error, requestID string, userID, chatID int64, operation, component string) {
	if customErr, ok := err.(*errors.CustomError); ok {
		// Логируем в зависимости от типа ошибки
		switch customErr.Type {
		case errors.ErrorTypeTelegramAPI:
			ls.telegramLogger.LogError(err, chatID, userID, operation, requestID)
		case errors.ErrorTypeDatabase:
			ls.databaseLogger.ErrorWithContext(
				"Database error",
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
		case errors.ErrorTypeCache:
			ls.cacheLogger.LogCacheError(err, operation, requestID)
		case errors.ErrorTypeValidation:
			ls.validationLogger.LogValidationError(err, operation, requestID)
		default:
			// Логируем в общий логгер
			ls.telegramLogger.LogError(err, chatID, userID, operation, requestID)
		}
	} else {
		// Логируем в зависимости от компонента
		switch component {
		case "telegram":
			ls.telegramLogger.LogError(err, chatID, userID, operation, requestID)
		case "database":
			ls.databaseLogger.ErrorWithContext(
				"Database error",
				requestID,
				userID,
				chatID,
				operation,
				map[string]interface{}{
					"error": err.Error(),
				},
			)
		case "cache":
			ls.cacheLogger.LogCacheError(err, operation, requestID)
		case "validation":
			ls.validationLogger.LogValidationError(err, operation, requestID)
		default:
			ls.telegramLogger.LogError(err, chatID, userID, operation, requestID)
		}
	}
}

// LogRequestStart логирует начало запроса
func (ls *LoggingService) LogRequestStart(requestID string, userID, chatID int64, operation string) {
	ls.telegramLogger.InfoWithContext(
		"Request started",
		requestID,
		userID,
		chatID,
		operation,
	)
}

// LogRequestEnd логирует завершение запроса
func (ls *LoggingService) LogRequestEnd(requestID string, userID, chatID int64, operation string, success bool) {
	level := "completed"
	if !success {
		level = "failed"
	}

	ls.telegramLogger.InfoWithContext(
		"Request "+level,
		requestID,
		userID,
		chatID,
		operation,
		map[string]interface{}{
			"success": success,
		},
	)
}

// LogPerformance логирует метрики производительности
func (ls *LoggingService) LogPerformance(operation string, duration string, requestID string, fields map[string]interface{}) {
	ls.telegramLogger.InfoWithContext(
		"Performance metric",
		requestID,
		0,
		0,
		operation,
		map[string]interface{}{
			"duration": duration,
			"fields":   fields,
		},
	)
}

// LogSecurityEvent логирует события безопасности
func (ls *LoggingService) LogSecurityEvent(event string, userID, chatID int64, requestID string, fields map[string]interface{}) {
	ls.telegramLogger.WarnWithContext(
		"Security event: "+event,
		requestID,
		userID,
		chatID,
		"SecurityEvent",
		fields,
	)
}

// LogAdminAction логирует действия администратора
func (ls *LoggingService) LogAdminAction(action string, adminID, targetUserID int64, requestID string, fields map[string]interface{}) {
	ls.telegramLogger.InfoWithContext(
		"Admin action: "+action,
		requestID,
		adminID,
		0,
		"AdminAction",
		map[string]interface{}{
			"target_user_id": targetUserID,
			"fields":         fields,
		},
	)
}

// SetLogLevel устанавливает уровень логирования для всех компонентов
func (ls *LoggingService) SetLogLevel(level LogLevel) {
	ls.telegramLogger.SetLevel(level)
	ls.databaseLogger.SetLevel(level)
	ls.cacheLogger.SetLevel(level)
	ls.validationLogger.SetLevel(level)
}

// GetLogLevel возвращает текущий уровень логирования
func (ls *LoggingService) GetLogLevel() LogLevel {
	return ls.telegramLogger.GetLevel()
}
