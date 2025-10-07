package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"language-exchange-bot/internal/adapters/telegram"
	"language-exchange-bot/internal/api/handlers"
	"language-exchange-bot/internal/api/middleware"
	"language-exchange-bot/internal/api/server"
	"language-exchange-bot/internal/config"
	"language-exchange-bot/internal/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Load configuration
	cfg := config.Load()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	logger.Info("starting Bot API service", zap.String("port", cfg.HTTPPort))

	// Connect to database
	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("connected to database successfully")

	// Create Telegram bot (for webhook processing)
	var telegramBot *telegram.TelegramBot
	if cfg.EnableTelegram && cfg.TelegramToken != "" {
		telegramBot, err = telegram.NewTelegramBotWithUsernames(cfg.TelegramToken, db, cfg.Debug, cfg.AdminUsernames)
		if err != nil {
			logger.Fatal("failed to create Telegram bot", zap.Error(err))
		}

		// Set admin chat IDs
		telegramBot.SetAdminChatIDs(cfg.AdminChatIDs)

		// Link feedback notification function
		service := telegramBot.GetService()
		if service != nil {
			service.SetFeedbackNotificationFunc(telegramBot.SendFeedbackNotification)
			logger.Info("linked feedback notification function")
		}

		logger.Info("Telegram bot initialized", zap.Int("admin_count", telegramBot.GetAdminCount()))
	}

	// Create handlers
	healthHandler := handlers.NewHealthHandler()
	var telegramHandler *handlers.TelegramHandler
	if telegramBot != nil {
		telegramHandler = handlers.NewTelegramHandler(telegramBot, logger)
	}

	// Setup router
	router := gin.New()

	// Middleware
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware(logger))

	// Health check endpoints
	router.GET("/healthz", healthHandler.Health)
	router.GET("/readyz", healthHandler.Ready)

	// API routes
	api := router.Group("/api/v1")
	{
		// Bot routes
		bot := api.Group("/bot")
		{
			if telegramHandler != nil {
				bot.POST("/telegram/webhook", telegramHandler.Webhook)
				bot.POST("/telegram/webhook/set", telegramHandler.SetWebhook)
				bot.GET("/telegram/webhook/info", telegramHandler.GetWebhookInfo)
				bot.POST("/telegram/send", telegramHandler.SendMessage)
			}
		}

		// Telegram routes (for backward compatibility)
		telegram := api.Group("/telegram")
		{
			if telegramHandler != nil {
				telegram.POST("/webhook", telegramHandler.Webhook)
				telegram.POST("/webhook/set", telegramHandler.SetWebhook)
				telegram.GET("/webhook/info", telegramHandler.GetWebhookInfo)
				telegram.POST("/send", telegramHandler.SendMessage)
			}
		}
	}

	// Start server
	srv := server.New(cfg.HTTPPort, router)
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()
	logger.Info("Bot API service started", zap.String("port", cfg.HTTPPort))

	// Start bot in polling mode if DEBUG is true
	if cfg.Debug && telegramBot != nil {
		logger.Info("Starting bot in polling mode (DEBUG=true)")
		go func() {
			ctx := context.Background()
			if err := telegramBot.Start(ctx); err != nil {
				logger.Error("Error starting bot in polling mode", zap.Error(err))
			}
		}()
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop bot gracefully if it's running
	if telegramBot != nil {
		if err := telegramBot.Stop(ctx); err != nil {
			logger.Error("Error stopping bot", zap.Error(err))
		}
	}

	_ = srv.Shutdown(ctx)
	logger.Info("Bot API service stopped")
}
