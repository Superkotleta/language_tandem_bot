package telegram

import (
	"context"
	"fmt"
	"log"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	api     *tgbotapi.BotAPI
	service *core.BotService
	debug   bool
}

func NewTelegramBot(token string, db *database.DB, debug bool) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}
	bot.Debug = debug
	service := core.NewBotService(db)
	return &TelegramBot{
		api:     bot,
		service: service,
		debug:   debug,
	}, nil
}

func (tb *TelegramBot) Start(ctx context.Context) error {
	log.Printf("Authorized on account %s", tb.api.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.api.GetUpdatesChan(u)
	handler := NewTelegramHandler(tb.api, tb.service)

	for {
		select {
		case update := <-updates:
			go func(upd tgbotapi.Update) {
				if err := handler.HandleUpdate(upd); err != nil {
					log.Printf("Error handling update: %v", err)
				}
			}(update)
		case <-ctx.Done():
			log.Println("Stopping Telegram bot...")
			tb.api.StopReceivingUpdates()
			return nil
		}
	}
}

func (tb *TelegramBot) Stop(ctx context.Context) error {
	tb.api.StopReceivingUpdates()
	return nil
}

func (tb *TelegramBot) GetPlatformName() string {
	return "telegram"
}
