package handlers

import (
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UtilityHandler интерфейс для вспомогательных функций.
type UtilityHandler interface {
	SendMessage(chatID int64, text string) error
}

// UtilityHandlerImpl реализация вспомогательного обработчика.
type UtilityHandlerImpl struct {
	service      *core.BotService
	bot          *tgbotapi.BotAPI
	errorHandler *errors.ErrorHandler
}

// NewUtilityHandler создает новый вспомогательный обработчик.
func NewUtilityHandler(
	service *core.BotService,
	bot *tgbotapi.BotAPI,
	errorHandler *errors.ErrorHandler,
) *UtilityHandlerImpl {
	return &UtilityHandlerImpl{
		service:      service,
		bot:          bot,
		errorHandler: errorHandler,
	}
}

// SendMessage отправляет сообщение пользователю.
func (h *UtilityHandlerImpl) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)

	_, err := h.bot.Send(msg)
	if err != nil {
		// Используем новую систему обработки ошибок
		return h.errorHandler.HandleTelegramError(
			err,
			chatID,
			0, // UserID неизвестен в этом контексте
			"SendMessage",
		)
	}

	return nil
}
