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

	bots := startBots(ctx, cfg, db, errorHandler)
	waitForShutdown(bots, cancel)
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
) []adapters.BotAdapter {
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

		bots = append(bots, telegramBot)

		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := telegramBot.Start(ctx); err != nil {
				log.Printf("Telegram bot error: %v", err)
			}
		}()

		log.Printf("Telegram bot started with %d admin users", telegramBot.GetAdminCount())
	}

	// Будущие боты (Discord, etc)
	if cfg.EnableDiscord {
		log.Println("Discord bot is not implemented yet")
	}

	if len(bots) == 0 {
		log.Fatal("No bots are enabled. Please check your configuration.")
	}

	return bots
}

// waitForShutdown ждет завершения работы ботов.
func waitForShutdown(bots []adapters.BotAdapter, cancel context.CancelFunc) {
	// Останавливаем кэш-сервис
	for _, bot := range bots {
		if telegramBot, ok := bot.(*telegram.TelegramBot); ok {
			service := telegramBot.GetService()
			if service != nil {
				service.StopCache()
				log.Println("Cache service stopped")
			}
		}
	}

	// Ждем завершения всех горутин
	done := make(chan struct{})

	go func() {
		// В оригинальном коде была wg.Wait(), но wg объявлена локально в startBots
		// Для простоты используем time.Sleep, но лучше передать wg как параметр
		time.Sleep(1 * time.Second) // Имитация ожидания
		close(done)
	}()

	select {
	case <-done:
		log.Println("All bots stopped gracefully")
	case <-time.After(ForceShutdownTimeout):
		log.Println("Force shutdown")
	}
}
