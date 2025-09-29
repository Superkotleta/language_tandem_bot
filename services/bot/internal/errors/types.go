package errors

import (
	"fmt"
	"time"
)

// ErrorType определяет категорию ошибки
type ErrorType int

const (
	ErrorTypeTelegramAPI ErrorType = iota
	ErrorTypeDatabase
	ErrorTypeValidation
	ErrorTypeCache
	ErrorTypeNetwork
	ErrorTypeInternal
)

// String возвращает строковое представление типа ошибки
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeTelegramAPI:
		return "TelegramAPI"
	case ErrorTypeDatabase:
		return "Database"
	case ErrorTypeValidation:
		return "Validation"
	case ErrorTypeCache:
		return "Cache"
	case ErrorTypeNetwork:
		return "Network"
	case ErrorTypeInternal:
		return "Internal"
	default:
		return "Unknown"
	}
}

// CustomError представляет типизированную ошибку с контекстом
type CustomError struct {
	Type        ErrorType
	Message     string
	UserMessage string
	Context     map[string]interface{}
	RequestID   string
	Timestamp   time.Time
	Cause       error
}

// Error реализует интерфейс error
func (e *CustomError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s (caused by: %v)", e.Type.String(), e.Message, e.Cause.Error())
	}
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// Unwrap возвращает причину ошибки для error wrapping
func (e *CustomError) Unwrap() error {
	return e.Cause
}

// NewCustomError создает новую типизированную ошибку
func NewCustomError(errorType ErrorType, message, userMessage string, requestID string) *CustomError {
	return &CustomError{
		Type:        errorType,
		Message:     message,
		UserMessage: userMessage,
		Context:     make(map[string]interface{}),
		RequestID:   requestID,
		Timestamp:   time.Now(),
	}
}

// WithContext добавляет контекст к ошибке
func (e *CustomError) WithContext(key string, value interface{}) *CustomError {
	e.Context[key] = value
	return e
}

// WithCause добавляет причину ошибки
func (e *CustomError) WithCause(cause error) *CustomError {
	e.Cause = cause
	return e
}

// IsTelegramAPIError проверяет, является ли ошибка ошибкой Telegram API
func IsTelegramAPIError(err error) bool {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type == ErrorTypeTelegramAPI
	}
	return false
}

// IsDatabaseError проверяет, является ли ошибка ошибкой базы данных
func IsDatabaseError(err error) bool {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type == ErrorTypeDatabase
	}
	return false
}

// IsValidationError проверяет, является ли ошибка ошибкой валидации
func IsValidationError(err error) bool {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type == ErrorTypeValidation
	}
	return false
}

// IsCacheError проверяет, является ли ошибка ошибкой кэша
func IsCacheError(err error) bool {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type == ErrorTypeCache
	}
	return false
}
