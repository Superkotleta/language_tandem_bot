package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"language-exchange-bot/internal/adapters/telegram"
	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/config"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/health"
	"language-exchange-bot/internal/logging"
	"language-exchange-bot/internal/middleware"
	"language-exchange-bot/internal/monitoring"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	logger.Info("Starting Language Exchange Bot (Optimized Version with Telegram)")

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
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

	// Парсим администраторов
	adminChatIDs := cfg.AdminChatIDs
	adminUsernames := cfg.AdminUsernames

	// Создаем оптимизированный Telegram бот
	telegramBot, err := telegram.NewOptimizedTelegramBot(
		cfg.TelegramToken,
		db,
		cacheInstance,
		metrics,
		logger,
		cfg.Debug,
		adminChatIDs,
		adminUsernames,
	)
	if err != nil {
		logger.Fatal("Failed to create Telegram bot", logging.ErrorField(err))
	}

	// Создаем health manager
	healthManager := health.NewHealthManager(logger, "2.0.0")

	// Добавляем проверки здоровья
	healthManager.AddChecker(health.NewDatabaseHealthChecker("database", func(ctx context.Context) error {
		return service.HealthCheck()
	}))

	healthManager.AddChecker(health.NewCacheHealthChecker("cache", func(ctx context.Context) error {
		return cacheInstance.Set(ctx, "health_check", "ok", time.Second)
	}))

	healthManager.AddChecker(health.NewMemoryHealthChecker("memory", 100*1024*1024)) // 100MB

	// Добавляем проверку Telegram бота
	healthManager.AddChecker(health.NewExternalServiceHealthChecker("telegram", "https://api.telegram.org", 5*time.Second))

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

	// API endpoints
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

	// Webhook endpoint для Telegram (production режим)
	mux.HandleFunc("/webhook/telegram", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Читаем обновление от Telegram
		var update tgbotapi.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			logger.Error("Failed to decode webhook update", logging.ErrorField(err))
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Обрабатываем обновление асинхронно
		go func() {
			if err := telegramBot.GetHandler().HandleUpdate(update); err != nil {
				logger.Error("Error handling webhook update", logging.ErrorField(err))
			}
		}()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
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

	// Запускаем HTTP сервер в горутине
	go func() {
		logger.Info("Starting HTTP server", logging.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", logging.ErrorField(err))
		}
	}()

	// Запускаем Telegram бот в горутине
	go func() {
		logger.Info("Starting Telegram bot")
		if err := telegramBot.Start(ctx); err != nil {
			logger.Error("Telegram bot failed", logging.ErrorField(err))
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

	// Останавливаем Telegram бот
	if err := telegramBot.Stop(shutdownCtx); err != nil {
		logger.Error("Telegram bot shutdown failed", logging.ErrorField(err))
	}

	// Останавливаем HTTP сервер
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", logging.ErrorField(err))
	}

	logger.Info("Server stopped")
}
