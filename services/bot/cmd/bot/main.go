// Package main provides the entry point for the Language Exchange Bot service.
//
//	@title			Language Exchange Bot Admin API
//	@version		3.0.0
//	@description	This is the administrative API for the Language Exchange Telegram Bot.
//	@description	It provides endpoints for monitoring, user management, feedback processing, and system statistics.
//
//	@contact.name	Language Exchange Bot Team
//	@contact.url	https://github.com/your-org/language-exchange-bot
//	@contact.email	support@language-exchange-bot.com
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-Admin-Key
//	@description				Admin API key for authentication
//
//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"language-exchange-bot/internal/adapters"
	"language-exchange-bot/internal/adapters/telegram"
	"language-exchange-bot/internal/config"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/server"
)

// Константы для таймаутов.
const (
	ForceShutdownTimeout = 10 * time.Second
)

func main() {
	cfg := config.Load()
	db := setupDatabase(cfg)

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	errorHandler := setupErrorHandler()

	ctx, cancel := setupGracefulShutdown()
	defer cancel()

	// Initialize Telegram handler for admin server (needed for webhook mode)
	var telegramHandler *telegram.TelegramHandler
	if cfg.EnableTelegram && cfg.TelegramToken != "" {
		// Create handler for admin server (will be reused by bot if needed)
		telegramHandler = telegram.NewTelegramHandlerWithAdmins(
			nil,        // bot API not needed for webhook endpoint
			nil,        // service will be set later
			[]int64{},  // empty admin chat IDs for webhook
			[]string{}, // empty admin usernames for webhook
			errorHandler,
		)
	}

	// Start admin API server
	adminServer := startAdminServer(cfg, db, errorHandler, telegramHandler)

	bots, wg := startBots(ctx, cfg, db, errorHandler, telegramHandler)
	waitForShutdown(bots, wg, adminServer, cancel)
}

// initializeTelegramBot инициализирует Telegram бота с сервисом.
func initializeTelegramBot(
	cfg *config.Config,
	db *database.DB,
	errorHandler *errors.ErrorHandler,
) (*telegram.TelegramBot, error) {
	// Создаем сервис с Redis кэшем
	service, err := core.NewBotServiceWithRedis(db, cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB, errorHandler)
	if err != nil {
		log.Printf("Failed to create Redis cache, falling back to in-memory cache: %v", err)
		// Fallback на in-memory кэш если Redis недоступен
		service = core.NewBotService(db, errorHandler)
	} else {
		log.Printf("Redis cache initialized: %s", service.Cache.String())
	}

	// Создаем бота с Chat ID для уведомлений и username для проверки прав
	telegramBot, err := telegram.NewTelegramBotWithService(cfg.TelegramToken, service, cfg.Debug, cfg.AdminUsernames)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	// Передаем errorHandler в TelegramBot
	telegramBot.SetErrorHandler(errorHandler)

	// Устанавливаем Chat ID для уведомлений
	telegramBot.SetAdminChatIDs(cfg.AdminChatIDs)

	// Связываем бота с сервисом для отправки уведомлений о новых отзывах
	botService := telegramBot.GetService()
	if botService != nil {
		botService.SetFeedbackNotificationFunc(telegramBot.SendFeedbackNotification)
		log.Printf("Связал функцию уведомлений с сервисом отзывов")
	}

	return telegramBot, nil
}

// setupDatabase подключается к базе данных.
func setupDatabase(cfg *config.Config) *database.DB {
	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database successfully")

	return db
}

// setupErrorHandler создает систему обработки ошибок.
func setupErrorHandler() *errors.ErrorHandler {
	adminNotifier := errors.NewAdminNotifier([]int64{}, nil) // TODO: Добавить реальные Chat ID администраторов

	return errors.NewErrorHandler(adminNotifier)
}

// setupGracefulShutdown настраивает graceful shutdown.
func setupGracefulShutdown() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal...")
		cancel()
	}()

	return ctx, cancel
}

// startBots запускает все настроенные боты.
func startBots(
	ctx context.Context,
	cfg *config.Config,
	db *database.DB,
	errorHandler *errors.ErrorHandler,
	telegramHandler *telegram.TelegramHandler,
) ([]adapters.BotAdapter, *sync.WaitGroup) {
	var (
		wg   sync.WaitGroup
		bots []adapters.BotAdapter
	)

	// Telegram Bot
	if cfg.EnableTelegram && cfg.TelegramToken != "" {
		telegramBot, err := initializeTelegramBot(cfg, db, errorHandler)
		if err != nil {
			log.Fatalf("Failed to initialize Telegram bot: %v", err)
		}

		// Если webhook режим, устанавливаем handler в admin server и настраиваем webhook
		if cfg.TelegramMode == "webhook" && telegramHandler != nil {
			// Обновляем handler в admin server с реальным сервисом
			telegramHandler.SetService(telegramBot.GetService())
			telegramHandler.SetBotAPI(telegramBot.GetBotAPI())

			// Настраиваем webhook в Telegram
			if cfg.WebhookURL != "" {
				webhookURL := fmt.Sprintf("%s/webhook/telegram/%s", cfg.WebhookURL, cfg.TelegramToken)
				if err := telegramBot.SetupWebhook(webhookURL); err != nil {
					log.Fatalf("Failed to setup webhook: %v", err)
				}
				log.Printf("Webhook mode: Telegram handler configured for admin server at %s", webhookURL)
			} else {
				log.Fatalf("Webhook mode requires WEBHOOK_URL to be set")
			}
		} else {
			// Polling режим - запускаем бота обычным образом
			bots = append(bots, telegramBot)

			wg.Add(1)

			go func() {
				defer wg.Done()

				if err := telegramBot.Start(ctx); err != nil {
					log.Printf("Telegram bot error: %v", err)
				}
			}()

			log.Printf("Telegram bot started in polling mode with %d admin users", telegramBot.GetAdminCount())
		}
	}

	// Будущие боты (Discord, etc)
	if cfg.EnableDiscord {
		log.Println("Discord bot is not implemented yet")
	}

	if len(bots) == 0 {
		log.Fatal("No bots are enabled. Please check your configuration.")
	}

	return bots, &wg
}

// startAdminServer запускает административный API сервер
func startAdminServer(cfg *config.Config, db *database.DB, errorHandler *errors.ErrorHandler, telegramHandler *telegram.TelegramHandler) *server.AdminServer {
	// Создаем сервис с Redis кэшем
	service, err := core.NewBotServiceWithRedis(db, cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB, errorHandler)
	if err != nil {
		log.Printf("Failed to create Redis cache, falling back to in-memory cache: %v", err)
		// Fallback на in-memory кэш если Redis недоступен
		service = core.NewBotService(db, errorHandler)
	} else {
		log.Printf("Redis cache initialized for admin API: %s", service.Cache.String())
	}

	// Создаем handler для Telegram (нужен для статистики rate limiting)
	telegramHandler = telegram.NewTelegramHandlerWithAdmins(
		nil, // bot API не нужен для admin API
		service,
		[]int64{},  // пустой список admin chat IDs
		[]string{}, // пустой список admin usernames
		errorHandler,
	)

	// Создаем и запускаем admin server
	adminPort := "8080" // можно вынести в конфиг
	webhookMode := cfg.TelegramMode == "webhook"
	adminServer := server.NewWithWebhook(adminPort, service, telegramHandler, webhookMode)

	go func() {
		if err := adminServer.Start(); err != nil {
			log.Printf("Admin server error: %v", err)
		}
	}()

	return adminServer
}

// waitForShutdown ждет завершения работы ботов и admin сервера.
func waitForShutdown(bots []adapters.BotAdapter, wg *sync.WaitGroup, adminServer *server.AdminServer, cancel context.CancelFunc) {
	// Ждем завершения всех горутин
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Ждем либо завершения работы, либо сигнала остановки
	select {
	case <-done:
		log.Println("All bots stopped gracefully")
	case <-time.After(ForceShutdownTimeout):
		log.Println("Force shutdown")
		cancel()
	}

	// Останавливаем admin сервер
	if adminServer != nil {
		ctx, cancelStop := context.WithTimeout(context.Background(), 5*time.Second)
		if err := adminServer.Stop(ctx); err != nil {
			log.Printf("Error stopping admin server: %v", err)
		}
		cancelStop()
		log.Println("Admin server stopped")
	}

	// Останавливаем все боты
	ctx, cancelStop := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelStop()

	for _, bot := range bots {
		if err := bot.Stop(ctx); err != nil {
			log.Printf("Error stopping bot: %v", err)
		}
	}

	// Останавливаем кэш-сервис только после остановки ботов
	for _, bot := range bots {
		if telegramBot, ok := bot.(*telegram.TelegramBot); ok {
			service := telegramBot.GetService()
			if service != nil {
				service.StopCache()
				log.Println("Cache service stopped")
			}
		}
	}
}
