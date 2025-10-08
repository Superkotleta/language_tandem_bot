package handlers

// UtilityHandler интерфейс для вспомогательных функций.
type UtilityHandler interface {
	SendMessage(chatID int64, text string) error
}

// UtilityHandlerImpl реализация вспомогательного обработчика.
type UtilityHandlerImpl struct {
	base *BaseHandler
}

// NewUtilityHandler создает новый вспомогательный обработчик.
func NewUtilityHandler(base *BaseHandler) *UtilityHandlerImpl {
	return &UtilityHandlerImpl{
		base: base,
	}
}

// SendMessage отправляет сообщение пользователю.
func (h *UtilityHandlerImpl) SendMessage(chatID int64, text string) error {
	// Используем MessageFactory для отправки сообщения
	return h.base.messageFactory.SendText(chatID, text)
}
