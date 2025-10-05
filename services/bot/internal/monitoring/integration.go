// Package monitoring provides integration of all monitoring components.
package monitoring

import (
	"context"
	"log"
	"os"
	"time"

	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/logging"
)

// MonitoringService интегрирует все компоненты мониторинга.
type MonitoringService struct {
	performanceMonitor *logging.PerformanceMonitor
	errorHandler       *errors.CentralizedErrorHandler
	dashboard          *Dashboard
	logger             *log.Logger
}

// NewMonitoringService создает новый сервис мониторинга.
func NewMonitoringService() *MonitoringService {
	performanceMonitor := logging.NewPerformanceMonitor()
	errorHandler := errors.NewCentralizedErrorHandler()
	dashboard := NewDashboard(performanceMonitor, errorHandler)

	return &MonitoringService{
		performanceMonitor: performanceMonitor,
		errorHandler:       errorHandler,
		dashboard:          dashboard,
		logger:             log.New(os.Stdout, "[MONITORING] ", log.LstdFlags),
	}
}

// Start запускает все компоненты мониторинга.
func (ms *MonitoringService) Start(ctx context.Context, dashboardPort int) error {
	ms.logger.Printf("Starting monitoring service...")

	// Запускаем дашборд в отдельной горутине
	go func() {
		if err := ms.dashboard.Start(dashboardPort); err != nil {
			ms.logger.Printf("Dashboard error: %v", err)
		}
	}()

	ms.logger.Printf("Monitoring service started successfully")
	ms.logger.Printf("Dashboard available at: http://localhost:%d", dashboardPort)

	return nil
}

// Stop останавливает все компоненты мониторинга.
func (ms *MonitoringService) Stop(ctx context.Context) error {
	ms.logger.Printf("Stopping monitoring service...")

	// Останавливаем дашборд
	if err := ms.dashboard.Stop(ctx); err != nil {
		ms.logger.Printf("Error stopping dashboard: %v", err)
	}

	ms.logger.Printf("Monitoring service stopped")
	return nil
}

// GetPerformanceMonitor возвращает монитор производительности.
func (ms *MonitoringService) GetPerformanceMonitor() *logging.PerformanceMonitor {
	return ms.performanceMonitor
}

// GetErrorHandler возвращает обработчик ошибок.
func (ms *MonitoringService) GetErrorHandler() *errors.CentralizedErrorHandler {
	return ms.errorHandler
}

// GetDashboard возвращает дашборд.
func (ms *MonitoringService) GetDashboard() *Dashboard {
	return ms.dashboard
}

// RecordOperation записывает операцию в мониторинг.
func (ms *MonitoringService) RecordOperation(requestID string, userID, chatID int64, operation, component string) *logging.RequestTrace {
	return ms.performanceMonitor.StartOperation(requestID, userID, chatID, operation, component)
}

// EndOperation завершает запись операции.
func (ms *MonitoringService) EndOperation(requestID string, status string, err error) {
	ms.performanceMonitor.EndOperation(requestID, status, err)
}

// RecordDatabaseOperation записывает операцию с базой данных.
func (ms *MonitoringService) RecordDatabaseOperation(requestID, operation string, duration time.Duration, err error) {
	ms.performanceMonitor.RecordDatabaseOperation(requestID, operation, duration, err)
}

// RecordCacheOperation записывает операцию с кэшем.
func (ms *MonitoringService) RecordCacheOperation(requestID, operation string, hit bool, duration time.Duration, err error) {
	ms.performanceMonitor.RecordCacheOperation(requestID, operation, hit, duration, err)
}

// RecordTelegramOperation записывает операцию с Telegram API.
func (ms *MonitoringService) RecordTelegramOperation(requestID, operation string, duration time.Duration, err error) {
	ms.performanceMonitor.RecordTelegramOperation(requestID, operation, duration, err)
}

// HandleError обрабатывает ошибку.
func (ms *MonitoringService) HandleError(ctx context.Context, err error, requestID string, userID, chatID int64, operation string) error {
	return ms.errorHandler.HandleError(ctx, err, requestID, userID, chatID, operation)
}

// HandleCustomError обрабатывает кастомную ошибку.
func (ms *MonitoringService) HandleCustomError(ctx context.Context, customErr *errors.CustomError, requestID string, userID, chatID int64, operation string) error {
	return ms.errorHandler.HandleCustomError(ctx, customErr, requestID, userID, chatID, operation)
}

// GetPerformanceReport возвращает отчет о производительности.
func (ms *MonitoringService) GetPerformanceReport() map[string]interface{} {
	return ms.performanceMonitor.GetPerformanceReport()
}

// GetActiveAlerts возвращает активные алерты.
func (ms *MonitoringService) GetActiveAlerts() map[string]*errors.Alert {
	return ms.errorHandler.GetActiveAlerts()
}

// LogPerformanceReport логирует отчет о производительности.
func (ms *MonitoringService) LogPerformanceReport() {
	ms.performanceMonitor.LogPerformanceReport()
}
