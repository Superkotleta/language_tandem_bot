// Package logging provides request tracing and performance monitoring.
package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Константы для трассировки.
const (
	// maxTracingMetrics - максимальное количество метрик для хранения.
	maxTracingMetrics = 1000
)

// traceContextKey тип ключа для контекста трейса.
type traceContextKey string

// RequestTrace представляет трейс запроса.
type RequestTrace struct {
	RequestID string                 `json:"requestId"`
	UserID    int64                  `json:"userId"`
	ChatID    int64                  `json:"chatId"`
	Operation string                 `json:"operation"`
	Component string                 `json:"component"`
	StartTime time.Time              `json:"startTime"`
	EndTime   time.Time              `json:"endTime"`
	Duration  time.Duration          `json:"durationMs"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	SubTraces []*RequestTrace        `json:"subTraces,omitempty"`
}

// PerformanceMetrics представляет метрики производительности.
type PerformanceMetrics struct {
	Operation    string        `json:"operation"`
	Component    string        `json:"component"`
	Duration     time.Duration `json:"durationMs"`
	MemoryUsage  int64         `json:"memoryBytes"`
	DatabaseHits int           `json:"databaseHits"`
	CacheHits    int           `json:"cacheHits"`
	CacheMisses  int           `json:"cacheMisses"`
	ErrorCount   int           `json:"errorCount"`
	SuccessCount int           `json:"successCount"`
	Timestamp    time.Time     `json:"timestamp"`
}

// TracingService предоставляет функциональность трейсинга.
type TracingService struct {
	activeTraces map[string]*RequestTrace
	metrics      []PerformanceMetrics
	logger       *log.Logger
}

// NewTracingService создает новый сервис трейсинга.
func NewTracingService() *TracingService {
	return &TracingService{
		activeTraces: make(map[string]*RequestTrace),
		metrics:      make([]PerformanceMetrics, 0),
		logger:       log.New(os.Stdout, "[TRACING] ", log.LstdFlags),
	}
}

// StartTrace начинает новый трейс запроса.
func (ts *TracingService) StartTrace(requestID string, userID, chatID int64, operation, component string) *RequestTrace {
	trace := &RequestTrace{
		RequestID: requestID,
		UserID:    userID,
		ChatID:    chatID,
		Operation: operation,
		Component: component,
		StartTime: time.Now(),
		Status:    "started",
		Metadata:  make(map[string]interface{}),
		SubTraces: make([]*RequestTrace, 0),
	}

	ts.activeTraces[requestID] = trace
	ts.logTrace(trace, "TRACE_START")

	return trace
}

// EndTrace завершает трейс запроса.
func (ts *TracingService) EndTrace(requestID string, status string, err error) {
	trace, exists := ts.activeTraces[requestID]
	if !exists {
		return
	}

	trace.EndTime = time.Now()
	trace.Duration = trace.EndTime.Sub(trace.StartTime)
	trace.Status = status

	if err != nil {
		trace.Error = err.Error()
	}

	ts.logTrace(trace, "TRACE_END")

	// Записываем метрики
	ts.recordMetrics(trace)

	// Удаляем из активных трейсов
	delete(ts.activeTraces, requestID)
}

// AddSubTrace добавляет под-трейс к основному трейсу.
func (ts *TracingService) AddSubTrace(parentRequestID, subOperation, subComponent string) *RequestTrace {
	parentTrace, exists := ts.activeTraces[parentRequestID]
	if !exists {
		return nil
	}

	subTrace := &RequestTrace{
		RequestID: fmt.Sprintf("%s_%s", parentRequestID, subOperation),
		UserID:    parentTrace.UserID,
		ChatID:    parentTrace.ChatID,
		Operation: subOperation,
		Component: subComponent,
		StartTime: time.Now(),
		Status:    "started",
		Metadata:  make(map[string]interface{}),
		SubTraces: make([]*RequestTrace, 0),
	}

	parentTrace.SubTraces = append(parentTrace.SubTraces, subTrace)

	return subTrace
}

// AddMetadata добавляет метаданные к трейсу.
func (ts *TracingService) AddMetadata(requestID string, key string, value interface{}) {
	trace, exists := ts.activeTraces[requestID]
	if !exists {
		return
	}

	trace.Metadata[key] = value
}

// RecordDatabaseOperation записывает операцию с базой данных.
func (ts *TracingService) RecordDatabaseOperation(requestID, operation string, duration time.Duration, err error) {
	ts.AddMetadata(requestID, "db_operation", operation)
	ts.AddMetadata(requestID, "db_duration_ms", duration.Milliseconds())

	if err != nil {
		ts.AddMetadata(requestID, "db_error", err.Error())
	}
}

// RecordCacheOperation записывает операцию с кэшем.
func (ts *TracingService) RecordCacheOperation(requestID, operation string, hit bool, duration time.Duration, err error) {
	ts.AddMetadata(requestID, "cache_operation", operation)
	ts.AddMetadata(requestID, "cache_hit", hit)
	ts.AddMetadata(requestID, "cache_duration_ms", duration.Milliseconds())

	if err != nil {
		ts.AddMetadata(requestID, "cache_error", err.Error())
	}
}

// RecordTelegramOperation записывает операцию с Telegram API.
func (ts *TracingService) RecordTelegramOperation(requestID, operation string, duration time.Duration, err error) {
	ts.AddMetadata(requestID, "telegram_operation", operation)
	ts.AddMetadata(requestID, "telegram_duration_ms", duration.Milliseconds())

	if err != nil {
		ts.AddMetadata(requestID, "telegram_error", err.Error())
	}
}

// logTrace логирует трейс в структурированном формате.
func (ts *TracingService) logTrace(trace *RequestTrace, event string) {
	traceData := map[string]interface{}{
		"event":     event,
		"trace":     trace,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(traceData)
	if err != nil {
		ts.logger.Printf("Failed to marshal trace data: %v", err)

		return
	}

	ts.logger.Printf("%s", string(jsonData))
}

// recordMetrics записывает метрики производительности.
func (ts *TracingService) recordMetrics(trace *RequestTrace) {
	metrics := PerformanceMetrics{
		Operation:    trace.Operation,
		Component:    trace.Component,
		Duration:     trace.Duration,
		Timestamp:    time.Now(),
		SuccessCount: 1,
	}

	if trace.Error != "" {
		metrics.ErrorCount = 1
		metrics.SuccessCount = 0
	}

	// Извлекаем метрики из метаданных
	if dbHits, ok := trace.Metadata["db_hits"].(int); ok {
		metrics.DatabaseHits = dbHits
	}

	if cacheHits, ok := trace.Metadata["cache_hits"].(int); ok {
		metrics.CacheHits = cacheHits
	}

	if cacheMisses, ok := trace.Metadata["cache_misses"].(int); ok {
		metrics.CacheMisses = cacheMisses
	}

	ts.metrics = append(ts.metrics, metrics)

	// Ограничиваем размер массива метрик
	if len(ts.metrics) > maxTracingMetrics {
		ts.metrics = ts.metrics[len(ts.metrics)-maxTracingMetrics:]
	}
}

// GetMetrics возвращает метрики производительности.
func (ts *TracingService) GetMetrics() []PerformanceMetrics {
	return ts.metrics
}

// GetActiveTraces возвращает активные трейсы.
func (ts *TracingService) GetActiveTraces() map[string]*RequestTrace {
	return ts.activeTraces
}

// GetTraceByRequestID возвращает трейс по ID запроса.
func (ts *TracingService) GetTraceByRequestID(requestID string) (*RequestTrace, bool) {
	trace, exists := ts.activeTraces[requestID]

	return trace, exists
}

// ClearMetrics очищает метрики.
func (ts *TracingService) ClearMetrics() {
	ts.metrics = make([]PerformanceMetrics, 0)
}

// GetPerformanceSummary возвращает сводку по производительности.
func (ts *TracingService) GetPerformanceSummary() map[string]interface{} {
	if len(ts.metrics) == 0 {
		return map[string]interface{}{
			"total_operations": 0,
			"average_duration": 0,
			"error_rate":       0,
		}
	}

	var (
		totalDuration                                 time.Duration
		errorCount, successCount                      int
		totalDbHits, totalCacheHits, totalCacheMisses int
	)

	for _, metric := range ts.metrics {
		totalDuration += metric.Duration
		errorCount += metric.ErrorCount
		successCount += metric.SuccessCount
		totalDbHits += metric.DatabaseHits
		totalCacheHits += metric.CacheHits
		totalCacheMisses += metric.CacheMisses
	}

	totalOperations := len(ts.metrics)
	avgDuration := totalDuration / time.Duration(totalOperations)
	errorRate := float64(errorCount) / float64(totalOperations) * 100

	return map[string]interface{}{
		"total_operations":       totalOperations,
		"average_duration_ms":    avgDuration.Milliseconds(),
		"error_rate_percent":     errorRate,
		"total_db_hits":          totalDbHits,
		"total_cache_hits":       totalCacheHits,
		"total_cache_misses":     totalCacheMisses,
		"cache_hit_rate_percent": float64(totalCacheHits) / float64(totalCacheHits+totalCacheMisses) * 100,
	}
}

// ContextWithTrace добавляет трейс в контекст.
func ContextWithTrace(ctx context.Context, trace *RequestTrace) context.Context {
	return context.WithValue(ctx, traceContextKey("trace"), trace)
}

// TraceFromContext извлекает трейс из контекста.
func TraceFromContext(ctx context.Context) (*RequestTrace, bool) {
	trace, ok := ctx.Value(traceContextKey("trace")).(*RequestTrace)

	return trace, ok
}
