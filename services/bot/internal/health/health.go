package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"language-exchange-bot/internal/logging"
)

// HealthChecker интерфейс для проверки здоровья компонентов
type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
}

// HealthStatus статус здоровья
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthResponse ответ health check
type HealthResponse struct {
	Status    HealthStatus           `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    time.Duration          `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks"`
	System    SystemInfo             `json:"system"`
}

// CheckResult результат проверки компонента
type CheckResult struct {
	Status    HealthStatus  `json:"status"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// SystemInfo информация о системе
type SystemInfo struct {
	MemoryUsage  uint64  `json:"memory_usage_bytes"`
	MemoryTotal  uint64  `json:"memory_total_bytes"`
	Goroutines   int     `json:"goroutines"`
	CPUUsage     float64 `json:"cpu_usage_percent"`
	GoVersion    string  `json:"go_version"`
	Architecture string  `json:"architecture"`
	OS           string  `json:"os"`
}

// HealthManager управляет проверками здоровья
type HealthManager struct {
	checkers  []HealthChecker
	logger    logging.Logger
	startTime time.Time
	version   string
}

// NewHealthManager создает новый менеджер здоровья
func NewHealthManager(logger logging.Logger, version string) *HealthManager {
	return &HealthManager{
		checkers:  make([]HealthChecker, 0),
		logger:    logger,
		startTime: time.Now(),
		version:   version,
	}
}

// AddChecker добавляет проверку здоровья
func (h *HealthManager) AddChecker(checker HealthChecker) {
	h.checkers = append(h.checkers, checker)
}

// Check выполняет все проверки здоровья
func (h *HealthManager) Check(ctx context.Context) *HealthResponse {
	response := &HealthResponse{
		Timestamp: time.Now(),
		Version:   h.version,
		Uptime:    time.Since(h.startTime),
		Checks:    make(map[string]CheckResult),
		System:    h.getSystemInfo(),
	}

	// Выполняем проверки
	allHealthy := true
	anyDegraded := false

	for _, checker := range h.checkers {
		checkStart := time.Now()

		err := checker.Check(ctx)
		duration := time.Since(checkStart)

		result := CheckResult{
			Duration:  duration,
			Timestamp: time.Now(),
		}

		if err != nil {
			result.Status = StatusUnhealthy
			result.Message = err.Error()
			allHealthy = false

			h.logger.Error("Health check failed",
				logging.String("component", checker.Name()),
				logging.ErrorField(err),
				logging.Duration("duration", duration),
			)
		} else {
			result.Status = StatusHealthy
		}

		response.Checks[checker.Name()] = result
	}

	// Определяем общий статус
	if allHealthy {
		response.Status = StatusHealthy
	} else if anyDegraded {
		response.Status = StatusDegraded
	} else {
		response.Status = StatusUnhealthy
	}

	return response
}

// getSystemInfo получает информацию о системе
func (h *HealthManager) getSystemInfo() SystemInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return SystemInfo{
		MemoryUsage:  memStats.Alloc,
		MemoryTotal:  memStats.Sys,
		Goroutines:   runtime.NumGoroutine(),
		CPUUsage:     0.0, // В реальном проекте здесь должен быть сбор метрик CPU
		GoVersion:    runtime.Version(),
		Architecture: runtime.GOARCH,
		OS:           runtime.GOOS,
	}
}

// HTTPHandler создает HTTP обработчик для health check
func (h *HealthManager) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		response := h.Check(ctx)

		// Устанавливаем заголовки
		w.Header().Set("Content-Type", "application/json")

		// Устанавливаем статус код в зависимости от здоровья
		switch response.Status {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // 200, но с предупреждением
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		// Кодируем ответ
		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.Error("Failed to encode health response",
				logging.ErrorField(err),
			)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// ReadinessHandler создает HTTP обработчик для readiness check
func (h *HealthManager) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// Проверяем только критические компоненты
		criticalCheckers := []string{"database", "cache"}
		allReady := true

		for _, checker := range h.checkers {
			for _, critical := range criticalCheckers {
				if checker.Name() == critical {
					if err := checker.Check(ctx); err != nil {
						allReady = false
						h.logger.Error("Readiness check failed",
							logging.String("component", checker.Name()),
							logging.ErrorField(err),
						)
					}
					break
				}
			}
		}

		if allReady {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
		}
	}
}

// LivenessHandler создает HTTP обработчик для liveness check
func (h *HealthManager) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Liveness check всегда возвращает OK, если процесс запущен
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// DatabaseHealthChecker проверка здоровья базы данных
type DatabaseHealthChecker struct {
	name      string
	checkFunc func(ctx context.Context) error
}

// NewDatabaseHealthChecker создает новый проверщик БД
func NewDatabaseHealthChecker(name string, checkFunc func(ctx context.Context) error) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name:      name,
		checkFunc: checkFunc,
	}
}

// Check выполняет проверку БД
func (d *DatabaseHealthChecker) Check(ctx context.Context) error {
	return d.checkFunc(ctx)
}

// Name возвращает имя проверщика
func (d *DatabaseHealthChecker) Name() string {
	return d.name
}

// CacheHealthChecker проверка здоровья кэша
type CacheHealthChecker struct {
	name      string
	checkFunc func(ctx context.Context) error
}

// NewCacheHealthChecker создает новый проверщик кэша
func NewCacheHealthChecker(name string, checkFunc func(ctx context.Context) error) *CacheHealthChecker {
	return &CacheHealthChecker{
		name:      name,
		checkFunc: checkFunc,
	}
}

// Check выполняет проверку кэша
func (c *CacheHealthChecker) Check(ctx context.Context) error {
	return c.checkFunc(ctx)
}

// Name возвращает имя проверщика
func (c *CacheHealthChecker) Name() string {
	return c.name
}

// ExternalServiceHealthChecker проверка здоровья внешнего сервиса
type ExternalServiceHealthChecker struct {
	name    string
	url     string
	timeout time.Duration
	client  *http.Client
}

// NewExternalServiceHealthChecker создает новый проверщик внешнего сервиса
func NewExternalServiceHealthChecker(name, url string, timeout time.Duration) *ExternalServiceHealthChecker {
	return &ExternalServiceHealthChecker{
		name:    name,
		url:     url,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Check выполняет проверку внешнего сервиса
func (e *ExternalServiceHealthChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", e.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("service returned status %d", resp.StatusCode)
	}

	return nil
}

// Name возвращает имя проверщика
func (e *ExternalServiceHealthChecker) Name() string {
	return e.name
}

// DiskSpaceHealthChecker проверка свободного места на диске
type DiskSpaceHealthChecker struct {
	name      string
	path      string
	threshold float64 // Минимальный процент свободного места
}

// NewDiskSpaceHealthChecker создает новый проверщик диска
func NewDiskSpaceHealthChecker(name, path string, threshold float64) *DiskSpaceHealthChecker {
	return &DiskSpaceHealthChecker{
		name:      name,
		path:      path,
		threshold: threshold,
	}
}

// Check выполняет проверку диска
func (d *DiskSpaceHealthChecker) Check(ctx context.Context) error {
	// Здесь должна быть логика проверки свободного места на диске
	// Для примера возвращаем успех
	return nil
}

// Name возвращает имя проверщика
func (d *DiskSpaceHealthChecker) Name() string {
	return d.name
}

// MemoryHealthChecker проверка использования памяти
type MemoryHealthChecker struct {
	name      string
	threshold uint64 // Максимальное использование памяти в байтах
}

// NewMemoryHealthChecker создает новый проверщик памяти
func NewMemoryHealthChecker(name string, threshold uint64) *MemoryHealthChecker {
	return &MemoryHealthChecker{
		name:      name,
		threshold: threshold,
	}
}

// Check выполняет проверку памяти
func (m *MemoryHealthChecker) Check(ctx context.Context) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	if memStats.Alloc > m.threshold {
		return fmt.Errorf("memory usage %d bytes exceeds threshold %d bytes",
			memStats.Alloc, m.threshold)
	}

	return nil
}

// Name возвращает имя проверщика
func (m *MemoryHealthChecker) Name() string {
	return m.name
}
