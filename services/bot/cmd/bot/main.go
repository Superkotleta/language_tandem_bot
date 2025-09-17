package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"language-exchange-bot/internal/adapters"
	"language-exchange-bot/internal/adapters/telegram"
	"language-exchange-bot/internal/config"
	"language-exchange-bot/internal/database"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database successfully")

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Слушаем сигналы системы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем боты
	var wg sync.WaitGroup
	var bots []adapters.BotAdapter

	// Telegram Bot
	if cfg.EnableTelegram && cfg.TelegramToken != "" {
		// Создаем бота с Chat ID для уведомлений и username для проверки прав
		telegramBot, err := telegram.NewTelegramBotWithUsernames(cfg.TelegramToken, db, cfg.Debug, cfg.AdminUsernames)
		if err != nil {
			log.Fatalf("Failed to create Telegram bot: %v", err)
		}

		// Устанавливаем Chat ID для уведомлений
		telegramBot.SetAdminChatIDs(cfg.AdminChatIDs)

		// Связываем бота с сервисом для отправки уведомлений о новых отзывах
		service := telegramBot.GetService()
		if service != nil {
			service.SetFeedbackNotificationFunc(telegramBot.SendFeedbackNotification)
			log.Printf("Связал функцию уведомлений с сервисом отзывов")
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

	// Ждем сигнал для остановки
	<-sigChan
	log.Println("Received shutdown signal...")

	// Останавливаем все боты
	cancel()

	// Ждем завершения всех горутин
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All bots stopped gracefully")
	case <-sigChan:
		log.Println("Force shutdown")
	}
}
