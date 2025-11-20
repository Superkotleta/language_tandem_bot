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
	"runtime/debug"
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
	ForceShutdownTimeout = 5 * time.Minute // Увеличен для предотвращения перезапуска Docker в dev режиме
)

func main() {
	// Глобальный recover для отлова всех паник
	defer func() {
		if r := recover(); r != nil {
			log.Printf("FATAL PANIC in main: %v", r)
			log.Printf("Stack trace: %s", debug.Stack())
			os.Exit(1)
		}
	}()

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

	// Создаем общий сервис для всех компонентов (бот и admin API)
	service, err := initializeService(cfg, db, errorHandler)
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// Start admin API server с общим сервисом
	adminServer := startAdminServerWithService(cfg, service, errorHandler)

	// Start bots с общим сервисом
	bots, wg, _ := startBotsWithService(ctx, cfg, service, errorHandler)

	waitForShutdown(bots, wg, adminServer, ctx, cancel)
}

// initializeService создает общий сервис для всех компонентов.
func initializeService(
	cfg *config.Config,
	db *database.DB,
	errorHandler *errors.ErrorHandler,
) (*core.BotService, error) {
	// Создаем сервис с Redis кэшем
	service, err := core.NewBotServiceWithRedis(db, cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB, errorHandler)
	if err != nil {
		log.Printf("Failed to create Redis cache, falling back to in-memory cache: %v", err)
		// Fallback на in-memory кэш если Redis недоступен
		service = core.NewBotService(db, errorHandler)
	} else {
		log.Printf("Redis cache initialized: %s", service.Cache.String())
	}

	return service, nil
}

// initializeTelegramBotWithService инициализирует Telegram бота с готовым сервисом.
func initializeTelegramBotWithService(
	service *core.BotService,
	cfg *config.Config,
	errorHandler *errors.ErrorHandler,
) (*telegram.TelegramBot, error) {
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
		sig := <-sigChan
		log.Printf("Received shutdown signal: %v", sig)
		cancel()
	}()

	return ctx, cancel
}

// startAdminServerWithService запускает административный API сервер с общим сервисом.
func startAdminServerWithService(cfg *config.Config, service *core.BotService, errorHandler *errors.ErrorHandler) *server.AdminServer {
	log.Printf("Admin API using shared service: %s", service.Cache.String())

	// Создаем handler для Telegram для статистики rate limiting
	telegramHandler := telegram.NewTelegramHandlerWithAdmins(
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
func waitForShutdown(bots []adapters.BotAdapter, wg *sync.WaitGroup, adminServer *server.AdminServer, ctx context.Context, cancel context.CancelFunc) {
	log.Printf("waitForShutdown called with %d bots", len(bots))

	// Если нет запущенных ботов, но есть admin API, ждем сигнала shutdown
	if len(bots) == 0 {
		log.Println("No bots running, waiting for shutdown signal...")
		<-ctx.Done()

		return
	}

	// Ждем завершения всех горутин
	done := make(chan struct{})

	go func() {
		log.Println("Waiting for all goroutines to finish...")
		wg.Wait()
		log.Println("All goroutines finished")
		close(done)
	}()

	// Ждем либо завершения работы, либо сигнала остановки
	select {
	case <-done:
		log.Println("All bots stopped gracefully")
	case <-time.After(ForceShutdownTimeout):
		log.Printf("Force shutdown after %v timeout", ForceShutdownTimeout)
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
	// Используем общий сервис из любого бота
	for _, bot := range bots {
		if telegramBot, ok := bot.(*telegram.TelegramBot); ok {
			service := telegramBot.GetService()
			if service != nil {
				service.StopCache()
				log.Println("Cache service stopped")

				break // останавливаем только один раз для общего сервиса
			}
		}
	}
}

// startBotsWithService запускает все боты с общим сервисом.
func startBotsWithService(
	ctx context.Context,
	cfg *config.Config,
	service *core.BotService,
	errorHandler *errors.ErrorHandler,
) ([]adapters.BotAdapter, *sync.WaitGroup, *telegram.TelegramHandler) {
	var (
		wg   sync.WaitGroup
		bots []adapters.BotAdapter
	)

	// Telegram Bot
	if cfg.EnableTelegram && cfg.TelegramToken != "" {
		telegramBot, err := initializeTelegramBotWithService(service, cfg, errorHandler)
		if err != nil {
			log.Printf("Failed to initialize Telegram bot, continuing with Admin API only: %v", err)
		} else {
			// Создаем handler для Telegram для admin server
			telegramHandler := telegram.NewTelegramHandlerWithAdmins(
				telegramBot.GetBotAPI(),
				service,
				cfg.AdminChatIDs,
				cfg.AdminUsernames,
				errorHandler,
			)

			// Запускаем бота
			wg.Add(1)

			go func() {
				defer wg.Done()

				log.Printf("Starting Telegram bot...")

				if err := telegramBot.Start(ctx); err != nil {
					log.Printf("Telegram bot error: %v", err)
				}

				log.Printf("Telegram bot stopped")
			}()

			bots = append(bots, telegramBot)

			return bots, &wg, telegramHandler
		}
	}

	return bots, &wg, nil
}
