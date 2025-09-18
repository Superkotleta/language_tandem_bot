package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/config"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/health"
	"language-exchange-bot/internal/logging"
	"language-exchange-bot/internal/middleware"
	"language-exchange-bot/internal/monitoring"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем логгер
	logger, err := logging.NewProductionLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.(*logging.ZapLogger).Sync()

	logging.SetGlobalLogger(logger)
	logger.Info("Starting Language Exchange Bot (Optimized Version)")

	// Создаем контекст с отменой
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализируем оптимизированную БД
	db, err := database.NewOptimizedDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", logging.ErrorField(err))
	}
	defer db.Close()

	// Инициализируем кэш
	var cacheInstance cache.Cache
	if cfg.RedisURL != "" {
		// Используем Redis кэш
		cacheInstance, err = cache.NewRedisCache(cfg.RedisURL, "", 0)
		if err != nil {
			logger.Warn("Failed to connect to Redis, using memory cache", logging.ErrorField(err))
			cacheInstance = cache.NewMemoryCache()
		}
	} else {
		// Используем кэш в памяти
		cacheInstance = cache.NewMemoryCache()
	}
	defer cacheInstance.Close()

	// Создаем оптимизированный сервис
	service := core.NewOptimizedBotService(db, cacheInstance)
	defer service.Close()

	// Инициализируем метрики
	metrics := monitoring.NewMetrics()

	// Создаем health manager
	healthManager := health.NewHealthManager(logger, "1.0.0")

	// Добавляем проверки здоровья
	healthManager.AddChecker(health.NewDatabaseHealthChecker("database", func(ctx context.Context) error {
		return service.HealthCheck()
	}))

	healthManager.AddChecker(health.NewCacheHealthChecker("cache", func(ctx context.Context) error {
		return cacheInstance.Set(ctx, "health_check", "ok", time.Second)
	}))

	healthManager.AddChecker(health.NewMemoryHealthChecker("memory", 100*1024*1024)) // 100MB

	// Создаем HTTP сервер
	mux := http.NewServeMux()

	// Добавляем middleware
	chain := middleware.ChainMiddleware(
		middleware.NewRecoveryMiddleware(logger),
		middleware.NewLoggingMiddleware(logger),
		middleware.NewMetricsMiddleware(metrics),
		middleware.NewCORSMiddleware(
			[]string{"*"},
			[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			[]string{"Content-Type", "Authorization"},
		),
	)

	// Health check endpoints
	mux.HandleFunc("/health", healthManager.HTTPHandler())
	mux.HandleFunc("/health/ready", healthManager.ReadinessHandler())
	mux.HandleFunc("/health/live", healthManager.LivenessHandler())

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// API endpoints (заглушки для примера)
	mux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Users endpoint"}`))
	})

	mux.HandleFunc("/api/v1/feedback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Feedback endpoint"}`))
	})

	// Применяем middleware
	handler := chain.Handle(mux)

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		logger.Info("Starting HTTP server", logging.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", logging.ErrorField(err))
		}
	}()

	// Запускаем сбор метрик
	metricsManager := monitoring.NewMetricsManager(30 * time.Second)
	metricsManager.AddCollector(monitoring.NewSystemMetricsCollector(metrics))
	metricsManager.AddCollector(monitoring.NewDBMetricsCollector(metrics, db))
	go metricsManager.Start()
	defer metricsManager.Stop()

	// Ждем сигнал завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", logging.ErrorField(err))
	}

	logger.Info("Server stopped")
}

