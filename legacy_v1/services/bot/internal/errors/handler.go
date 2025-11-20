// Package errors provides error handling and alerting functionality.
package errors

import (
	"errors"
	"log"
)

// ErrorHandler обрабатывает ошибки централизованно.
type ErrorHandler struct {
	notifier AdminNotifier
}

// AdminNotifier интерфейс для уведомления администраторов.
type AdminNotifier interface {
	NotifyCriticalError(err *CustomError)
	NotifyTelegramAPIError(err *CustomError, chatID int64)
}

// NewErrorHandler создает новый обработчик ошибок.
func NewErrorHandler(notifier AdminNotifier) *ErrorHandler {
	return &ErrorHandler{
		notifier: notifier,
	}
}

// Handle обрабатывает ошибку с контекстом.
func (h *ErrorHandler) Handle(err error, ctx *RequestContext) error {
	if err == nil {
		return nil
	}

	// Логируем ошибку
	h.logError(err, ctx)

	// Если это критическая ошибка, уведомляем администраторов
	if h.isCriticalError(err) {
		var customErr *CustomError
		if errors.As(err, &customErr) {
			h.notifier.NotifyCriticalError(customErr)
		}
	}

	// Если это ошибка Telegram API, уведомляем администраторов
	if IsTelegramAPIError(err) {
		var customErr *CustomError
		if errors.As(err, &customErr) {
			h.notifier.NotifyTelegramAPIError(customErr, ctx.ChatID)
		}
	}

	return err
}

// HandleTelegramError обрабатывает ошибки Telegram API.
func (h *ErrorHandler) HandleTelegramError(err error, chatID, userID int64, operation string) error {
	if err == nil {
		return nil
	}

	ctx := NewRequestContext(userID, chatID, operation)

	// Создаем типизированную ошибку
	customErr := NewTelegramError(
		err.Error(),
		"Произошла ошибка при обработке запроса. Попробуйте позже.",
		ctx,
	).WithCause(err)

	return h.Handle(customErr, ctx)
}

// HandleDatabaseError обрабатывает ошибки базы данных.
func (h *ErrorHandler) HandleDatabaseError(err error, userID, chatID int64, operation string) error {
	if err == nil {
		return nil
	}

	ctx := NewRequestContext(userID, chatID, operation)

	// Создаем типизированную ошибку
	customErr := NewDatabaseError(
		err.Error(),
		"Временные проблемы с базой данных. Попробуйте позже.",
		ctx,
	).WithCause(err)

	return h.Handle(customErr, ctx)
}

// HandleValidationError обрабатывает ошибки валидации.
func (h *ErrorHandler) HandleValidationError(err error, userID, chatID int64, operation string) error {
	if err == nil {
		return nil
	}

	ctx := NewRequestContext(userID, chatID, operation)

	// Создаем типизированную ошибку
	customErr := NewValidationError(
		err.Error(),
		"Некорректные данные. Проверьте введенную информацию.",
		ctx,
	).WithCause(err)

	return h.Handle(customErr, ctx)
}

// HandleCacheError обрабатывает ошибки кэша.
func (h *ErrorHandler) HandleCacheError(err error, userID, chatID int64, operation string) error {
	if err == nil {
		return nil
	}

	ctx := NewRequestContext(userID, chatID, operation)

	// Создаем типизированную ошибку
	customErr := NewCacheError(
		err.Error(),
		"Временные проблемы с кэшем. Попробуйте позже.",
		ctx,
	).WithCause(err)

	return h.Handle(customErr, ctx)
}

// logError логирует ошибку с контекстом.
func (h *ErrorHandler) logError(err error, ctx *RequestContext) {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		log.Printf("[%s] %s: %s (User: %d, Chat: %d, Operation: %s)",
			customErr.RequestID,
			customErr.Type.String(),
			customErr.Message,
			ctx.UserID,
			ctx.ChatID,
			ctx.Operation,
		)
	} else {
		log.Printf("[%s] Error: %v (User: %d, Chat: %d, Operation: %s)",
			ctx.RequestID,
			err,
			ctx.UserID,
			ctx.ChatID,
			ctx.Operation,
		)
	}
}

// isCriticalError определяет, является ли ошибка критической.
func (h *ErrorHandler) isCriticalError(err error) bool {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		// Критические ошибки: Database, Network, Internal
		return customErr.Type == ErrorTypeDatabase ||
			customErr.Type == ErrorTypeNetwork ||
			customErr.Type == ErrorTypeInternal
	}

	return false
}
