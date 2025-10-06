package errors

import (
	"errors"
	"fmt"
	"time"
)

// ErrorType определяет категорию ошибки.
type ErrorType int

// Типы ошибок для категоризации.
const (
	ErrorTypeTelegramAPI ErrorType = iota
	ErrorTypeDatabase
	ErrorTypeValidation
	ErrorTypeCache
	ErrorTypeNetwork
	ErrorTypeInternal
)

// String возвращает строковое представление типа ошибки.
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

// CustomError представляет типизированную ошибку с контекстом.
type CustomError struct {
	Type        ErrorType
	Message     string
	UserMessage string
	Context     map[string]interface{}
	RequestID   string
	Timestamp   time.Time
	Cause       error
}

// Error реализует интерфейс error.
func (e *CustomError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s (caused by: %v)", e.Type.String(), e.Message, e.Cause.Error())
	}

	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// Unwrap возвращает причину ошибки для error wrapping.
func (e *CustomError) Unwrap() error {
	return e.Cause
}

// NewCustomError создает новую типизированную ошибку.
func NewCustomError(errorType ErrorType, message, userMessage, requestID string) *CustomError {
	return &CustomError{
		Type:        errorType,
		Message:     message,
		UserMessage: userMessage,
		Context:     make(map[string]interface{}),
		RequestID:   requestID,
		Timestamp:   time.Now(),
		Cause:       nil,
	}
}

// WithContext добавляет контекст к ошибке.
func (e *CustomError) WithContext(key string, value interface{}) *CustomError {
	e.Context[key] = value

	return e
}

// WithCause добавляет причину ошибки.
func (e *CustomError) WithCause(cause error) *CustomError {
	e.Cause = cause

	return e
}

// IsTelegramAPIError проверяет, является ли ошибка ошибкой Telegram API.
func IsTelegramAPIError(err error) bool {
	customErr := &CustomError{
		Type:        ErrorTypeInternal,
		Message:     "check error",
		UserMessage: "internal error",
		Context:     map[string]interface{}{},
		RequestID:   "",
		Timestamp:   time.Now(),
		Cause:       nil,
	}
	if errors.As(err, &customErr) {
		return customErr.Type == ErrorTypeTelegramAPI
	}

	return false
}

// IsDatabaseError проверяет, является ли ошибка ошибкой базы данных.
func IsDatabaseError(err error) bool {
	customErr := &CustomError{
		Type:        ErrorTypeInternal,
		Message:     "check error",
		UserMessage: "internal error",
		Context:     map[string]interface{}{},
		RequestID:   "",
		Timestamp:   time.Now(),
		Cause:       nil,
	}
	if errors.As(err, &customErr) {
		return customErr.Type == ErrorTypeDatabase
	}

	return false
}

// IsValidationError проверяет, является ли ошибка ошибкой валидации.
func IsValidationError(err error) bool {
	customErr := &CustomError{
		Type:        ErrorTypeInternal,
		Message:     "check error",
		UserMessage: "internal error",
		Context:     map[string]interface{}{},
		RequestID:   "",
		Timestamp:   time.Now(),
		Cause:       nil,
	}
	if errors.As(err, &customErr) {
		return customErr.Type == ErrorTypeValidation
	}

	return false
}

// IsCacheError проверяет, является ли ошибка ошибкой кэша.
func IsCacheError(err error) bool {
	customErr := &CustomError{
		Type:        ErrorTypeInternal,
		Message:     "check error",
		UserMessage: "internal error",
		Context:     map[string]interface{}{},
		RequestID:   "",
		Timestamp:   time.Now(),
		Cause:       nil,
	}
	if errors.As(err, &customErr) {
		return customErr.Type == ErrorTypeCache
	}

	return false
}

// Статические ошибки для замены динамических.
var (
	// ErrInterestAlreadySelected - ошибка валидации.
	ErrInterestAlreadySelected = NewCustomError(
		ErrorTypeValidation, "интерес уже выбран", "Этот интерес уже выбран", "",
	)
	// ErrMaxPrimaryInterestsReached - ошибка валидации.
	ErrMaxPrimaryInterestsReached = NewCustomError(
		ErrorTypeValidation, "достигнут максимум основных интересов",
		"Достигнут максимум основных интересов", "",
	)
	// ErrMinPrimaryInterestsRequired - ошибка валидации.
	ErrMinPrimaryInterestsRequired = NewCustomError(
		ErrorTypeValidation, "необходимо выбрать минимум основных интересов",
		"Необходимо выбрать минимум основных интересов", "",
	)

	// ErrUnsafeFilePath - ошибка файловой системы.
	ErrUnsafeFilePath = NewCustomError(ErrorTypeInternal, "небезопасный путь к файлу", "Ошибка доступа к файлу", "")

	// ErrFeedbackTooShort - ошибка отзывов.
	ErrFeedbackTooShort = NewCustomError(
		ErrorTypeValidation, "отзыв слишком короткий", "Отзыв должен содержать минимум символов", "",
	)
	// ErrFeedbackTooLong - ошибка отзывов.
	ErrFeedbackTooLong = NewCustomError(
		ErrorTypeValidation, "отзыв слишком длинный", "Отзыв превышает максимальную длину", "",
	)
	// ErrFeedbackNotFound - ошибка отзывов.
	ErrFeedbackNotFound = NewCustomError(ErrorTypeDatabase, "отзыв не найден", "Отзыв не найден в базе данных", "")

	// ErrUserNotFound - ошибка пользователей.
	ErrUserNotFound = NewCustomError(ErrorTypeDatabase, "пользователь не найден", "Пользователь не найден", "")

	// ErrTelegramAPIRateLimit - ошибка тестов.
	ErrTelegramAPIRateLimit = NewCustomError(
		ErrorTypeTelegramAPI, "превышен лимит запросов Telegram API", "Превышен лимит запросов", "",
	)
	// ErrRateLimitExceeded - превышен лимит запросов пользователя.
	ErrRateLimitExceeded = NewCustomError(
		ErrorTypeValidation, "превышен лимит запросов пользователя", "Слишком много запросов. Попробуйте позже", "",
	)
	// ErrDatabaseConnectionFailed - ошибка тестов.
	ErrDatabaseConnectionFailed = NewCustomError(
		ErrorTypeDatabase, "ошибка подключения к базе данных", "Ошибка подключения к базе данных", "",
	)
	// ErrInvalidUserInput - ошибка тестов.
	ErrInvalidUserInput = NewCustomError(
		ErrorTypeValidation, "некорректные данные пользователя", "Некорректные данные", "",
	)
	// ErrRedisConnectionFailed - ошибка тестов.
	ErrRedisConnectionFailed = NewCustomError(
		ErrorTypeCache, "ошибка подключения к Redis", "Ошибка подключения к кэшу", "",
	)

	// ===== НОВЫЕ ТИПЫ ОШИБОК =====

	// Ошибки изолированного редактирования
	ErrEditSessionNotFound = NewCustomError(
		ErrorTypeInternal, "сессия редактирования не найдена", "Сессия редактирования не найдена", "",
	)
	ErrEditSessionExpired = NewCustomError(
		ErrorTypeInternal, "сессия редактирования истекла", "Сессия редактирования истекла", "",
	)
	ErrEditSessionInvalid = NewCustomError(
		ErrorTypeValidation, "некорректная сессия редактирования", "Некорректная сессия редактирования", "",
	)

	// Ошибки батчинга
	ErrBatchOperationFailed = NewCustomError(
		ErrorTypeDatabase, "ошибка батчевой операции", "Ошибка при выполнении батчевой операции", "",
	)
	ErrBatchSizeExceeded = NewCustomError(
		ErrorTypeValidation, "превышен размер батча", "Превышен максимальный размер батча", "",
	)
	ErrBatchTimeout = NewCustomError(
		ErrorTypeDatabase, "таймаут батчевой операции", "Превышено время выполнения батчевой операции", "",
	)

	// Ошибки кеширования
	ErrCacheOperationFailed = NewCustomError(
		ErrorTypeCache, "ошибка операции с кэшем", "Ошибка при работе с кэшем", "",
	)
	ErrCacheSerializationFailed = NewCustomError(
		ErrorTypeCache, "ошибка сериализации кэша", "Ошибка сериализации данных кэша", "",
	)
	ErrCacheDeserializationFailed = NewCustomError(
		ErrorTypeCache, "ошибка десериализации кэша", "Ошибка десериализации данных кэша", "",
	)

	// Ошибки трейсинга
	ErrTraceNotFound = NewCustomError(
		ErrorTypeInternal, "трейс не найден", "Трейс запроса не найден", "",
	)
	ErrTraceExpired = NewCustomError(
		ErrorTypeInternal, "трейс истек", "Трейс запроса истек", "",
	)

	// Ошибки метрик
	ErrMetricCollectionFailed = NewCustomError(
		ErrorTypeInternal, "ошибка сбора метрик", "Ошибка при сборе метрик производительности", "",
	)
	ErrMetricExportFailed = NewCustomError(
		ErrorTypeInternal, "ошибка экспорта метрик", "Ошибка при экспорте метрик", "",
	)

	// Ошибки конфигурации
	ErrConfigNotFound = NewCustomError(
		ErrorTypeInternal, "конфигурация не найдена", "Конфигурация не найдена", "",
	)
	ErrConfigInvalid = NewCustomError(
		ErrorTypeValidation, "некорректная конфигурация", "Некорректная конфигурация", "",
	)

	// Ошибки локализации
	ErrLocalizationNotFound = NewCustomError(
		ErrorTypeInternal, "локализация не найдена", "Перевод не найден", "",
	)
	ErrLocalizationInvalid = NewCustomError(
		ErrorTypeValidation, "некорректная локализация", "Некорректная локализация", "",
	)

	// Ошибки уведомлений
	ErrNotificationFailed = NewCustomError(
		ErrorTypeTelegramAPI, "ошибка отправки уведомления", "Ошибка отправки уведомления", "",
	)
	ErrNotificationRateLimit = NewCustomError(
		ErrorTypeTelegramAPI, "превышен лимит уведомлений", "Превышен лимит отправки уведомлений", "",
	)
)
