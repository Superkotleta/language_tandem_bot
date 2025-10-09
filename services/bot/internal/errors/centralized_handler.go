// Package errors provides centralized error handling and alerting.
package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// AlertLevel определяет уровень алерта.
type AlertLevel int

const (
	AlertLevelInfo AlertLevel = iota
	AlertLevelWarning
	AlertLevelCritical
	AlertLevelEmergency
)

// String возвращает строковое представление уровня алерта.
func (al AlertLevel) String() string {
	switch al {
	case AlertLevelInfo:
		return "INFO"
	case AlertLevelWarning:
		return "WARNING"
	case AlertLevelCritical:
		return "CRITICAL"
	case AlertLevelEmergency:
		return "EMERGENCY"
	default:
		return "UNKNOWN"
	}
}

// Alert представляет алерт для администраторов.
type Alert struct {
	ID         string                 `json:"id"`
	Level      AlertLevel             `json:"level"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Error      *CustomError           `json:"error"`
	Context    map[string]interface{} `json:"context"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolvedAt,omitempty"`
}

// CentralizedErrorHandler предоставляет централизованную обработку ошибок.
type CentralizedErrorHandler struct {
	alerts       map[string]*Alert
	notifiers    []AlertNotifier
	logger       *log.Logger
	alertCounter int
	mutex        chan struct{} // Семафор для потокобезопасности
}

// AlertNotifier интерфейс для уведомления об алертах.
type AlertNotifier interface {
	NotifyAlert(alert *Alert) error
	GetName() string
}

// NewCentralizedErrorHandler создает новый централизованный обработчик ошибок.
func NewCentralizedErrorHandler() *CentralizedErrorHandler {
	return &CentralizedErrorHandler{
		alerts:    make(map[string]*Alert),
		notifiers: make([]AlertNotifier, 0),
		logger:    log.New(os.Stdout, "[ERROR_HANDLER] ", log.LstdFlags),
		mutex:     make(chan struct{}, 1),
	}
}

// RegisterNotifier регистрирует уведомитель алертов.
func (ceh *CentralizedErrorHandler) RegisterNotifier(notifier AlertNotifier) {
	ceh.mutex <- struct{}{} // Блокируем

	defer func() { <-ceh.mutex }() // Разблокируем

	ceh.notifiers = append(ceh.notifiers, notifier)
	ceh.logger.Printf("Registered alert notifier: %s", notifier.GetName())
}

// HandleError обрабатывает ошибку централизованно.
func (ceh *CentralizedErrorHandler) HandleError(ctx context.Context, err error, requestID string, userID, chatID int64, operation string) error {
	if err == nil {
		return nil
	}

	// Логируем ошибку
	ceh.logError(err, requestID, userID, chatID, operation)

	// Создаем алерт если это критическая ошибка
	if ceh.isCriticalError(err) {
		alert := ceh.createAlert(err, requestID, userID, chatID, operation)
		ceh.sendAlert(alert)
	}

	return err
}

// HandleCustomError обрабатывает кастомную ошибку.
func (ceh *CentralizedErrorHandler) HandleCustomError(ctx context.Context, customErr *CustomError, requestID string, userID, chatID int64, operation string) error {
	if customErr == nil {
		return nil
	}

	// Логируем ошибку
	ceh.logCustomError(customErr, requestID, userID, chatID, operation)

	// Создаем алерт если это критическая ошибка
	if ceh.isCriticalCustomError(customErr) {
		alert := ceh.createAlertFromCustomError(customErr, requestID, userID, chatID, operation)
		ceh.sendAlert(alert)
	}

	return customErr
}

// logError логирует обычную ошибку.
func (ceh *CentralizedErrorHandler) logError(err error, requestID string, userID, chatID int64, operation string) {
	logData := map[string]interface{}{
		"error":      err.Error(),
		"request_id": requestID,
		"user_id":    userID,
		"chat_id":    chatID,
		"operation":  operation,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		ceh.logger.Printf("Failed to marshal error log: %v", err)

		return
	}

	ceh.logger.Printf("ERROR: %s", string(jsonData))
}

// logCustomError логирует кастомную ошибку.
func (ceh *CentralizedErrorHandler) logCustomError(customErr *CustomError, requestID string, userID, chatID int64, operation string) {
	logData := map[string]interface{}{
		"error_type":    customErr.Type.String(),
		"error_message": customErr.Message,
		"user_message":  customErr.UserMessage,
		"request_id":    requestID,
		"user_id":       userID,
		"chat_id":       chatID,
		"operation":     operation,
		"context":       customErr.Context,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		ceh.logger.Printf("Failed to marshal custom error log: %v", err)

		return
	}

	ceh.logger.Printf("CUSTOM_ERROR: %s", string(jsonData))
}

// isCriticalError определяет, является ли ошибка критической.
func (ceh *CentralizedErrorHandler) isCriticalError(err error) bool {
	// Простая проверка по тексту ошибки
	errorText := err.Error()

	criticalKeywords := []string{
		"database connection failed",
		"redis connection failed",
		"telegram api rate limit",
		"critical error",
		"fatal error",
		"out of memory",
		"disk full",
		"network unreachable",
	}

	for _, keyword := range criticalKeywords {
		if contains(errorText, keyword) {
			return true
		}
	}

	return false
}

// isCriticalCustomError определяет, является ли кастомная ошибка критической.
func (ceh *CentralizedErrorHandler) isCriticalCustomError(customErr *CustomError) bool {
	// Критические типы ошибок
	criticalTypes := []ErrorType{
		ErrorTypeDatabase,
		ErrorTypeNetwork,
		ErrorTypeInternal,
	}

	for _, errorType := range criticalTypes {
		if customErr.Type == errorType {
			return true
		}
	}

	return false
}

// createAlert создает алерт из обычной ошибки.
func (ceh *CentralizedErrorHandler) createAlert(err error, requestID string, userID, chatID int64, operation string) *Alert {
	ceh.alertCounter++
	alertID := fmt.Sprintf("alert_%d_%d", time.Now().Unix(), ceh.alertCounter)

	alert := &Alert{
		ID:      alertID,
		Level:   ceh.determineAlertLevel(err),
		Title:   "Critical Error Alert",
		Message: err.Error(),
		Context: map[string]interface{}{
			"request_id": requestID,
			"user_id":    userID,
			"chat_id":    chatID,
			"operation":  operation,
			"timestamp":  time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now(),
		Resolved:  false,
	}

	return alert
}

// createAlertFromCustomError создает алерт из кастомной ошибки.
func (ceh *CentralizedErrorHandler) createAlertFromCustomError(customErr *CustomError, requestID string, userID, chatID int64, operation string) *Alert {
	ceh.alertCounter++
	alertID := fmt.Sprintf("alert_%d_%d", time.Now().Unix(), ceh.alertCounter)

	alert := &Alert{
		ID:      alertID,
		Level:   ceh.determineCustomErrorAlertLevel(customErr),
		Title:   customErr.Type.String() + " Error Alert",
		Message: customErr.Message,
		Error:   customErr,
		Context: map[string]interface{}{
			"request_id":   requestID,
			"user_id":      userID,
			"chat_id":      chatID,
			"operation":    operation,
			"error_type":   customErr.Type.String(),
			"user_message": customErr.UserMessage,
			"timestamp":    time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now(),
		Resolved:  false,
	}

	return alert
}

// determineAlertLevel определяет уровень алерта для обычной ошибки.
func (ceh *CentralizedErrorHandler) determineAlertLevel(err error) AlertLevel {
	errorText := err.Error()

	if contains(errorText, "database connection failed") || contains(errorText, "redis connection failed") {
		return AlertLevelEmergency
	}

	if contains(errorText, "telegram api rate limit") {
		return AlertLevelCritical
	}

	if contains(errorText, "critical error") || contains(errorText, "fatal error") {
		return AlertLevelCritical
	}

	return AlertLevelWarning
}

// determineCustomErrorAlertLevel определяет уровень алерта для кастомной ошибки.
func (ceh *CentralizedErrorHandler) determineCustomErrorAlertLevel(customErr *CustomError) AlertLevel {
	switch customErr.Type {
	case ErrorTypeDatabase:
		return AlertLevelEmergency
	case ErrorTypeNetwork:
		return AlertLevelCritical
	case ErrorTypeInternal:
		return AlertLevelCritical
	case ErrorTypeTelegramAPI:
		return AlertLevelWarning
	case ErrorTypeCache:
		return AlertLevelWarning
	case ErrorTypeValidation:
		return AlertLevelInfo
	default:
		return AlertLevelWarning
	}
}

// sendAlert отправляет алерт всем зарегистрированным уведомителям.
func (ceh *CentralizedErrorHandler) sendAlert(alert *Alert) {
	ceh.mutex <- struct{}{} // Блокируем

	defer func() { <-ceh.mutex }() // Разблокируем

	// Сохраняем алерт
	ceh.alerts[alert.ID] = alert

	// Отправляем всем уведомителям
	for _, notifier := range ceh.notifiers {
		go func(n AlertNotifier) {
			if err := n.NotifyAlert(alert); err != nil {
				ceh.logger.Printf("Failed to send alert via %s: %v", n.GetName(), err)
			}
		}(notifier)
	}

	ceh.logger.Printf("Alert sent: %s (Level: %s)", alert.ID, alert.Level.String())
}

// ResolveAlert разрешает алерт.
func (ceh *CentralizedErrorHandler) ResolveAlert(alertID string) error {
	ceh.mutex <- struct{}{} // Блокируем

	defer func() { <-ceh.mutex }() // Разблокируем

	alert, exists := ceh.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.Resolved = true
	alert.ResolvedAt = &now

	ceh.logger.Printf("Alert resolved: %s", alertID)

	return nil
}

// GetAlerts возвращает все алерты.
func (ceh *CentralizedErrorHandler) GetAlerts() map[string]*Alert {
	ceh.mutex <- struct{}{} // Блокируем

	defer func() { <-ceh.mutex }() // Разблокируем

	// Создаем копию для безопасности
	result := make(map[string]*Alert)
	for key, alert := range ceh.alerts {
		result[key] = alert
	}

	return result
}

// GetActiveAlerts возвращает активные алерты.
func (ceh *CentralizedErrorHandler) GetActiveAlerts() map[string]*Alert {
	ceh.mutex <- struct{}{} // Блокируем

	defer func() { <-ceh.mutex }() // Разблокируем

	result := make(map[string]*Alert)

	for key, alert := range ceh.alerts {
		if !alert.Resolved {
			result[key] = alert
		}
	}

	return result
}

// GetAlertsByLevel возвращает алерты по уровню.
func (ceh *CentralizedErrorHandler) GetAlertsByLevel(level AlertLevel) map[string]*Alert {
	ceh.mutex <- struct{}{} // Блокируем

	defer func() { <-ceh.mutex }() // Разблокируем

	result := make(map[string]*Alert)

	for key, alert := range ceh.alerts {
		if alert.Level == level {
			result[key] = alert
		}
	}

	return result
}

// ClearResolvedAlerts очищает разрешенные алерты.
func (ceh *CentralizedErrorHandler) ClearResolvedAlerts() {
	ceh.mutex <- struct{}{} // Блокируем

	defer func() { <-ceh.mutex }() // Разблокируем

	for key, alert := range ceh.alerts {
		if alert.Resolved {
			delete(ceh.alerts, key)
		}
	}

	ceh.logger.Printf("Cleared resolved alerts")
}

// contains проверяет, содержит ли строка подстроку (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

// containsSubstring проверяет наличие подстроки в строке.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
