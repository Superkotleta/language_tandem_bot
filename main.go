package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"language-exchange-bot/internal/config"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	bot.Debug = cfg.Debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Создаем обработчик
	handler := handlers.NewTelegramHandler(bot, db)

	// Настраиваем webhook или polling
	if cfg.WebhookURL != "" {
		// Webhook mode для продакшена
		wh, _ := tgbotapi.NewWebhook(cfg.WebhookURL)
		bot.Request(wh)

		updates := bot.ListenForWebhook("/")
		go http.ListenAndServe(":"+cfg.Port, nil)

		log.Printf("Webhook started on port %s", cfg.Port)
		handler.HandleUpdates(updates)
	} else {
		// Polling mode для разработки
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates := bot.GetUpdatesChan(u)
		log.Println("Bot started in polling mode")

		handler.HandleUpdates(updates)
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Bot stopped")
}
