package cache

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Константы для метрик.
const (
	// secondsPerMinute - количество секунд в минуте для расчета запросов в секунду.
	secondsPerMinute = 60.0
)

// MetricsService сервис для сбора метрик кэша.
type MetricsService struct {
	cache ServiceInterface

	// Метрики производительности
	avgResponseTime time.Duration
	totalRequests   int64
	errorCount      int64

	// Метрики использования памяти
	memoryUsage    int64
	maxMemoryUsage int64

	// Метрики эффективности
	cacheEfficiency float64
	cleanupCount    int64
}

// NewMetricsService создает новый сервис метрик.
func NewMetricsService(cache ServiceInterface) *MetricsService {
	return &MetricsService{
		cache:           cache,
		avgResponseTime: 0,
		totalRequests:   0,
		errorCount:      0,
		memoryUsage:     0,
		maxMemoryUsage:  0,
		cacheEfficiency: 0,
		cleanupCount:    0,
	}
}

// RecordRequest записывает метрику запроса.
func (ms *MetricsService) RecordRequest(responseTime time.Duration, hit bool) {
	ms.totalRequests++
	ms.avgResponseTime = (ms.avgResponseTime*time.Duration(ms.totalRequests-1) + responseTime) /
		time.Duration(ms.totalRequests)

	if !hit {
		ms.errorCount++
	}

	// Обновляем эффективность кэша
	ms.updateCacheEfficiency()
}

// RecordError записывает ошибку.
func (ms *MetricsService) RecordError() {
	ms.errorCount++
}

// RecordCleanup записывает очистку кэша.
func (ms *MetricsService) RecordCleanup() {
	ms.cleanupCount++
}

// GetMetrics возвращает все метрики.
func (ms *MetricsService) GetMetrics() map[string]interface{} {
	stats := ms.cache.GetCacheStats(context.Background())

	return map[string]interface{}{
		"performance": map[string]interface{}{
			"total_requests":    ms.totalRequests,
			"avg_response_time": ms.avgResponseTime.String(),
			"error_count":       ms.errorCount,
			"error_rate":        ms.getErrorRate(),
		},
		"cache": map[string]interface{}{
			"hits":       stats.Hits,
			"misses":     stats.Misses,
			"hit_rate":   ms.getHitRate(stats),
			"size":       stats.Size,
			"efficiency": ms.cacheEfficiency,
		},
		"memory": map[string]interface{}{
			"current_usage": ms.memoryUsage,
			"max_usage":     ms.maxMemoryUsage,
		},
		"maintenance": map[string]interface{}{
			"cleanup_count": ms.cleanupCount,
			"last_cleanup":  time.Now().Format(time.RFC3339),
		},
	}
}

// GetPerformanceMetrics возвращает метрики производительности.
func (ms *MetricsService) GetPerformanceMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_requests":      ms.totalRequests,
		"avg_response_time":   ms.avgResponseTime.String(),
		"error_count":         ms.errorCount,
		"error_rate":          ms.getErrorRate(),
		"requests_per_second": ms.getRequestsPerSecond(),
	}
}

// GetCacheMetrics возвращает метрики кэша.
func (ms *MetricsService) GetCacheMetrics() map[string]interface{} {
	stats := ms.cache.GetCacheStats(context.Background())

	return map[string]interface{}{
		"hits":       stats.Hits,
		"misses":     stats.Misses,
		"hit_rate":   ms.getHitRate(stats),
		"size":       stats.Size,
		"efficiency": ms.cacheEfficiency,
	}
}

// GetMemoryMetrics возвращает метрики памяти.
func (ms *MetricsService) GetMemoryMetrics() map[string]interface{} {
	return map[string]interface{}{
		"current_usage": ms.memoryUsage,
		"max_usage":     ms.maxMemoryUsage,
		"usage_percent": ms.getMemoryUsagePercent(),
	}
}

// GetMaintenanceMetrics возвращает метрики обслуживания.
func (ms *MetricsService) GetMaintenanceMetrics() map[string]interface{} {
	return map[string]interface{}{
		"cleanup_count":     ms.cleanupCount,
		"last_cleanup":      time.Now().Format(time.RFC3339),
		"cleanup_frequency": ms.getCleanupFrequency(),
	}
}

// LogMetrics выводит метрики в лог.
func (ms *MetricsService) LogMetrics() {
	metrics := ms.GetMetrics()

	log.Printf("=== Cache Metrics ===")
	log.Printf("Performance: %+v", metrics["performance"])
	log.Printf("Cache: %+v", metrics["cache"])
	log.Printf("Memory: %+v", metrics["memory"])
	log.Printf("Maintenance: %+v", metrics["maintenance"])
}

// GetMetricsSummary возвращает краткую сводку метрик.
func (ms *MetricsService) GetMetricsSummary() string {
	stats := ms.cache.GetCacheStats(context.Background())
	hitRate := ms.getHitRate(stats)
	errorRate := ms.getErrorRate()

	return fmt.Sprintf("Cache: %d/%d hits (%.1f%%), %d errors (%.1f%%), %d entries, %s avg response",
		stats.Hits, stats.Hits+stats.Misses, hitRate,
		ms.errorCount, errorRate, stats.Size, ms.avgResponseTime)
}

// ResetMetrics сбрасывает все метрики.
func (ms *MetricsService) ResetMetrics() {
	ms.avgResponseTime = 0
	ms.totalRequests = 0
	ms.errorCount = 0
	ms.memoryUsage = 0
	ms.maxMemoryUsage = 0
	ms.cacheEfficiency = 0
	ms.cleanupCount = 0

	log.Printf("Metrics: All metrics reset")
}

// updateCacheEfficiency обновляет эффективность кэша.
func (ms *MetricsService) updateCacheEfficiency() {
	stats := ms.cache.GetCacheStats(context.Background())

	total := stats.Hits + stats.Misses

	if total > 0 {
		ms.cacheEfficiency = float64(stats.Hits) / float64(total) * PercentageMultiplier
	}
}

// getHitRate возвращает процент попаданий в кэш.
func (ms *MetricsService) getHitRate(stats Stats) float64 {
	total := stats.Hits + stats.Misses
	if total == 0 {
		return 0
	}

	return float64(stats.Hits) / float64(total) * PercentageMultiplier
}

// getErrorRate возвращает процент ошибок.
func (ms *MetricsService) getErrorRate() float64 {
	if ms.totalRequests == 0 {
		return 0
	}

	return float64(ms.errorCount) / float64(ms.totalRequests) * PercentageMultiplier
}

// getRequestsPerSecond возвращает количество запросов в секунду.
func (ms *MetricsService) getRequestsPerSecond() float64 {
	// Упрощенный расчет - в реальности нужно учитывать временные интервалы
	if ms.totalRequests == 0 {
		return 0
	}

	return float64(ms.totalRequests) / secondsPerMinute // Предполагаем 1 минуту работы
}

// getMemoryUsagePercent возвращает процент использования памяти.
func (ms *MetricsService) getMemoryUsagePercent() float64 {
	if ms.maxMemoryUsage == 0 {
		return 0
	}

	return float64(ms.memoryUsage) / float64(ms.maxMemoryUsage) * PercentageMultiplier
}

// getCleanupFrequency возвращает частоту очистки.
func (ms *MetricsService) getCleanupFrequency() string {
	if ms.cleanupCount == 0 {
		return "No cleanups yet"
	}

	return fmt.Sprintf("%d cleanups", ms.cleanupCount)
}
