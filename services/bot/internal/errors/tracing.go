package errors

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// Константы для трассировки.
const (
	// maxRandomValue - максимальное значение для генерации случайной части RequestID.
	maxRandomValue = 10000
)

// RequestContext содержит контекст запроса.
type RequestContext struct {
	RequestID string
	UserID    int64
	ChatID    int64
	Operation string
	Timestamp time.Time
}

// NewRequestContext создает новый контекст запроса.
func NewRequestContext(userID, chatID int64, operation string) *RequestContext {
	return &RequestContext{
		RequestID: generateRequestID(),
		UserID:    userID,
		ChatID:    chatID,
		Operation: operation,
		Timestamp: time.Now(),
	}
}

// generateRequestID генерирует уникальный RequestID.
func generateRequestID() string {
	// Используем timestamp + криптографически случайные символы для уникальности
	timestamp := time.Now().UnixNano()

	// Генерируем случайное число с помощью crypto/rand
	var randomBytes [8]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		// Fallback на timestamp-based случайность в случае ошибки
		randomPart := timestamp % maxRandomValue
		return fmt.Sprintf("req_%d_%d", timestamp, randomPart)
	}

	randomValue := binary.BigEndian.Uint64(randomBytes[:])
	randomPart := randomValue % maxRandomValue

	return fmt.Sprintf("req_%d_%d", timestamp, randomPart)
}

// WithContext создает ошибку с контекстом.
func WithContext(err error, ctx *RequestContext) *CustomError {
	customErr := &CustomError{
		Type:        ErrorTypeInternal,
		Message:     "context error",
		UserMessage: "internal error",
		Context:     map[string]interface{}{},
		RequestID:   "",
		Timestamp:   time.Now(),
		Cause:       nil,
	}
	if errors.As(err, &customErr) {
		customErr.RequestID = ctx.RequestID
		customErr.Context["user_id"] = ctx.UserID
		customErr.Context["chat_id"] = ctx.ChatID
		customErr.Context["operation"] = ctx.Operation

		return customErr
	}

	// Если ошибка не CustomError, оборачиваем её
	return &CustomError{
		Type:        ErrorTypeInternal,
		Message:     err.Error(),
		UserMessage: "internal error",
		RequestID:   ctx.RequestID,
		Context: map[string]interface{}{
			"user_id":   ctx.UserID,
			"chat_id":   ctx.ChatID,
			"operation": ctx.Operation,
		},
		Timestamp: time.Now(),
		Cause:     err,
	}
}

// NewTelegramError создает ошибку Telegram API с контекстом.
func NewTelegramError(message, userMessage string, ctx *RequestContext) *CustomError {
	return &CustomError{
		Type:        ErrorTypeTelegramAPI,
		Message:     message,
		UserMessage: userMessage,
		RequestID:   ctx.RequestID,
		Context: map[string]interface{}{
			"user_id":   ctx.UserID,
			"chat_id":   ctx.ChatID,
			"operation": ctx.Operation,
		},
		Timestamp: time.Now(),
		Cause:     nil,
	}
}

// NewDatabaseError создает ошибку базы данных с контекстом.
func NewDatabaseError(message, userMessage string, ctx *RequestContext) *CustomError {
	return &CustomError{
		Type:        ErrorTypeDatabase,
		Message:     message,
		UserMessage: userMessage,
		RequestID:   ctx.RequestID,
		Context: map[string]interface{}{
			"user_id":   ctx.UserID,
			"chat_id":   ctx.ChatID,
			"operation": ctx.Operation,
		},
		Timestamp: time.Now(),
		Cause:     nil,
	}
}

// NewValidationError создает ошибку валидации с контекстом.
func NewValidationError(message, userMessage string, ctx *RequestContext) *CustomError {
	return &CustomError{
		Type:        ErrorTypeValidation,
		Message:     message,
		UserMessage: userMessage,
		RequestID:   ctx.RequestID,
		Context: map[string]interface{}{
			"user_id":   ctx.UserID,
			"chat_id":   ctx.ChatID,
			"operation": ctx.Operation,
		},
		Timestamp: time.Now(),
		Cause:     nil,
	}
}

// NewCacheError создает ошибку кэша с контекстом.
func NewCacheError(message, userMessage string, ctx *RequestContext) *CustomError {
	return &CustomError{
		Type:        ErrorTypeCache,
		Message:     message,
		UserMessage: userMessage,
		RequestID:   ctx.RequestID,
		Context: map[string]interface{}{
			"user_id":   ctx.UserID,
			"chat_id":   ctx.ChatID,
			"operation": ctx.Operation,
		},
		Timestamp: time.Now(),
		Cause:     nil,
	}
}
