// Package logging provides performance metrics and monitoring.
package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// MetricType определяет тип метрики.
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeTimer     MetricType = "timer"
)

// Metric представляет отдельную метрику.
type Metric struct {
	Name      string                 `json:"name"`
	Type      MetricType             `json:"type"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// MetricCollector собирает и хранит метрики.
type MetricCollector struct {
	metrics map[string]*Metric
	mutex   sync.RWMutex
	logger  *log.Logger
}

// NewMetricCollector создает новый сборщик метрик.
func NewMetricCollector() *MetricCollector {
	return &MetricCollector{
		metrics: make(map[string]*Metric),
		logger:  log.New(os.Stdout, "[METRICS] ", log.LstdFlags),
	}
}

// IncrementCounter увеличивает счетчик.
func (mc *MetricCollector) IncrementCounter(name string, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.getMetricKey(name, labels)

	metric, exists := mc.metrics[key]
	if !exists {
		metric = &Metric{
			Name:      name,
			Type:      MetricTypeCounter,
			Value:     0,
			Labels:    labels,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}
		mc.metrics[key] = metric
	}

	metric.Value++
	metric.Timestamp = time.Now()
}

// SetGauge устанавливает значение gauge.
func (mc *MetricCollector) SetGauge(name string, value float64, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.getMetricKey(name, labels)
	metric := &Metric{
		Name:      name,
		Type:      MetricTypeGauge,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	mc.metrics[key] = metric
}

// RecordHistogram записывает значение в гистограмму.
func (mc *MetricCollector) RecordHistogram(name string, value float64, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.getMetricKey(name, labels)

	metric, exists := mc.metrics[key]
	if !exists {
		metric = &Metric{
			Name:      name,
			Type:      MetricTypeHistogram,
			Value:     0,
			Labels:    labels,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}
		mc.metrics[key] = metric
	}

	// Для гистограммы храним статистику
	if metric.Metadata["count"] == nil {
		metric.Metadata["count"] = 0
		metric.Metadata["sum"] = 0.0
		metric.Metadata["min"] = value
		metric.Metadata["max"] = value
	}

	metric.Metadata["count"] = metric.Metadata["count"].(int) + 1
	metric.Metadata["sum"] = metric.Metadata["sum"].(float64) + value

	if value < metric.Metadata["min"].(float64) {
		metric.Metadata["min"] = value
	}

	if value > metric.Metadata["max"].(float64) {
		metric.Metadata["max"] = value
	}

	metric.Value = metric.Metadata["sum"].(float64) / float64(metric.Metadata["count"].(int))
	metric.Timestamp = time.Now()
}

// RecordTimer записывает время выполнения операции.
func (mc *MetricCollector) RecordTimer(name string, duration time.Duration, labels map[string]string) {
	mc.RecordHistogram(name, float64(duration.Milliseconds()), labels)
}

// GetMetric возвращает метрику по имени и лейблам.
func (mc *MetricCollector) GetMetric(name string, labels map[string]string) (*Metric, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	key := mc.getMetricKey(name, labels)
	metric, exists := mc.metrics[key]

	return metric, exists
}

// GetAllMetrics возвращает все метрики.
func (mc *MetricCollector) GetAllMetrics() map[string]*Metric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	// Создаем копию для безопасности
	result := make(map[string]*Metric)
	for key, metric := range mc.metrics {
		result[key] = metric
	}

	return result
}

// GetMetricsByType возвращает метрики определенного типа.
func (mc *MetricCollector) GetMetricsByType(metricType MetricType) map[string]*Metric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	result := make(map[string]*Metric)

	for key, metric := range mc.metrics {
		if metric.Type == metricType {
			result[key] = metric
		}
	}

	return result
}

// ClearMetrics очищает все метрики.
func (mc *MetricCollector) ClearMetrics() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.metrics = make(map[string]*Metric)
}

// ExportMetrics экспортирует метрики в JSON.
func (mc *MetricCollector) ExportMetrics() ([]byte, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return json.Marshal(mc.metrics)
}

// LogMetrics логирует все метрики.
func (mc *MetricCollector) LogMetrics() {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	for key, metric := range mc.metrics {
		metricData := map[string]interface{}{
			"metric_key": key,
			"metric":     metric,
			"timestamp":  time.Now().Format(time.RFC3339),
		}

		jsonData, err := json.Marshal(metricData)
		if err != nil {
			mc.logger.Printf("Failed to marshal metric data: %v", err)

			continue
		}

		mc.logger.Printf("%s", string(jsonData))
	}
}

// getMetricKey создает ключ для метрики.
func (mc *MetricCollector) getMetricKey(name string, labels map[string]string) string {
	key := name
	for labelKey, labelValue := range labels {
		key += fmt.Sprintf("_%s_%s", labelKey, labelValue)
	}

	return key
}

// PerformanceMonitor предоставляет мониторинг производительности.
type PerformanceMonitor struct {
	collector *MetricCollector
	tracing   *TracingService
	logger    *log.Logger
}

// NewPerformanceMonitor создает новый монитор производительности.
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		collector: NewMetricCollector(),
		tracing:   NewTracingService(),
		logger:    log.New(os.Stdout, "[PERFORMANCE] ", log.LstdFlags),
	}
}

// StartOperation начинает мониторинг операции.
func (pm *PerformanceMonitor) StartOperation(requestID string, userID, chatID int64, operation, component string) *RequestTrace {
	return pm.tracing.StartTrace(requestID, userID, chatID, operation, component)
}

// EndOperation завершает мониторинг операции.
func (pm *PerformanceMonitor) EndOperation(requestID string, status string, err error) {
	pm.tracing.EndTrace(requestID, status, err)
}

// RecordDatabaseOperation записывает операцию с базой данных.
func (pm *PerformanceMonitor) RecordDatabaseOperation(requestID, operation string, duration time.Duration, err error) {
	pm.tracing.RecordDatabaseOperation(requestID, operation, duration, err)

	labels := map[string]string{
		"operation": operation,
		"status":    "success",
	}
	if err != nil {
		labels["status"] = "error"
	}

	pm.collector.RecordTimer("database_operation_duration", duration, labels)
	pm.collector.IncrementCounter("database_operations_total", labels)
}

// RecordCacheOperation записывает операцию с кэшем.
func (pm *PerformanceMonitor) RecordCacheOperation(requestID, operation string, hit bool, duration time.Duration, err error) {
	pm.tracing.RecordCacheOperation(requestID, operation, hit, duration, err)

	labels := map[string]string{
		"operation": operation,
		"hit":       strconv.FormatBool(hit),
		"status":    "success",
	}
	if err != nil {
		labels["status"] = "error"
	}

	pm.collector.RecordTimer("cache_operation_duration", duration, labels)
	pm.collector.IncrementCounter("cache_operations_total", labels)
}

// RecordTelegramOperation записывает операцию с Telegram API.
func (pm *PerformanceMonitor) RecordTelegramOperation(requestID, operation string, duration time.Duration, err error) {
	pm.tracing.RecordTelegramOperation(requestID, operation, duration, err)

	labels := map[string]string{
		"operation": operation,
		"status":    "success",
	}
	if err != nil {
		labels["status"] = "error"
	}

	pm.collector.RecordTimer("telegram_operation_duration", duration, labels)
	pm.collector.IncrementCounter("telegram_operations_total", labels)
}

// GetPerformanceReport возвращает отчет о производительности.
func (pm *PerformanceMonitor) GetPerformanceReport() map[string]interface{} {
	summary := pm.tracing.GetPerformanceSummary()
	metrics := pm.collector.GetAllMetrics()

	return map[string]interface{}{
		"summary":   summary,
		"metrics":   metrics,
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

// LogPerformanceReport логирует отчет о производительности.
func (pm *PerformanceMonitor) LogPerformanceReport() {
	report := pm.GetPerformanceReport()

	jsonData, err := json.Marshal(report)
	if err != nil {
		pm.logger.Printf("Failed to marshal performance report: %v", err)

		return
	}

	pm.logger.Printf("PERFORMANCE_REPORT: %s", string(jsonData))
}
