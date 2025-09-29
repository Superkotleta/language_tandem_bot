package handlers

import (
	"language-exchange-bot/internal/core"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UtilityHandler интерфейс для вспомогательных функций
type UtilityHandler interface {
	SendMessage(chatID int64, text string) error
}

// UtilityHandlerImpl реализация вспомогательного обработчика
type UtilityHandlerImpl struct {
	service *core.BotService
	bot     *tgbotapi.BotAPI
}

// NewUtilityHandler создает новый вспомогательный обработчик
func NewUtilityHandler(service *core.BotService, bot *tgbotapi.BotAPI) UtilityHandler {
	return &UtilityHandlerImpl{
		service: service,
		bot:     bot,
	}
}

// SendMessage отправляет сообщение пользователю
func (h *UtilityHandlerImpl) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	return err
}
