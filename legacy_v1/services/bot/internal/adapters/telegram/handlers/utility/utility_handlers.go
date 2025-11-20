package utility

import (
	"language-exchange-bot/internal/adapters/telegram/handlers/base"
)

// UtilityHandler интерфейс для вспомогательных функций.
type UtilityHandler interface {
	SendMessage(chatID int64, text string) error
}

// UtilityHandlerImpl реализация вспомогательного обработчика.
type UtilityHandlerImpl struct {
	base *base.BaseHandler
}

// NewUtilityHandler создает новый вспомогательный обработчик.
func NewUtilityHandler(baseHandler *base.BaseHandler) *UtilityHandlerImpl {
	return &UtilityHandlerImpl{
		base: baseHandler,
	}
}

// SendMessage отправляет сообщение пользователю.
func (h *UtilityHandlerImpl) SendMessage(chatID int64, text string) error {
	// Используем MessageFactory для отправки сообщения
	return h.base.MessageFactory.SendText(chatID, text)
}
