package base

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
)

// BaseHandler содержит общие зависимости для всех handlers.
// Используется через композицию для уменьшения дублирования кода.
//
// Пример использования:
//
//	base := NewBaseHandler(bot, service, keyboardBuilder, errorHandler, messageFactory)
//	handler := NewFeedbackHandler(base, adminChatIDs, adminUsernames)
//	handler.Base.MessageFactory.SendText(chatID, "Hello")
type BaseHandler struct {
	Bot             *tgbotapi.BotAPI
	Service         *core.BotService
	KeyboardBuilder *KeyboardBuilder
	ErrorHandler    *errors.ErrorHandler
	MessageFactory  *MessageFactory
}

// NewBaseHandler создает новый BaseHandler с общими зависимостями.
func NewBaseHandler(
	bot *tgbotapi.BotAPI,
	service *core.BotService,
	keyboardBuilder *KeyboardBuilder,
	errorHandler *errors.ErrorHandler,
	messageFactory *MessageFactory,
) *BaseHandler {
	return &BaseHandler{
		Bot:             bot,
		Service:         service,
		KeyboardBuilder: keyboardBuilder,
		ErrorHandler:    errorHandler,
		MessageFactory:  messageFactory,
	}
}

// GetBot возвращает экземпляр Telegram Bot API.
func (b *BaseHandler) GetBot() *tgbotapi.BotAPI {
	return b.Bot
}

// GetService возвращает основной сервис бота.
func (b *BaseHandler) GetService() *core.BotService {
	return b.Service
}

// GetKeyboardBuilder возвращает построитель клавиатур.
func (b *BaseHandler) GetKeyboardBuilder() *KeyboardBuilder {
	return b.KeyboardBuilder
}

// GetErrorHandler возвращает обработчик ошибок.
func (b *BaseHandler) GetErrorHandler() *errors.ErrorHandler {
	return b.ErrorHandler
}

// GetMessageFactory возвращает фабрику сообщений.
func (b *BaseHandler) GetMessageFactory() *MessageFactory {
	return b.MessageFactory
}
