package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"language-exchange-bot/internal/core"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UtilityHandler интерфейс для вспомогательных функций.
type UtilityHandler interface {
	SendMessage(chatID int64, text string) error
	CalculatePercentage(part, total int) int
	GetFeedbackNavigationState(userID int64, feedbackType string, currentIndex int) string
	ParseFeedbackNavigationState(stateStr string) (userID int64, feedbackType string, currentIndex int)
}

// UtilityHandlerImpl реализация вспомогательного обработчика.
type UtilityHandlerImpl struct {
	service *core.BotService
	bot     *tgbotapi.BotAPI
}

// NewUtilityHandler создает новый вспомогательный обработчик.
func NewUtilityHandler(service *core.BotService, bot *tgbotapi.BotAPI) UtilityHandler {
	return &UtilityHandlerImpl{
		service: service,
		bot:     bot,
	}
}

// SendMessage отправляет сообщение пользователю.
func (h *UtilityHandlerImpl) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	return err
}

// CalculatePercentage рассчитывает процент от общего числа.
func (h *UtilityHandlerImpl) CalculatePercentage(part, total int) int {
	if total == 0 {
		return 0
	}
	return (part * 100) / total
}

// GetFeedbackNavigationState создает строку состояния навигации для отзывов.
func (h *UtilityHandlerImpl) GetFeedbackNavigationState(userID int64, feedbackType string, currentIndex int) string {
	return fmt.Sprintf("fb_nav_%d_%s_%d", userID, feedbackType, currentIndex)
}

// ParseFeedbackNavigationState извлекает состояние навигации из строки.
func (h *UtilityHandlerImpl) ParseFeedbackNavigationState(stateStr string) (userID int64, feedbackType string, currentIndex int) {
	parts := strings.Split(stateStr, "_")
	if len(parts) >= 4 && parts[0] == "fb" && parts[1] == "nav" {
		userID, _ = strconv.ParseInt(parts[2], 10, 64)
		feedbackType = parts[3]
		if len(parts) >= 5 {
			currentIndex, _ = strconv.Atoi(parts[4])
		}
	}
	return
}
