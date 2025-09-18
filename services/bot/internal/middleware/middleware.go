package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"language-exchange-bot/internal/logging"
	"language-exchange-bot/internal/monitoring"

	"github.com/prometheus/client_golang/prometheus"
)

// RequestIDKey ключ для request ID в контексте
type RequestIDKey struct{}

// UserIDKey ключ для user ID в контексте
type UserIDKey struct{}

// TraceIDKey ключ для trace ID в контексте
type TraceIDKey struct{}

// Middleware интерфейс для middleware
type Middleware interface {
	Handle(next http.Handler) http.Handler
}

// LoggingMiddleware middleware для логирования
type LoggingMiddleware struct {
	logger logging.Logger
}

// NewLoggingMiddleware создает новый middleware для логирования
func NewLoggingMiddleware(logger logging.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Handle обрабатывает HTTP запросы с логированием
func (m *LoggingMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Генерируем request ID
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
		r = r.WithContext(ctx)

		// Создаем response writer для отслеживания статуса
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Логируем начало запроса
		m.logger.Info("HTTP request started",
			logging.String("method", r.Method),
			logging.String("path", r.URL.Path),
			logging.String("remote_addr", r.RemoteAddr),
			logging.String("user_agent", r.UserAgent()),
			logging.RequestID(requestID),
		)

		// Обрабатываем запрос
		next.ServeHTTP(wrapped, r)

		// Логируем завершение запроса
		duration := time.Since(start)
		m.logger.Info("HTTP request completed",
			logging.String("method", r.Method),
			logging.String("path", r.URL.Path),
			logging.Int("status_code", wrapped.statusCode),
			logging.Duration("duration", duration),
			logging.RequestID(requestID),
		)
	})
}

// MetricsMiddleware middleware для метрик
type MetricsMiddleware struct {
	metrics *monitoring.Metrics
}

// NewMetricsMiddleware создает новый middleware для метрик
func NewMetricsMiddleware(metrics *monitoring.Metrics) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics,
	}
}

// Handle обрабатывает HTTP запросы с метриками
func (m *MetricsMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем response writer для отслеживания статуса
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Обрабатываем запрос
		next.ServeHTTP(wrapped, r)

		// Записываем метрики
		duration := time.Since(start)
		statusCode := strconv.Itoa(wrapped.statusCode)

		m.metrics.RecordHTTPRequest(r.Method, r.URL.Path, statusCode, duration)
	})
}

// RecoveryMiddleware middleware для восстановления после паники
type RecoveryMiddleware struct {
	logger logging.Logger
}

// NewRecoveryMiddleware создает новый middleware для восстановления
func NewRecoveryMiddleware(logger logging.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

// Handle обрабатывает HTTP запросы с восстановлением после паники
func (m *RecoveryMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Получаем stack trace
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)

				// Логируем панику
				m.logger.Error("Panic recovered",
					logging.Any("error", err),
					logging.String("stack", string(stack[:length])),
					logging.String("method", r.Method),
					logging.String("path", r.URL.Path),
					logging.String("remote_addr", r.RemoteAddr),
				)

				// Отправляем 500 ошибку
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware middleware для CORS
type CORSMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
}

// NewCORSMiddleware создает новый middleware для CORS
func NewCORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders []string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: allowedOrigins,
		allowedMethods: allowedMethods,
		allowedHeaders: allowedHeaders,
	}
}

// Handle обрабатывает HTTP запросы с CORS
func (m *CORSMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Проверяем разрешенные origins
		if m.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.allowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.allowedHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Обрабатываем preflight запросы
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed проверяет, разрешен ли origin
func (m *CORSMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range m.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}

// RateLimitMiddleware middleware для ограничения скорости запросов
type RateLimitMiddleware struct {
	limiter RateLimiter
	logger  logging.Logger
}

// RateLimiter интерфейс для ограничения скорости
type RateLimiter interface {
	Allow(key string) bool
	Reset(key string)
}

// NewRateLimitMiddleware создает новый middleware для ограничения скорости
func NewRateLimitMiddleware(limiter RateLimiter, logger logging.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
		logger:  logger,
	}
}

// Handle обрабатывает HTTP запросы с ограничением скорости
func (m *RateLimitMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Используем IP адрес как ключ для ограничения
		key := getClientIP(r)

		if !m.limiter.Allow(key) {
			m.logger.Warn("Rate limit exceeded",
				logging.String("client_ip", key),
				logging.String("method", r.Method),
				logging.String("path", r.URL.Path),
			)

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware middleware для аутентификации
type AuthMiddleware struct {
	logger logging.Logger
}

// NewAuthMiddleware создает новый middleware для аутентификации
func NewAuthMiddleware(logger logging.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		logger: logger,
	}
}

// Handle обрабатывает HTTP запросы с аутентификацией
func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из заголовка
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Проверяем формат Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Здесь должна быть логика валидации токена
		// Для примера просто проверяем, что токен не пустой
		if token == "" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Извлекаем user ID из токена (заглушка)
		userID := int64(123) // В реальном проекте здесь должна быть логика извлечения user ID

		// Добавляем user ID в контекст
		ctx := context.WithValue(r.Context(), UserIDKey{}, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// responseWriter обертка для http.ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader записывает статус код
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper функции

// generateRequestID генерирует уникальный request ID
func generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getClientIP извлекает IP адрес клиента
func getClientIP(r *http.Request) string {
	// Проверяем заголовки прокси
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Используем RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// ChainMiddleware объединяет несколько middleware
func ChainMiddleware(middlewares ...Middleware) Middleware {
	return &chainMiddleware{middlewares: middlewares}
}

// chainMiddleware объединяет несколько middleware
type chainMiddleware struct {
	middlewares []Middleware
}

// Handle обрабатывает HTTP запросы через цепочку middleware
func (c *chainMiddleware) Handle(next http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		next = c.middlewares[i].Handle(next)
	}
	return next
}

// PrometheusMiddleware middleware для Prometheus метрик
type PrometheusMiddleware struct {
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
}

// NewPrometheusMiddleware создает новый middleware для Prometheus
func NewPrometheusMiddleware() *PrometheusMiddleware {
	return &PrometheusMiddleware{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		requestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),
	}
}

// Handle обрабатывает HTTP запросы с Prometheus метриками
func (m *PrometheusMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Увеличиваем счетчик активных запросов
		m.requestsInFlight.Inc()
		defer m.requestsInFlight.Dec()

		// Создаем response writer для отслеживания статуса
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Обрабатываем запрос
		next.ServeHTTP(wrapped, r)

		// Записываем метрики
		duration := time.Since(start)
		statusCode := strconv.Itoa(wrapped.statusCode)

		m.requestsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode).Inc()
		m.requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
	})
}

// RegisterPrometheusMetrics регистрирует Prometheus метрики
func (m *PrometheusMiddleware) RegisterPrometheusMetrics() {
	prometheus.MustRegister(m.requestsTotal)
	prometheus.MustRegister(m.requestDuration)
	prometheus.MustRegister(m.requestsInFlight)
}
