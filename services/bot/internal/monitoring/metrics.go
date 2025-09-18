package monitoring

import (
	"context"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics структура для метрик Prometheus
type Metrics struct {
	// HTTP метрики
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Database метрики
	DBConnectionsActive prometheus.Gauge
	DBConnectionsIdle   prometheus.Gauge
	DBQueriesTotal      *prometheus.CounterVec
	DBQueryDuration     *prometheus.HistogramVec
	DBTransactionsTotal *prometheus.CounterVec

	// Cache метрики
	CacheHitsTotal         *prometheus.CounterVec
	CacheMissesTotal       *prometheus.CounterVec
	CacheOperationsTotal   *prometheus.CounterVec
	CacheOperationDuration *prometheus.HistogramVec

	// Business метрики
	UsersRegistered   prometheus.Counter
	UsersActive       prometheus.Gauge
	ProfilesCompleted prometheus.Counter
	FeedbackSubmitted prometheus.Counter
	FeedbackProcessed prometheus.Counter

	// System метрики
	MemoryUsage     prometheus.Gauge
	CPUUsage        prometheus.Gauge
	GoroutinesCount prometheus.Gauge
}

// NewMetrics создает новый экземпляр метрик
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP метрики
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		// Database метрики
		DBConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_idle",
				Help: "Number of idle database connections",
			},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"operation", "table"},
		),
		DBTransactionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_transactions_total",
				Help: "Total number of database transactions",
			},
			[]string{"status"},
		),

		// Cache метрики
		CacheHitsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type", "key_pattern"},
		),
		CacheMissesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type", "key_pattern"},
		),
		CacheOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_operations_total",
				Help: "Total number of cache operations",
			},
			[]string{"operation", "cache_type", "status"},
		),
		CacheOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "cache_operation_duration_seconds",
				Help:    "Cache operation duration in seconds",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"operation", "cache_type"},
		),

		// Business метрики
		UsersRegistered: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "users_registered_total",
				Help: "Total number of registered users",
			},
		),
		UsersActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "users_active",
				Help: "Number of active users",
			},
		),
		ProfilesCompleted: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "profiles_completed_total",
				Help: "Total number of completed profiles",
			},
		),
		FeedbackSubmitted: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "feedback_submitted_total",
				Help: "Total number of feedback submissions",
			},
		),
		FeedbackProcessed: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "feedback_processed_total",
				Help: "Total number of processed feedback",
			},
		),

		// System метрики
		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),
		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "Current CPU usage percentage",
			},
		),
		GoroutinesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "goroutines_count",
				Help: "Current number of goroutines",
			},
		),
	}
}

// RecordHTTPRequest записывает метрику HTTP запроса
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordDBQuery записывает метрику запроса к БД
func (m *Metrics) RecordDBQuery(operation, table, status string, duration time.Duration) {
	m.DBQueriesTotal.WithLabelValues(operation, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordDBTransaction записывает метрику транзакции БД
func (m *Metrics) RecordDBTransaction(status string) {
	m.DBTransactionsTotal.WithLabelValues(status).Inc()
}

// RecordCacheOperation записывает метрику операции с кэшем
func (m *Metrics) RecordCacheOperation(operation, cacheType, keyPattern, status string, duration time.Duration) {
	m.CacheOperationsTotal.WithLabelValues(operation, cacheType, status).Inc()
	m.CacheOperationDuration.WithLabelValues(operation, cacheType).Observe(duration.Seconds())
}

// RecordCacheHit записывает метрику попадания в кэш
func (m *Metrics) RecordCacheHit(cacheType, keyPattern string) {
	m.CacheHitsTotal.WithLabelValues(cacheType, keyPattern).Inc()
}

// RecordCacheMiss записывает метрику промаха кэша
func (m *Metrics) RecordCacheMiss(cacheType, keyPattern string) {
	m.CacheMissesTotal.WithLabelValues(cacheType, keyPattern).Inc()
}

// RecordUserRegistration записывает метрику регистрации пользователя
func (m *Metrics) RecordUserRegistration() {
	m.UsersRegistered.Inc()
}

// RecordProfileCompletion записывает метрику завершения профиля
func (m *Metrics) RecordProfileCompletion() {
	m.ProfilesCompleted.Inc()
}

// RecordFeedbackSubmission записывает метрику отправки отзыва
func (m *Metrics) RecordFeedbackSubmission() {
	m.FeedbackSubmitted.Inc()
}

// RecordFeedbackProcessing записывает метрику обработки отзыва
func (m *Metrics) RecordFeedbackProcessing() {
	m.FeedbackProcessed.Inc()
}

// UpdateDBConnections обновляет метрики соединений с БД
func (m *Metrics) UpdateDBConnections(active, idle int) {
	m.DBConnectionsActive.Set(float64(active))
	m.DBConnectionsIdle.Set(float64(idle))
}

// UpdateSystemMetrics обновляет системные метрики
func (m *Metrics) UpdateSystemMetrics(memoryBytes uint64, cpuPercent float64, goroutines int) {
	m.MemoryUsage.Set(float64(memoryBytes))
	m.CPUUsage.Set(cpuPercent)
	m.GoroutinesCount.Set(float64(goroutines))
}

// UpdateActiveUsers обновляет метрику активных пользователей
func (m *Metrics) UpdateActiveUsers(count int) {
	m.UsersActive.Set(float64(count))
}

// MetricsCollector интерфейс для сбора метрик
type MetricsCollector interface {
	Collect(ctx context.Context) error
}

// SystemMetricsCollector собирает системные метрики
type SystemMetricsCollector struct {
	metrics *Metrics
}

// NewSystemMetricsCollector создает новый сборщик системных метрик
func NewSystemMetricsCollector(metrics *Metrics) *SystemMetricsCollector {
	return &SystemMetricsCollector{
		metrics: metrics,
	}
}

// Collect собирает системные метрики
func (c *SystemMetricsCollector) Collect(ctx context.Context) error {
	// Здесь должна быть логика сбора системных метрик
	// Для примера используем заглушки

	// Сбор метрик памяти
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Сбор метрик CPU (упрощенная версия)
	cpuPercent := 0.0 // В реальном проекте здесь должен быть сбор метрик CPU

	// Сбор количества горутин
	goroutines := runtime.NumGoroutine()

	c.metrics.UpdateSystemMetrics(memStats.Alloc, cpuPercent, goroutines)

	return nil
}

// DBMetricsCollector собирает метрики БД
type DBMetricsCollector struct {
	metrics *Metrics
	db      interface{} // Интерфейс для получения статистики БД
}

// NewDBMetricsCollector создает новый сборщик метрик БД
func NewDBMetricsCollector(metrics *Metrics, db interface{}) *DBMetricsCollector {
	return &DBMetricsCollector{
		metrics: metrics,
		db:      db,
	}
}

// Collect собирает метрики БД
func (c *DBMetricsCollector) Collect(ctx context.Context) error {
	// Здесь должна быть логика сбора метрик БД
	// Для примера используем заглушки

	// Получение статистики соединений
	active := 10 // Заглушка
	idle := 5    // Заглушка

	c.metrics.UpdateDBConnections(active, idle)

	return nil
}

// MetricsManager управляет сбором метрик
type MetricsManager struct {
	collectors []MetricsCollector
	interval   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewMetricsManager создает новый менеджер метрик
func NewMetricsManager(interval time.Duration) *MetricsManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &MetricsManager{
		collectors: make([]MetricsCollector, 0),
		interval:   interval,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// AddCollector добавляет сборщик метрик
func (m *MetricsManager) AddCollector(collector MetricsCollector) {
	m.collectors = append(m.collectors, collector)
}

// Start запускает сбор метрик
func (m *MetricsManager) Start() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			for _, collector := range m.collectors {
				if err := collector.Collect(m.ctx); err != nil {
					// Логируем ошибку, но не останавливаем сбор метрик
					continue
				}
			}
		}
	}
}

// Stop останавливает сбор метрик
func (m *MetricsManager) Stop() {
	m.cancel()
}
