package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"language-exchange-bot/internal/domain"
	"language-exchange-bot/internal/pkg/i18n"
	"language-exchange-bot/internal/service"
	"language-exchange-bot/internal/ui/keyboards"
	"language-exchange-bot/internal/ui/messages"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api             *tgbotapi.BotAPI
	userService     *service.UserService
	localizer       *i18n.Localizer
	messageFactory  *messages.MessageFactory
	keyboardBuilder *keyboards.KeyboardBuilder
}

func NewBot(
	token string,
	userService *service.UserService,
	localizer *i18n.Localizer,
) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot api: %w", err)
	}

	return &Bot{
		api:             api,
		userService:     userService,
		localizer:       localizer,
		messageFactory:  messages.NewMessageFactory(),
		keyboardBuilder: keyboards.NewKeyboardBuilder(localizer),
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
			b.api.StopReceivingUpdates()
			return
		case update := <-updates:
			if update.Message != nil {
				b.handleMessage(ctx, update.Message)
			}
		}
	}
}

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	if message.From == nil {
		log.Printf("Received message from nil user (likely channel post or service message), ignoring. Message ID: %d", message.MessageID)
		return
	}

	socialID := strconv.FormatInt(message.From.ID, 10)

	// 1. Get or Create User
	user, err := b.userService.GetUserBySocialID(ctx, socialID, domain.PlatformTelegram)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return
	}

	if user == nil {
		// User doesn't exist, register them
		user, err = b.userService.RegisterUser(
			ctx,
			socialID,
			domain.PlatformTelegram,
			message.From.FirstName,
			message.From.UserName,
			message.From.LanguageCode,
		)
		if err != nil {
			log.Printf("Error registering user: %v", err)
			return
		}
	}

	// 2. Handle Commands
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			b.handleStart(ctx, message, user)
		default:
			// Unknown command
		}
		return
	}

	// 3. Handle Text Messages (Menu)
	b.handleMenu(ctx, message, user)
}

func (b *Bot) handleStart(ctx context.Context, message *tgbotapi.Message, user *domain.User) {
	lang := user.InterfaceLang
	text := b.localizer.Get(lang, "welcome_message")

	// Replace template variables
	text = b.localizer.Replace(text, map[string]string{
		"{name}": user.FirstName,
	})

	hasProfile := user.Status == domain.StatusActive
	keyboard := b.keyboardBuilder.MainMenu(lang, hasProfile)

	msg := b.messageFactory.NewKeyboardMessage(message.Chat.ID, text, keyboard)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to send start message: %v", err)
	}
}

func (b *Bot) handleMenu(ctx context.Context, message *tgbotapi.Message, user *domain.User) {
	lang := user.InterfaceLang
	text := message.Text

	var err error
	switch text {
	case b.localizer.Get(lang, "btn_fill_profile"):
		_, err = b.api.Send(b.messageFactory.NewText(message.Chat.ID, "Starting profile setup... (Wizard not implemented yet)"))
	case b.localizer.Get(lang, "btn_profile"):
		// Show profile with inline actions
		profileText := fmt.Sprintf("Name: %s\nNative: %s\nTarget: %s", user.FirstName, user.NativeLang, user.TargetLang)
		keyboard := b.keyboardBuilder.ProfileActions(lang)
		_, err = b.api.Send(b.messageFactory.NewInlineKeyboardMessage(message.Chat.ID, profileText, keyboard))
	default:
		_, err = b.api.Send(b.messageFactory.NewText(message.Chat.ID, b.localizer.Get(lang, "unknown_command")))
	}

	if err != nil {
		log.Printf("Failed to send menu message: %v", err)
	}
}
