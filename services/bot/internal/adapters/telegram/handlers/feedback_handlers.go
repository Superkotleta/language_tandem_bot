package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// FeedbackHandler интерфейс для обработчиков отзывов
type FeedbackHandler interface {
	HandleFeedbackCommand(message *tgbotapi.Message, user *models.User) error
	HandleFeedbacksCommand(message *tgbotapi.Message, user *models.User) error
	HandleMainFeedback(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleFeedbackMessage(message *tgbotapi.Message, user *models.User) error
	HandleFeedbackContactMessage(message *tgbotapi.Message, user *models.User) error
	HandleFeedbackProcess(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error
	HandleFeedbackUnprocess(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error
	HandleFeedbackDelete(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error
	HandleShowActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleShowArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleShowAllFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleBrowseActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error
	HandleBrowseArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error
	HandleBrowseAllFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error
	HandleNavigateFeedback(callback *tgbotapi.CallbackQuery, user *models.User, feedbackType string, indexStr string) error
	HandleArchiveFeedback(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error
	HandleBackToFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, feedbackType string) error
	HandleBackToFeedbackStats(callback *tgbotapi.CallbackQuery, user *models.User) error
	editActiveFeedbacks(chatID int64, messageID int, user *models.User) error
	editArchiveFeedbacks(chatID int64, messageID int, user *models.User) error
	editAllFeedbacks(chatID int64, messageID int, user *models.User) error
	editActiveFeedbacksList(chatID int64, messageID int, user *models.User) error
	editArchiveFeedbacksList(chatID int64, messageID int, user *models.User) error
	editAllFeedbacksList(chatID int64, messageID int, user *models.User) error
	HandleDeleteCurrentFeedback(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error
	HandleDeleteAllArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleConfirmDeleteAllArchive(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleUnarchiveFeedback(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error
	HandleFeedbackPrev(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error
	HandleFeedbackNext(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error
	HandleFeedbackBack(callback *tgbotapi.CallbackQuery, user *models.User, feedbackType string) error
}

// FeedbackHandlerImpl реализация обработчиков отзывов
type FeedbackHandlerImpl struct {
	bot             *tgbotapi.BotAPI
	service         *core.BotService
	keyboardBuilder *KeyboardBuilder
	adminChatIDs    []int64
	adminUsernames  []string
}

// NewFeedbackHandler создает новый экземпляр FeedbackHandler
func NewFeedbackHandler(bot *tgbotapi.BotAPI, service *core.BotService, keyboardBuilder *KeyboardBuilder, adminChatIDs []int64, adminUsernames []string) FeedbackHandler {
	return &FeedbackHandlerImpl{
		bot:             bot,
		service:         service,
		keyboardBuilder: keyboardBuilder,
		adminChatIDs:    adminChatIDs,
		adminUsernames:  adminUsernames,
	}
}

// HandleFeedbackCommand обрабатывает команду /feedback
func (fh *FeedbackHandlerImpl) HandleFeedbackCommand(message *tgbotapi.Message, user *models.User) error {
	text := fh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_text")
	_ = fh.service.DB.UpdateUserState(user.ID, models.StateWaitingFeedback)
	return fh.sendMessage(message.Chat.ID, text)
}

// HandleFeedbacksCommand обрабатывает команду /feedbacks (только для администраторов)
func (fh *FeedbackHandlerImpl) HandleFeedbacksCommand(message *tgbotapi.Message, user *models.User) error {
	// Проверяем права администратора по Chat ID и username
	isAdminByID := false
	isAdminByUsername := false

	// Проверяем по Chat ID
	for _, adminID := range fh.adminChatIDs {
		if message.Chat.ID == adminID {
			isAdminByID = true
			break
		}
	}

	// Проверяем по username
	if message.From != nil && message.From.UserName != "" {
		for _, adminUsername := range fh.adminUsernames {
			if message.From.UserName == adminUsername {
				isAdminByUsername = true
				break
			}
		}
	}

	// Если пользователь не является администратором, отправляем сообщение об отказе
	if !isAdminByID && !isAdminByUsername {
		return fh.sendMessage(message.Chat.ID, fh.service.Localizer.Get(user.InterfaceLanguageCode, "access_denied"))
	}

	// Показываем статистику отзывов и меню управления
	return fh.showFeedbackStatistics(message.Chat.ID, user)
}

// HandleMainFeedback обрабатывает нажатие кнопки "Отзыв" в главном меню
func (fh *FeedbackHandlerImpl) HandleMainFeedback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Создаем message объект для handleFeedbackCommand
	message := &tgbotapi.Message{
		Chat: callback.Message.Chat,
	}
	return fh.HandleFeedbackCommand(message, user)
}

// sendMessage отправляет сообщение
func (fh *FeedbackHandlerImpl) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := fh.bot.Send(msg)
	return err
}

// editFeedbackStatistics редактирует сообщение со статистикой отзывов
func (fh *FeedbackHandlerImpl) editFeedbackStatistics(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(chatID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Подсчитываем статистику
	activeCount := 0
	archivedCount := 0
	totalCount := len(allFeedbacks)

	for _, feedback := range allFeedbacks {
		if isArchived, ok := feedback["is_processed"].(bool); ok && isArchived {
			archivedCount++
		} else {
			activeCount++
		}
	}

	// Формируем текст
	text := "📊 Статистика отзывов:\n\n"
	text += fmt.Sprintf("🔥 Активные: %d\n", activeCount)
	text += fmt.Sprintf("📦 Обработанные: %d\n", archivedCount)
	text += fmt.Sprintf("📈 Всего: %d", totalCount)

	// Создаем клавиатуру
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔥 Активные", "show_active_feedbacks"),
			tgbotapi.NewInlineKeyboardButtonData("📦 Обработанные", "show_archive_feedbacks"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Все отзывы", "show_all_feedbacks"),
		),
	)

	// Редактируем сообщение
	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ReplyMarkup = &keyboard
	_, err = fh.bot.Send(editMsg)
	return err
}

// showFeedbackStatistics показывает статистику отзывов
func (fh *FeedbackHandlerImpl) showFeedbackStatistics(chatID int64, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return fh.sendMessage(chatID, "❌ Ошибка загрузки отзывов")
	}

	// Подсчитываем статистику
	activeCount := 0
	archivedCount := 0
	totalCount := len(allFeedbacks)

	for _, fb := range allFeedbacks {
		if isArchived, ok := fb["is_processed"].(bool); ok && isArchived {
			archivedCount++
		} else {
			activeCount++
		}
	}

	// Формируем текст статистики
	text := "📊 Статистика отзывов:\n\n"
	text += fmt.Sprintf("🔥 Активные: %d\n", activeCount)
	text += fmt.Sprintf("📦 Обработанные: %d\n", archivedCount)
	text += fmt.Sprintf("📈 Всего: %d", totalCount)

	// Создаем клавиатуру управления отзывами
	keyboard := fh.keyboardBuilder.CreateFeedbackAdminKeyboard(user.InterfaceLanguageCode)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err = fh.bot.Send(msg)
	return err
}

// sendFeedbackList отправляет список отзывов
func (fh *FeedbackHandlerImpl) sendFeedbackList(chatID int64, feedbackList []map[string]interface{}) error {
	for _, feedback := range feedbackList {
		if err := fh.sendFeedbackItem(chatID, feedback); err != nil {
			return err
		}
	}
	return nil
}

// sendFeedbackListWithPagination отправляет список отзывов с навигацией в одном сообщении
func (fh *FeedbackHandlerImpl) sendFeedbackListWithPagination(chatID int64, feedbackList []map[string]interface{}, feedbackType string) error {
	if len(feedbackList) == 0 {
		return fh.sendMessage(chatID, "📝 Отзывов нет")
	}

	// Показываем первый отзыв с навигацией
	return fh.sendFeedbackWithNavigation(chatID, feedbackList, 0, feedbackType)
}

// sendFeedbackWithNavigation отправляет один отзыв с кнопками навигации
func (fh *FeedbackHandlerImpl) sendFeedbackWithNavigation(chatID int64, feedbackList []map[string]interface{}, currentIndex int, feedbackType string) error {
	if currentIndex < 0 || currentIndex >= len(feedbackList) {
		return fh.sendMessage(chatID, "❌ Неверный индекс отзыва")
	}

	feedback := feedbackList[currentIndex]

	// Формируем текст отзыва
	text := fh.formatFeedbackText(feedback, currentIndex+1, len(feedbackList))

	// Создаем клавиатуру навигации
	keyboard := fh.createNavigationKeyboard(currentIndex, len(feedbackList), feedbackType)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := fh.bot.Send(msg)
	return err
}

// editFeedbackWithNavigation обновляет существующее сообщение с отзывом
func (fh *FeedbackHandlerImpl) editFeedbackWithNavigation(chatID int64, messageID int, feedbackList []map[string]interface{}, currentIndex int, feedbackType string) error {
	if currentIndex < 0 || currentIndex >= len(feedbackList) {
		return fh.sendMessage(chatID, "❌ Неверный индекс отзыва")
	}

	feedback := feedbackList[currentIndex]

	// Формируем текст отзыва
	text := fh.formatFeedbackText(feedback, currentIndex+1, len(feedbackList))

	// Создаем клавиатуру навигации
	keyboard := fh.createNavigationKeyboard(currentIndex, len(feedbackList), feedbackType)

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ReplyMarkup = &keyboard
	editMsg.ParseMode = tgbotapi.ModeHTML
	_, err := fh.bot.Send(editMsg)
	return err
}

// formatFeedbackText форматирует текст отзыва
func (fh *FeedbackHandlerImpl) formatFeedbackText(feedback map[string]interface{}, currentNum, totalCount int) string {
	feedbackID := feedback["id"].(int)
	firstName := feedback["first_name"].(string)
	telegramID := feedback["telegram_id"].(int64)
	feedbackText := feedback["feedback_text"].(string)
	createdAt := feedback["created_at"].(time.Time)

	text := fmt.Sprintf("📝 <b>Отзыв #%d (%d из %d)</b>\n\n", feedbackID, currentNum, totalCount)
	text += fmt.Sprintf("👤 <b>Имя:</b> %s\n", firstName)
	text += fmt.Sprintf("🆔 <b>Telegram ID:</b> %d\n", telegramID)

	// Добавляем username если есть
	if username, ok := feedback["username"].(string); ok && username != "" {
		text += fmt.Sprintf("👤 <b>Username:</b> @%s\n", username)
	}

	text += fmt.Sprintf("📅 <b>Дата:</b> %s\n\n", createdAt.Format("02.01.2006 15:04"))
	text += fmt.Sprintf("💬 <b>Отзыв:</b>\n%s", feedbackText)

	// Добавляем контактную информацию если есть
	if contactInfo, ok := feedback["contact_info"].(*string); ok && contactInfo != nil {
		text += fmt.Sprintf("\n\n📞 <b>Контакты:</b> %s", *contactInfo)
	}

	return text
}

// createNavigationKeyboard создает клавиатуру навигации
func (fh *FeedbackHandlerImpl) createNavigationKeyboard(currentIndex, totalCount int, feedbackType string) tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton

	// Кнопка "Предыдущий"
	if currentIndex > 0 {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			"⬅️ Предыдущий",
			fmt.Sprintf("nav_%s_feedback_%d", feedbackType, currentIndex-1),
		))
	}

	// Кнопка "Следующий"
	if currentIndex < totalCount-1 {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			"➡️ Следующий",
			fmt.Sprintf("nav_%s_feedback_%d", feedbackType, currentIndex+1),
		))
	}

	// Кнопка "В обработанные" (только для активных отзывов)
	if feedbackType == "active" {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			"📦 В обработанные",
			fmt.Sprintf("archive_feedback_%d", currentIndex),
		))
	}

	// Кнопки для архивных отзывов
	if feedbackType == "archive" {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			"🔄 Вернуть в активные",
			fmt.Sprintf("unarchive_feedback_%d", currentIndex),
		))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			"🗑️ Удалить текущий",
			fmt.Sprintf("delete_current_feedback_%d", currentIndex),
		))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			"🗑️ Удалить все",
			"delete_all_archive_feedbacks",
		))
	}

	// Кнопка "Назад к списку"
	buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
		"📋 К списку",
		fmt.Sprintf("back_to_%s_feedbacks", feedbackType),
	))

	// Кнопка "К статистике"
	buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
		"📊 К статистике",
		"back_to_feedback_stats",
	))

	// Разбиваем кнопки на строки
	var rows [][]tgbotapi.InlineKeyboardButton
	if len(buttons) > 0 {
		// Первая строка: навигация
		if len(buttons) >= 2 {
			rows = append(rows, []tgbotapi.InlineKeyboardButton{buttons[0], buttons[1]})
			buttons = buttons[2:]
		} else if len(buttons) == 1 {
			rows = append(rows, []tgbotapi.InlineKeyboardButton{buttons[0]})
			buttons = buttons[1:]
		}

		// Остальные кнопки
		for _, button := range buttons {
			rows = append(rows, []tgbotapi.InlineKeyboardButton{button})
		}
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// sendFeedbackItem отправляет один отзыв
func (fh *FeedbackHandlerImpl) sendFeedbackItem(chatID int64, fb map[string]interface{}) error {
	feedbackID := fb["id"].(int)
	firstName := fb["first_name"].(string)
	feedbackTextContent := strings.ReplaceAll(fb["feedback_text"].(string), "\n", " ")
	charCount := len([]rune(feedbackTextContent))

	// Информация об авторе
	username := "–"
	if fb["username"] != nil {
		username = "@" + fb["username"].(string)
	}

	// Форматируем дату
	createdAt := fb["created_at"].(time.Time)
	dateStr := createdAt.Format("02.01.2006 15:04")

	// Иконка статуса отзыва
	statusIcon := "🏷️"
	statusText := "Ожидает обработки"
	if fb["is_processed"].(bool) {
		statusIcon = "✅"
		statusText = "Обработан"
	}

	// Иконка длины отзыва
	charIcon := "📝"
	if charCount < 50 {
		charIcon = "💬"
	} else if charCount < 200 {
		charIcon = "📝"
	} else {
		charIcon = "📖"
	}

	// Контактная информация
	contactStr := ""
	if fb["contact_info"] != nil && fb["contact_info"].(string) != "" {
		contactStr = fmt.Sprintf("\n🔗 <i>Контакты: %s</i>", fb["contact_info"].(string))
	}

	// Формируем полное объединенное сообщение
	fullMessage := fmt.Sprintf(
		"%s <b>%s</b> %s\n"+
			"👤 <b>Автор:</b> %s\n"+
			"📊 <b>Статус:</b> %s (%d символов)\n"+
			"⏰ <b>Дата:</b> %s%s\n\n"+
			"<b>📨 Содержание отзыва:</b>\n"+
			"<i>%s</i>",
		statusIcon, firstName, username,
		statusText,
		charIcon,
		charCount,
		dateStr,
		contactStr,
		feedbackTextContent,
	)

	// Создаем клавиатуру с кнопками управления
	var buttons [][]tgbotapi.InlineKeyboardButton
	if fb["is_processed"].(bool) {
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("🔄 Вернуть в обработку", fmt.Sprintf("fb_unprocess_%d", feedbackID)),
				tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("fb_delete_%d", feedbackID)),
			},
		}
	} else {
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("✅ Обработан", fmt.Sprintf("fb_process_%d", feedbackID)),
				tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("fb_delete_%d", feedbackID)),
			},
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, fullMessage)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	_, err := fh.bot.Send(msg)
	return err
}

// ========== Заглушки для интерфейса (будут реализованы позже) ==========

// HandleFeedbackMessage обрабатывает сообщение с отзывом
func (fh *FeedbackHandlerImpl) HandleFeedbackMessage(message *tgbotapi.Message, user *models.User) error {
	feedbackText := message.Text

	// Проверяем валидность отзыва
	if len([]rune(feedbackText)) < 10 {
		return fh.handleFeedbackTooShort(message, user)
	}
	if len([]rune(feedbackText)) > 1000 {
		return fh.handleFeedbackTooLong(message, user)
	}

	// Проверяем наличие username
	if user.Username == "" {
		return fh.handleFeedbackContactRequest(message, user, feedbackText)
	}

	// Логируем принятие отзыва
	log.Printf("Отзыв принят: len=%d, has_username=%v", len([]rune(feedbackText)), user.Username != "")

	// Сохраняем полный отзыв и отправляем уведомление
	return fh.handleFeedbackComplete(message, user, feedbackText, nil)
}

// handleFeedbackTooShort обрабатывает слишком короткий отзыв
func (fh *FeedbackHandlerImpl) handleFeedbackTooShort(message *tgbotapi.Message, user *models.User) error {
	feedbackText := message.Text
	count := len([]rune(feedbackText))

	errorText := fmt.Sprintf("%s\n\n%s",
		fh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_too_short"),
		fh.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "feedback_char_count", map[string]string{
			"count": strconv.Itoa(count),
		}),
	)

	return fh.sendMessage(message.Chat.ID, errorText)
}

// handleFeedbackTooLong обрабатывает слишком длинный отзыв
func (fh *FeedbackHandlerImpl) handleFeedbackTooLong(message *tgbotapi.Message, user *models.User) error {
	feedbackText := message.Text
	count := len([]rune(feedbackText))

	errorText := fmt.Sprintf("%s\n\n%s",
		fh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_too_long"),
		fh.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "feedback_char_count", map[string]string{
			"count": strconv.Itoa(count),
		}),
	)

	return fh.sendMessage(message.Chat.ID, errorText)
}

// handleFeedbackContactRequest запрашивает контактные данные при отсутствии username
func (fh *FeedbackHandlerImpl) handleFeedbackContactRequest(message *tgbotapi.Message, user *models.User, feedbackText string) error {
	// Сохраняем отзыв во временном хранилище (в будущем можно добавить в redis/кэш)
	// Пока просто переходим к следующему состоянию

	// Обновляем состояние для ожидания контактных данных
	err := fh.service.DB.UpdateUserState(user.ID, models.StateWaitingFeedbackContact)
	if err != nil {
		return err
	}

	// Запрашиваем контактные данные
	contactText := fh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_request")
	return fh.sendMessage(message.Chat.ID, contactText)
}

// handleFeedbackComplete завершает процесс обратной связи
func (fh *FeedbackHandlerImpl) handleFeedbackComplete(message *tgbotapi.Message, user *models.User, feedbackText string, contactInfo *string) error {
	// Используем ID администраторов из обработчика
	adminIDs := fh.adminChatIDs

	// Сохраняем отзыв через сервис
	err := fh.service.SaveUserFeedback(user.ID, feedbackText, contactInfo, adminIDs)
	if err != nil {
		log.Printf("Ошибка сохранения отзыва: %v", err)
		// Используем локализацию для ошибки
		errorText := fh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_error_generic")
		if errorText == "feedback_error_generic" { // fallback в случае отсутствия перевода
			errorText = "❌ Произошла ошибка при сохранении отзыва. Попробуйте позже."
		}
		return fh.sendMessage(message.Chat.ID, errorText)
	}

	// Отправляем подтверждение пользователю
	successText := fh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_saved")
	if successText == "feedback_saved" { // fallback в случае отсутствия перевода
		successText = "✅ Спасибо за ваш отзыв! Мы обязательно его рассмотрим."
	}

	// Возвращаем пользователя в активное состояние
	err = fh.service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		log.Printf("Ошибка обновления состояния пользователя: %v", err)
	}

	err = fh.service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		log.Printf("Ошибка обновления статуса пользователя: %v", err)
	}

	return fh.sendMessage(message.Chat.ID, successText)
}

// HandleFeedbackContactMessage обрабатывает сообщение с контактными данными
func (fh *FeedbackHandlerImpl) HandleFeedbackContactMessage(message *tgbotapi.Message, user *models.User) error {
	contactInfo := strings.TrimSpace(message.Text)

	// Валидируем контактные данные
	if contactInfo == "" {
		return fh.sendMessage(message.Chat.ID,
			fh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_placeholder"))
	}

	// Подтверждаем получение контактов
	confirmedText := fh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_provided")
	fh.sendMessage(message.Chat.ID, confirmedText)

	// Теперь нужно получить сохраненный отзыв пользователя
	// Пока что используем временное решение - просим написать отзыв заново
	// В будущем здесь будет получение из кэша

	feedbackText := "Отзыв был сохранен в предыдущем шаге (требуется интеграция с кэшем)" // временное решение

	return fh.handleFeedbackComplete(message, user, feedbackText, &contactInfo)
}

// HandleFeedbackProcess обрабатывает отметку отзыва как обработанного
func (fh *FeedbackHandlerImpl) HandleFeedbackProcess(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error {
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка идентификатора отзыва")
	}

	// Обновляем статус отзыва как обработанный
	err = fh.service.UpdateFeedbackStatus(feedbackID, true)
	if err != nil {
		log.Printf("Ошибка обновления статуса отзыва %d: %v", feedbackID, err)
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка обновления статуса")
	}

	// Отправляем обновление администратору
	confirmMsg := fmt.Sprintf("✅ Отзыв #%d отмечен как <b>обработанный</b>", feedbackID)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, confirmMsg)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := fh.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки подтверждения обработки: %v", err)
	}

	return nil
}

// HandleFeedbackUnprocess возвращает отзыв в необработанные
func (fh *FeedbackHandlerImpl) HandleFeedbackUnprocess(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error {
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка идентификатора отзыва")
	}

	// Возвращаем отзыв в необработанный статус
	err = fh.service.UpdateFeedbackStatus(feedbackID, false)
	if err != nil {
		log.Printf("Ошибка возврата отзыва в обработку %d: %v", feedbackID, err)
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка возврата статуса")
	}

	// Отправляем обновление администратору
	confirmMsg := fmt.Sprintf("🔄 Отзыв #%d возвращен в <b>обработку</b>", feedbackID)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, confirmMsg)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := fh.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки подтверждения возврата: %v", err)
	}

	return nil
}

// HandleFeedbackDelete удаляет отзыв
func (fh *FeedbackHandlerImpl) HandleFeedbackDelete(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error {
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка идентификатора отзыва")
	}

	// Удаляем отзыв
	err = fh.service.DeleteFeedback(feedbackID)
	if err != nil {
		log.Printf("Ошибка удаления отзыва %d: %v", feedbackID, err)
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка удаления отзыва")
	}

	// Отправляем подтверждение удаления
	deleteMsg := fmt.Sprintf("🗑️ Отзыв #%d <b>удален</b>", feedbackID)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, deleteMsg)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := fh.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки подтверждения удаления: %v", err)
	}

	return nil
}

// HandleShowActiveFeedbacks показывает активные отзывы
func (fh *FeedbackHandlerImpl) HandleShowActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все отзывы
	feedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		return fh.sendMessage(callback.Message.Chat.ID, "📝 Отзывов пока нет")
	}

	// Фильтруем только активные отзывы (не архивные)
	var activeFeedbacks []map[string]interface{}
	for _, fb := range feedbacks {
		if isArchived, ok := fb["is_processed"].(bool); !ok || !isArchived {
			activeFeedbacks = append(activeFeedbacks, fb)
		}
	}

	if len(activeFeedbacks) == 0 {
		return fh.sendMessage(callback.Message.Chat.ID, "🎉 Все отзывы в архиве!")
	}

	// Показываем первый отзыв с навигацией (редактируем существующее сообщение)
	return fh.editFeedbackWithNavigation(callback.Message.Chat.ID, callback.Message.MessageID, activeFeedbacks, 0, "active")
}

// HandleShowArchiveFeedbacks показывает архивные отзывы
func (fh *FeedbackHandlerImpl) HandleShowArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все отзывы
	feedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		return fh.sendMessage(callback.Message.Chat.ID, "📝 Отзывов пока нет")
	}

	// Фильтруем только архивные отзывы
	var archivedFeedbacks []map[string]interface{}
	for _, fb := range feedbacks {
		if isArchived, ok := fb["is_processed"].(bool); ok && isArchived {
			archivedFeedbacks = append(archivedFeedbacks, fb)
		}
	}

	if len(archivedFeedbacks) == 0 {
		return fh.sendMessage(callback.Message.Chat.ID, "📦 Архив пуст")
	}

	// Показываем первый отзыв с навигацией (редактируем существующее сообщение)
	return fh.editFeedbackWithNavigation(callback.Message.Chat.ID, callback.Message.MessageID, archivedFeedbacks, 0, "archive")
}

// HandleShowAllFeedbacks показывает все отзывы
func (fh *FeedbackHandlerImpl) HandleShowAllFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все отзывы
	feedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		return fh.sendMessage(callback.Message.Chat.ID, "📝 Отзывов пока нет")
	}

	// Показываем первый отзыв с навигацией (редактируем существующее сообщение)
	return fh.editFeedbackWithNavigation(callback.Message.Chat.ID, callback.Message.MessageID, feedbacks, 0, "all")
}

// HandleBrowseActiveFeedbacks просматривает активные отзывы
func (fh *FeedbackHandlerImpl) HandleBrowseActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	return fh.handleBrowseFeedbacks(callback, user, indexStr, "active")
}

// HandleBrowseArchiveFeedbacks просматривает архивные отзывы
func (fh *FeedbackHandlerImpl) HandleBrowseArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	return fh.handleBrowseFeedbacks(callback, user, indexStr, "archive")
}

// HandleBrowseAllFeedbacks просматривает все отзывы
func (fh *FeedbackHandlerImpl) HandleBrowseAllFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	return fh.handleBrowseFeedbacks(callback, user, indexStr, "all")
}

// handleBrowseFeedbacks общая функция для навигации по отзывам
func (fh *FeedbackHandlerImpl) handleBrowseFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error {
	// Парсим индекс
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка в параметрах")
	}

	// Получаем все отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	// Фильтруем отзывы по типу
	var feedbacks []map[string]interface{}
	switch feedbackType {
	case "active":
		for _, fb := range allFeedbacks {
			if isArchived, ok := fb["is_processed"].(bool); !ok || !isArchived {
				feedbacks = append(feedbacks, fb)
			}
		}
	case "archive":
		for _, fb := range allFeedbacks {
			if isArchived, ok := fb["is_processed"].(bool); ok && isArchived {
				feedbacks = append(feedbacks, fb)
			}
		}
	case "all":
		feedbacks = allFeedbacks
	}

	if len(feedbacks) == 0 {
		return fh.sendMessage(callback.Message.Chat.ID, "📝 Отзывов нет")
	}

	if index < 0 || index >= len(feedbacks) {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Неверный индекс отзыва")
	}

	// Показываем отзыв с навигацией (редактируем существующее сообщение)
	return fh.editFeedbackWithNavigation(callback.Message.Chat.ID, callback.Message.MessageID, feedbacks, index, feedbackType)
}

// HandleNavigateFeedback обрабатывает навигацию по отзывам
func (fh *FeedbackHandlerImpl) HandleNavigateFeedback(callback *tgbotapi.CallbackQuery, user *models.User, feedbackType string, indexStr string) error {
	return fh.handleBrowseFeedbacks(callback, user, indexStr, feedbackType)
}

// HandleArchiveFeedback архивирует отзыв
func (fh *FeedbackHandlerImpl) HandleArchiveFeedback(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	// Получаем все активные отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	// Фильтруем активные отзывы
	var activeFeedbacks []map[string]interface{}
	for _, fb := range allFeedbacks {
		if isArchived, ok := fb["is_processed"].(bool); !ok || !isArchived {
			activeFeedbacks = append(activeFeedbacks, fb)
		}
	}

	// Парсим индекс
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(activeFeedbacks) {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Неверный индекс отзыва")
	}

	// Получаем ID отзыва для архивирования
	feedback := activeFeedbacks[index]
	feedbackID := feedback["id"].(int)

	// Архивируем отзыв
	err = fh.service.ArchiveFeedback(feedbackID)
	if err != nil {
		log.Printf("Ошибка архивирования отзыва: %v", err)
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка архивирования отзыва")
	}

	// Обновляем список активных отзывов
	activeFeedbacks = append(activeFeedbacks[:index], activeFeedbacks[index+1:]...)

	// Показываем следующий отзыв или сообщение об отсутствии отзывов
	if len(activeFeedbacks) == 0 {
		// Редактируем сообщение, показывая что все отзывы обработаны
		text := "✅ Отзыв обработан!\n\n🎉 Все отзывы обработаны!"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		editMsg.ReplyMarkup = &keyboard
		_, err := fh.bot.Send(editMsg)
		return err
	}

	// Показываем следующий отзыв (или предыдущий, если это был последний)
	nextIndex := index
	if nextIndex >= len(activeFeedbacks) {
		nextIndex = len(activeFeedbacks) - 1
	}

	return fh.editFeedbackWithNavigation(callback.Message.Chat.ID, callback.Message.MessageID, activeFeedbacks, nextIndex, "active")
}

// HandleBackToFeedbacks возвращает к списку отзывов
func (fh *FeedbackHandlerImpl) HandleBackToFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, feedbackType string) error {
	switch feedbackType {
	case "active":
		return fh.editActiveFeedbacksList(callback.Message.Chat.ID, callback.Message.MessageID, user)
	case "archive":
		return fh.editArchiveFeedbacksList(callback.Message.Chat.ID, callback.Message.MessageID, user)
	case "all":
		return fh.editAllFeedbacksList(callback.Message.Chat.ID, callback.Message.MessageID, user)
	default:
		return fh.editFeedbackStatistics(callback.Message.Chat.ID, callback.Message.MessageID, user)
	}
}

// HandleBackToFeedbackStats возвращает к статистике отзывов
func (fh *FeedbackHandlerImpl) HandleBackToFeedbackStats(callback *tgbotapi.CallbackQuery, user *models.User) error {
	return fh.editFeedbackStatistics(callback.Message.Chat.ID, callback.Message.MessageID, user)
}

// editActiveFeedbacks редактирует сообщение со списком активных отзывов
func (fh *FeedbackHandlerImpl) editActiveFeedbacks(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(chatID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Фильтруем только активные отзывы
	var activeFeedbacks []map[string]interface{}
	for _, feedback := range allFeedbacks {
		if isArchived, ok := feedback["is_processed"].(bool); !ok || !isArchived {
			activeFeedbacks = append(activeFeedbacks, feedback)
		}
	}

	// Проверяем, есть ли активные отзывы
	if len(activeFeedbacks) == 0 {
		// Показываем сообщение об отсутствии активных отзывов
		text := "🎉 Все отзывы обработаны!"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
		editMsg.ReplyMarkup = &keyboard
		_, err := fh.bot.Send(editMsg)
		return err
	}

	// Показываем первый отзыв с навигацией
	return fh.editFeedbackWithNavigation(chatID, messageID, activeFeedbacks, 0, "active")
}

// editArchiveFeedbacks редактирует сообщение со списком обработанных отзывов
func (fh *FeedbackHandlerImpl) editArchiveFeedbacks(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(chatID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Фильтруем только обработанные отзывы
	var archiveFeedbacks []map[string]interface{}
	for _, feedback := range allFeedbacks {
		if isArchived, ok := feedback["is_processed"].(bool); ok && isArchived {
			archiveFeedbacks = append(archiveFeedbacks, feedback)
		}
	}

	// Проверяем, есть ли обработанные отзывы
	if len(archiveFeedbacks) == 0 {
		// Показываем сообщение об отсутствии обработанных отзывов
		text := "📦 Обработанных отзывов пока нет"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
		editMsg.ReplyMarkup = &keyboard
		_, err := fh.bot.Send(editMsg)
		return err
	}

	// Показываем первый отзыв с навигацией
	return fh.editFeedbackWithNavigation(chatID, messageID, archiveFeedbacks, 0, "archive")
}

// editAllFeedbacks редактирует сообщение со списком всех отзывов
func (fh *FeedbackHandlerImpl) editAllFeedbacks(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(chatID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Проверяем, есть ли отзывы
	if len(allFeedbacks) == 0 {
		// Показываем сообщение об отсутствии отзывов
		text := "📝 Отзывов пока нет"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
		editMsg.ReplyMarkup = &keyboard
		_, err := fh.bot.Send(editMsg)
		return err
	}

	// Показываем первый отзыв с навигацией
	return fh.editFeedbackWithNavigation(chatID, messageID, allFeedbacks, 0, "all")
}

// editActiveFeedbacksList редактирует сообщение со списком активных отзывов (заголовок)
func (fh *FeedbackHandlerImpl) editActiveFeedbacksList(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(chatID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Фильтруем только активные отзывы
	var activeFeedbacks []map[string]interface{}
	for _, feedback := range allFeedbacks {
		if isArchived, ok := feedback["is_processed"].(bool); !ok || !isArchived {
			activeFeedbacks = append(activeFeedbacks, feedback)
		}
	}

	// Проверяем, есть ли активные отзывы
	if len(activeFeedbacks) == 0 {
		// Показываем сообщение об отсутствии активных отзывов
		text := "🎉 Все отзывы обработаны!"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
		editMsg.ReplyMarkup = &keyboard
		_, err := fh.bot.Send(editMsg)
		return err
	}

	// Показываем заголовок списка активных отзывов
	text := fmt.Sprintf("🔥 <b>Активные отзывы (%d):</b>\n\nВыберите отзыв для просмотра:", len(activeFeedbacks))

	// Создаем клавиатуру с кнопками для каждого отзыва
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, feedback := range activeFeedbacks {
		feedbackID := feedback["id"].(int)
		firstName := feedback["first_name"].(string)

		buttonText := fmt.Sprintf("📝 %s (ID: %d)", firstName, feedbackID)
		if username, ok := feedback["username"].(string); ok && username != "" {
			buttonText = fmt.Sprintf("📝 %s (@%s) (ID: %d)", firstName, username, feedbackID)
		}
		buttonData := fmt.Sprintf("nav_active_feedback_%d", i)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonData),
		))
	}

	// Добавляем кнопки навигации
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ReplyMarkup = &keyboard
	editMsg.ParseMode = tgbotapi.ModeHTML
	_, err = fh.bot.Send(editMsg)
	return err
}

// editArchiveFeedbacksList редактирует сообщение со списком обработанных отзывов (заголовок)
func (fh *FeedbackHandlerImpl) editArchiveFeedbacksList(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(chatID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Фильтруем только обработанные отзывы
	var archiveFeedbacks []map[string]interface{}
	for _, feedback := range allFeedbacks {
		if isArchived, ok := feedback["is_processed"].(bool); ok && isArchived {
			archiveFeedbacks = append(archiveFeedbacks, feedback)
		}
	}

	// Проверяем, есть ли обработанные отзывы
	if len(archiveFeedbacks) == 0 {
		// Показываем сообщение об отсутствии обработанных отзывов
		text := "📦 Обработанных отзывов пока нет"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
		editMsg.ReplyMarkup = &keyboard
		_, err := fh.bot.Send(editMsg)
		return err
	}

	// Показываем заголовок списка обработанных отзывов
	text := fmt.Sprintf("📦 <b>Обработанные отзывы (%d):</b>\n\nВыберите отзыв для просмотра:", len(archiveFeedbacks))

	// Создаем клавиатуру с кнопками для каждого отзыва
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, feedback := range archiveFeedbacks {
		feedbackID := feedback["id"].(int)
		firstName := feedback["first_name"].(string)

		buttonText := fmt.Sprintf("📝 %s (ID: %d)", firstName, feedbackID)
		if username, ok := feedback["username"].(string); ok && username != "" {
			buttonText = fmt.Sprintf("📝 %s (@%s) (ID: %d)", firstName, username, feedbackID)
		}
		buttonData := fmt.Sprintf("nav_archive_feedback_%d", i)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonData),
		))
	}

	// Добавляем кнопки навигации
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ReplyMarkup = &keyboard
	editMsg.ParseMode = tgbotapi.ModeHTML
	_, err = fh.bot.Send(editMsg)
	return err
}

// editAllFeedbacksList редактирует сообщение со списком всех отзывов (заголовок)
func (fh *FeedbackHandlerImpl) editAllFeedbacksList(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(chatID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Проверяем, есть ли отзывы
	if len(allFeedbacks) == 0 {
		// Показываем сообщение об отсутствии отзывов
		text := "📝 Отзывов пока нет"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
		editMsg.ReplyMarkup = &keyboard
		_, err := fh.bot.Send(editMsg)
		return err
	}

	// Показываем заголовок списка всех отзывов
	text := fmt.Sprintf("📋 <b>Все отзывы (%d):</b>\n\nВыберите отзыв для просмотра:", len(allFeedbacks))

	// Создаем клавиатуру с кнопками для каждого отзыва
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, feedback := range allFeedbacks {
		feedbackID := feedback["id"].(int)
		firstName := feedback["first_name"].(string)
		isProcessed := feedback["is_processed"].(bool)

		status := "🔥"
		if isProcessed {
			status = "📦"
		}

		buttonText := fmt.Sprintf("%s %s (ID: %d)", status, firstName, feedbackID)
		if username, ok := feedback["username"].(string); ok && username != "" {
			buttonText = fmt.Sprintf("%s %s (@%s) (ID: %d)", status, firstName, username, feedbackID)
		}
		buttonData := fmt.Sprintf("nav_all_feedback_%d", i)

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonData),
		))
	}

	// Добавляем кнопки навигации
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ReplyMarkup = &keyboard
	editMsg.ParseMode = tgbotapi.ModeHTML
	_, err = fh.bot.Send(editMsg)
	return err
}

// HandleDeleteCurrentFeedback удаляет текущий отзыв
func (fh *FeedbackHandlerImpl) HandleDeleteCurrentFeedback(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	// Получаем все обработанные отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Фильтруем только обработанные отзывы
	var archiveFeedbacks []map[string]interface{}
	for _, feedback := range allFeedbacks {
		if isProcessed, ok := feedback["is_processed"].(bool); ok && isProcessed {
			archiveFeedbacks = append(archiveFeedbacks, feedback)
		}
	}

	// Парсим индекс
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(archiveFeedbacks) {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Неверный индекс отзыва")
	}

	// Получаем ID отзыва для удаления
	feedbackID := archiveFeedbacks[index]["id"].(int)

	// Удаляем отзыв
	err = fh.service.DeleteFeedback(feedbackID)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка удаления отзыва: "+err.Error())
	}

	// Обновляем список (удаляем удаленный отзыв)
	archiveFeedbacks = append(archiveFeedbacks[:index], archiveFeedbacks[index+1:]...)

	// Показываем следующий отзыв или сообщение об отсутствии отзывов
	if len(archiveFeedbacks) == 0 {
		// Редактируем сообщение, показывая что все отзывы удалены
		text := "✅ Отзыв удален!\n\n🎉 Все обработанные отзывы удалены!"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		editMsg.ReplyMarkup = &keyboard
		_, err := fh.bot.Send(editMsg)
		return err
	}

	// Показываем следующий отзыв (или предыдущий, если это был последний)
	nextIndex := index
	if nextIndex >= len(archiveFeedbacks) {
		nextIndex = len(archiveFeedbacks) - 1
	}

	return fh.editFeedbackWithNavigation(callback.Message.Chat.ID, callback.Message.MessageID, archiveFeedbacks, nextIndex, "archive")
}

// HandleDeleteAllArchiveFeedbacks показывает подтверждение удаления всех обработанных отзывов
func (fh *FeedbackHandlerImpl) HandleDeleteAllArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем количество обработанных отзывов
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Подсчитываем обработанные отзывы
	processedCount := 0
	for _, feedback := range allFeedbacks {
		if isProcessed, ok := feedback["is_processed"].(bool); ok && isProcessed {
			processedCount++
		}
	}

	if processedCount == 0 {
		return fh.sendMessage(callback.Message.Chat.ID, "📦 Нет обработанных отзывов для удаления")
	}

	// Показываем подтверждение
	text := fmt.Sprintf("⚠️ <b>Подтверждение удаления</b>\n\nВы действительно хотите удалить <b>%d обработанных отзывов</b>?\n\n❗️ <b>Это действие нельзя отменить!</b>", processedCount)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить все", "confirm_delete_all_archive"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "back_to_archive_feedbacks"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ReplyMarkup = &keyboard
	editMsg.ParseMode = tgbotapi.ModeHTML
	_, err = fh.bot.Send(editMsg)
	return err
}

// HandleConfirmDeleteAllArchive подтверждает и выполняет удаление всех обработанных отзывов
func (fh *FeedbackHandlerImpl) HandleConfirmDeleteAllArchive(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Удаляем все обработанные отзывы
	deletedCount, err := fh.service.DeleteAllProcessedFeedbacks()
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка удаления отзывов: "+err.Error())
	}

	// Показываем результат
	text := fmt.Sprintf("✅ <b>Удаление завершено!</b>\n\n🗑️ Удалено отзывов: <b>%d</b>\n\n📊 Все обработанные отзывы удалены из базы данных.", deletedCount)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	editMsg.ReplyMarkup = &keyboard
	editMsg.ParseMode = tgbotapi.ModeHTML
	_, err = fh.bot.Send(editMsg)
	return err
}

// HandleUnarchiveFeedback возвращает отзыв в активные
func (fh *FeedbackHandlerImpl) HandleUnarchiveFeedback(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	// Получаем все обработанные отзывы
	allFeedbacks, err := fh.service.GetAllFeedback()
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов: "+err.Error())
	}

	// Фильтруем только обработанные отзывы
	var archiveFeedbacks []map[string]interface{}
	for _, feedback := range allFeedbacks {
		if isProcessed, ok := feedback["is_processed"].(bool); ok && isProcessed {
			archiveFeedbacks = append(archiveFeedbacks, feedback)
		}
	}

	// Парсим индекс
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(archiveFeedbacks) {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Неверный индекс отзыва")
	}

	// Получаем ID отзыва для возврата в активные
	feedbackID := archiveFeedbacks[index]["id"].(int)

	// Возвращаем отзыв в активные
	err = fh.service.UnarchiveFeedback(feedbackID)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка возврата отзыва в активные: "+err.Error())
	}

	// Обновляем список (удаляем возвращенный отзыв)
	archiveFeedbacks = append(archiveFeedbacks[:index], archiveFeedbacks[index+1:]...)

	// Показываем следующий отзыв или сообщение об отсутствии отзывов
	if len(archiveFeedbacks) == 0 {
		// Редактируем сообщение, показывая что все отзывы возвращены
		text := "✅ Отзыв возвращен в активные!\n\n🎉 Все обработанные отзывы возвращены!"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		editMsg.ReplyMarkup = &keyboard
		_, err := fh.bot.Send(editMsg)
		return err
	}

	// Показываем следующий отзыв (или предыдущий, если это был последний)
	nextIndex := index
	if nextIndex >= len(archiveFeedbacks) {
		nextIndex = len(archiveFeedbacks) - 1
	}

	return fh.editFeedbackWithNavigation(callback.Message.Chat.ID, callback.Message.MessageID, archiveFeedbacks, nextIndex, "archive")
}

// HandleFeedbackPrev переходит к предыдущему отзыву
func (fh *FeedbackHandlerImpl) HandleFeedbackPrev(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error {
	// TODO: Реализовать позже - навигация назад
	return fh.sendMessage(callback.Message.Chat.ID, "⬅️ Предыдущий отзыв (в разработке)")
}

// HandleFeedbackNext переходит к следующему отзыву
func (fh *FeedbackHandlerImpl) HandleFeedbackNext(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error {
	// TODO: Реализовать позже - навигация вперед
	return fh.sendMessage(callback.Message.Chat.ID, "➡️ Следующий отзыв (в разработке)")
}

// HandleFeedbackBack возвращается к списку отзывов
func (fh *FeedbackHandlerImpl) HandleFeedbackBack(callback *tgbotapi.CallbackQuery, user *models.User, feedbackType string) error {
	// TODO: Реализовать позже - возврат к списку
	return fh.sendMessage(callback.Message.Chat.ID, "🔙 Назад к списку (в разработке)")
}
