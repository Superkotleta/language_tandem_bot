package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AdminHandler интерфейс для обработки административных функций
type AdminHandler interface {
	ShowFeedbackStatisticsEdit(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleBrowseActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error
	HandleBrowseArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error
	ShowFeedbackItemWithNavigation(chatID int64, fb map[string]interface{}, currentIndex int, totalCount int, feedbackType string) error
	ShowFeedbackItemWithNavigationEdit(callback *tgbotapi.CallbackQuery, fb map[string]interface{}, currentIndex int, totalCount int, feedbackType string) error
	IsAdmin(chatID int64, username string) bool
}

// AdminHandlerImpl реализация административного обработчика
type AdminHandlerImpl struct {
	service         *core.BotService
	bot             *tgbotapi.BotAPI
	keyboardBuilder *KeyboardBuilder
	adminChatIDs    []int64
	adminUsernames  []string
}

// NewAdminHandler создает новый административный обработчик
func NewAdminHandler(service *core.BotService, bot *tgbotapi.BotAPI, keyboardBuilder *KeyboardBuilder, adminChatIDs []int64, adminUsernames []string) AdminHandler {
	return &AdminHandlerImpl{
		service:         service,
		bot:             bot,
		keyboardBuilder: keyboardBuilder,
		adminChatIDs:    adminChatIDs,
		adminUsernames:  adminUsernames,
	}
}

// IsAdmin проверяет права администратора
func (h *AdminHandlerImpl) IsAdmin(chatID int64, username string) bool {
	// Проверяем по Chat ID
	for _, adminID := range h.adminChatIDs {
		if chatID == adminID {
			return true
		}
	}

	// Проверяем по username
	if username != "" {
		for _, adminUsername := range h.adminUsernames {
			cleanUsername := strings.TrimPrefix(adminUsername, "@")
			if username == cleanUsername {
				return true
			}
		}
	}

	return false
}

// ShowFeedbackStatisticsEdit показывает статистику отзывов с редактированием текущего сообщения
func (h *AdminHandlerImpl) ShowFeedbackStatisticsEdit(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Проверяем права администратора
	if !h.IsAdmin(callback.Message.Chat.ID, user.Username) {
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ Данная команда доступна только администраторам бота.")
		_, err := h.bot.Send(msg)
		return err
	}

	// Получаем все отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
		_, err := h.bot.Send(msg)
		return err
	}

	if len(feedbacks) == 0 {
		editMsg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"📝 Отзывов пока нет",
		)
		_, err := h.bot.Request(editMsg)
		return err
	}

	// Подсчитываем статистику
	totalCount := len(feedbacks)
	processedCount := 0
	for _, fb := range feedbacks {
		if fb["is_processed"].(bool) {
			processedCount++
		}
	}
	pendingCount := totalCount - processedCount

	// Формируем текст статистики
	statsText := fmt.Sprintf("📊 Статистика отзывов:\n\n"+
		"📝 Всего отзывов: %d\n"+
		"✅ Обработано: %d\n"+
		"⏳ Ожидают обработки: %d",
		totalCount, processedCount, pendingCount)

	// Создаем клавиатуру для управления отзывами
	keyboard := h.keyboardBuilder.CreateFeedbackAdminKeyboard(user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		statsText,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	return err
}

// HandleBrowseActiveFeedbacks показывает активные отзывы в интерактивном режиме с редактированием
func (h *AdminHandlerImpl) HandleBrowseActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		log.Printf("Ошибка парсинга индекса: %v", err)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ Ошибка индекса")
		_, err := h.bot.Send(msg)
		return err
	}

	// Получаем активные отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
		_, err := h.bot.Send(msg)
		return err
	}

	// Фильтруем только необработанные отзывы
	var activeFeedbacks []map[string]interface{}
	type feedbackKey struct {
		userID       int64
		feedbackText string
	}
	seen := make(map[feedbackKey][]map[string]interface{})
	for _, fb := range feedbacks {
		if !fb["is_processed"].(bool) {
			key := feedbackKey{
				userID:       fb["telegram_id"].(int64),
				feedbackText: fb["feedback_text"].(string),
			}
			seen[key] = append(seen[key], fb)
		}
	}

	for _, group := range seen {
		for _, fb := range group {
			activeFeedbacks = append(activeFeedbacks, fb)
			break
		}
	}

	if len(activeFeedbacks) == 0 {
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "🎉 Все отзывы обработаны!")
		_, err := h.bot.Send(msg)
		return err
	}

	// Проверяем границы
	if index < 0 || index >= len(activeFeedbacks) {
		index = 0
	}

	// Показываем текущий отзыв с редактированием текущего сообщения
	return h.ShowFeedbackItemWithNavigationEdit(callback, activeFeedbacks[index], index, len(activeFeedbacks), "active")
}

// HandleBrowseArchiveFeedbacks показывает обработанные отзывы в интерактивном режиме
func (h *AdminHandlerImpl) HandleBrowseArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		log.Printf("Ошибка парсинга индекса: %v", err)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ Ошибка индекса")
		_, err := h.bot.Send(msg)
		return err
	}

	// Получаем обработанные отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
		_, err := h.bot.Send(msg)
		return err
	}

	// Фильтруем только обработанные отзывы
	var archivedFeedbacks []map[string]interface{}
	type feedbackKey struct {
		userID       int64
		feedbackText string
	}
	seen := make(map[feedbackKey][]map[string]interface{})
	for _, fb := range feedbacks {
		if fb["is_processed"].(bool) {
			key := feedbackKey{
				userID:       fb["telegram_id"].(int64),
				feedbackText: fb["feedback_text"].(string),
			}
			seen[key] = append(seen[key], fb)
		}
	}

	for _, group := range seen {
		for _, fb := range group {
			archivedFeedbacks = append(archivedFeedbacks, fb)
			break
		}
	}

	if len(archivedFeedbacks) == 0 {
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "📝 Обработанных отзывов пока нет")
		_, err := h.bot.Send(msg)
		return err
	}

	// Проверяем границы
	if index < 0 || index >= len(archivedFeedbacks) {
		index = 0
	}

	// Показываем текущий отзыв с редактированием текущего сообщения
	return h.ShowFeedbackItemWithNavigationEdit(callback, archivedFeedbacks[index], index, len(archivedFeedbacks), "archive")
}

// ShowFeedbackItemWithNavigation показывает отзыв с навигацией (отправляет новое сообщение)
func (h *AdminHandlerImpl) ShowFeedbackItemWithNavigation(chatID int64, fb map[string]interface{}, currentIndex int, totalCount int, feedbackType string) error {
	// Формируем текст отзыва
	feedbackText := h.formatFeedbackText(fb, currentIndex, totalCount, feedbackType)

	// Создаем клавиатуру навигации
	keyboard := h.createFeedbackNavigationKeyboard(fb, currentIndex, totalCount, feedbackType)

	msg := tgbotapi.NewMessage(chatID, feedbackText)
	msg.ReplyMarkup = keyboard
	_, err := h.bot.Send(msg)
	return err
}

// ShowFeedbackItemWithNavigationEdit показывает отзыв с навигацией (редактирует текущее сообщение)
func (h *AdminHandlerImpl) ShowFeedbackItemWithNavigationEdit(callback *tgbotapi.CallbackQuery, fb map[string]interface{}, currentIndex int, totalCount int, feedbackType string) error {
	// Формируем текст отзыва
	feedbackText := h.formatFeedbackText(fb, currentIndex, totalCount, feedbackType)

	// Создаем клавиатуру навигации
	keyboard := h.createFeedbackNavigationKeyboard(fb, currentIndex, totalCount, feedbackType)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		feedbackText,
		keyboard,
	)
	_, err := h.bot.Request(editMsg)
	return err
}

// formatFeedbackText форматирует текст отзыва для отображения
func (h *AdminHandlerImpl) formatFeedbackText(fb map[string]interface{}, currentIndex int, totalCount int, feedbackType string) string {
	userID := fb["telegram_id"].(int64)
	username := ""
	if fb["username"] != nil {
		username = fb["username"].(string)
	}
	firstName := ""
	if fb["first_name"] != nil {
		firstName = fb["first_name"].(string)
	}
	feedbackText := fb["feedback_text"].(string)
	contactInfo := ""
	if fb["contact_info"] != nil && fb["contact_info"].(string) != "" {
		contactInfo = fb["contact_info"].(string)
	}

	// Определяем статус и тип
	status := "❌ Не обработан"
	if fb["is_processed"].(bool) {
		status = "✅ Обработан"
	}

	typeText := "📝 Активные отзывы"
	if feedbackType == "archive" {
		typeText = "📁 Архив отзывов"
	} else if feedbackType == "all" {
		typeText = "📊 Все отзывы"
	}

	// Формируем информацию о пользователе
	userInfo := fmt.Sprintf("ID: %d", userID)
	if username != "" {
		userInfo += fmt.Sprintf(" (@%s)", username)
	}
	if firstName != "" {
		userInfo += fmt.Sprintf(" - %s", firstName)
	}

	// Формируем полный текст
	text := fmt.Sprintf("%s (%d/%d)\n\n"+
		"👤 Пользователь: %s\n"+
		"📝 Отзыв: %s\n"+
		"📊 Статус: %s",
		typeText, currentIndex+1, totalCount,
		userInfo, feedbackText, status)

	if contactInfo != "" {
		text += fmt.Sprintf("\n📞 Контакт: %s", contactInfo)
	}

	return text
}

// createFeedbackNavigationKeyboard создает клавиатуру навигации для отзывов
func (h *AdminHandlerImpl) createFeedbackNavigationKeyboard(fb map[string]interface{}, currentIndex int, totalCount int, feedbackType string) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Кнопки навигации
	if totalCount > 1 {
		var navRow []tgbotapi.InlineKeyboardButton
		if currentIndex > 0 {
			navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ Пред", fmt.Sprintf("browse_%s_%d", feedbackType, currentIndex-1)))
		}
		if currentIndex < totalCount-1 {
			navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("След ➡️", fmt.Sprintf("browse_%s_%d", feedbackType, currentIndex+1)))
		}
		if len(navRow) > 0 {
			keyboard = append(keyboard, navRow)
		}
	}

	// Кнопки действий для необработанных отзывов
	if feedbackType == "active" && !fb["is_processed"].(bool) {
		actionRow := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("✅ Обработать", fmt.Sprintf("fb_process_%v", fb["id"])),
		}
		keyboard = append(keyboard, actionRow)
	}

	// Кнопки действий для обработанных отзывов
	if feedbackType == "archive" && fb["is_processed"].(bool) {
		actionRow := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("❌ Снять обработку", fmt.Sprintf("fb_unprocess_%v", fb["id"])),
		}
		keyboard = append(keyboard, actionRow)
	}

	// Кнопка "Назад к статистике"
	backRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔙 К статистике", "feedback_stats"),
	}
	keyboard = append(keyboard, backRow)

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}
