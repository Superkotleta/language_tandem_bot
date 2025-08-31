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
		telegramBot, err := telegram.NewTelegramBot(cfg.TelegramToken, db, cfg.Debug)
		if err != nil {
			log.Fatalf("Failed to create Telegram bot: %v", err)
		}

		bots = append(bots, telegramBot)

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := telegramBot.Start(ctx); err != nil {
				log.Printf("Telegram bot error: %v", err)
			}
		}()

		log.Println("Telegram bot started")
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
