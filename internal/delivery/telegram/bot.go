package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"language-exchange-bot/internal/domain"
	"language-exchange-bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api         *tgbotapi.BotAPI
	userService *service.UserService
}

func NewBot(token string, userService *service.UserService) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot api: %w", err)
	}

	return &Bot{
		api:         api,
		userService: userService,
	}, nil
}

func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	log.Println("Telegram bot started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Telegram bot...")
			return
		case update := <-updates:
			if update.Message != nil {
				b.handleMessage(ctx, update.Message)
			}
		}
	}
}

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			b.handleStart(ctx, message)
		default:
			msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command")
			b.api.Send(msg)
		}
	}
}

func (b *Bot) handleStart(ctx context.Context, message *tgbotapi.Message) {
	socialID := strconv.FormatInt(message.From.ID, 10)
	firstName := message.From.FirstName
	username := message.From.UserName
	languageCode := message.From.LanguageCode

	user, err := b.userService.RegisterUser(
		ctx,
		socialID,
		domain.PlatformTelegram,
		firstName,
		username,
		languageCode,
	)

	if err != nil {
		log.Printf("Failed to register user: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error registering user. Please try again later.")
		b.api.Send(msg)
		return
	}

	welcomeText := fmt.Sprintf("Hello, %s! You have been registered.\nPlatform: %s\nLanguage: %s", 
		user.FirstName, user.Platform, user.Language)
	
	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	b.api.Send(msg)
}


