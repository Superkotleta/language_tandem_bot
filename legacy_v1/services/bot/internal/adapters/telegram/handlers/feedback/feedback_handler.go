package feedback

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	"language-exchange-bot/internal/adapters/telegram/handlers/base"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Feedback handler constants are now defined in localization/constants.go

// FeedbackHandler интерфейс для обработчиков отзывов.
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

// FeedbackHandlerImpl реализация обработчиков отзывов.
type FeedbackHandlerImpl struct {
	base           *base.BaseHandler
	adminChatIDs   []int64
	adminUsernames []string
}

// NewFeedbackHandler создает новый экземпляр FeedbackHandler.
func NewFeedbackHandler(
	base *base.BaseHandler,
	adminChatIDs []int64,
	adminUsernames []string,
) *FeedbackHandlerImpl {
	return &FeedbackHandlerImpl{
		base:           base,
		adminChatIDs:   adminChatIDs,
		adminUsernames: adminUsernames,
	}
}

// HandleFeedbackCommand обрабатывает команду /feedback.
func (fh *FeedbackHandlerImpl) HandleFeedbackCommand(message *tgbotapi.Message, user *models.User) error {
	text := fh.base.Service.Localizer.Get(user.InterfaceLanguageCode, "feedback_text")
	if err := fh.base.Service.DB.UpdateUserState(user.ID, models.StateWaitingFeedback); err != nil {
		log.Printf("Failed to update user state to waiting feedback for user %d: %v", user.ID, err)
	}

	return fh.sendMessage(message.Chat.ID, text)
}

// HandleFeedbacksCommand обрабатывает команду /feedbacks (только для администраторов).

// HandleMainFeedback обрабатывает нажатие кнопки "Отзыв" в главном меню.
func (fh *FeedbackHandlerImpl) HandleMainFeedback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Создаем message объект для handleFeedbackCommand
	message := &tgbotapi.Message{
		Chat: callback.Message.Chat,
	}

	return fh.HandleFeedbackCommand(message, user)
}

// sendMessage отправляет сообщение (deprecated - используйте messageFactory.SendText).
func (fh *FeedbackHandlerImpl) sendMessage(chatID int64, text string) error {
	return fh.base.MessageFactory.SendText(chatID, text)
}

// editFeedbackStatistics редактирует сообщение со статистикой отзывов.
func (fh *FeedbackHandlerImpl) editFeedbackStatistics(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.Service.GetAllFeedback()
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
	err = fh.base.MessageFactory.EditWithKeyboard(chatID, messageID, text, &keyboard)

	return err
}

// showFeedbackStatistics показывает статистику отзывов.
func (fh *FeedbackHandlerImpl) showFeedbackStatistics(chatID int64, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.Service.GetAllFeedback()
	if err != nil {
		// Используем структурированное логирование
		fh.base.Service.LoggingService.Database().ErrorWithContext(
			"Failed to get feedbacks",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			chatID,
			"GetAllFeedback",
			map[string]interface{}{
				"error": err.Error(),
			},
		)

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
	keyboard := fh.base.KeyboardBuilder.CreateFeedbackAdminKeyboard(user.InterfaceLanguageCode)

	// Используем MessageFactory для отправки сообщения
	return fh.base.MessageFactory.SendWithKeyboard(chatID, text, keyboard)
}

// editFeedbackWithNavigation обновляет существующее сообщение с отзывом.
func (fh *FeedbackHandlerImpl) editFeedbackWithNavigation(
	chatID int64,
	messageID int,
	feedbackList []map[string]interface{},
	currentIndex int,
	feedbackType string,
) error {
	if currentIndex < 0 || currentIndex >= len(feedbackList) {
		return fh.sendMessage(chatID, "❌ Неверный индекс отзыва")
	}

	feedback := feedbackList[currentIndex]

	// Формируем текст отзыва
	text := fh.formatFeedbackText(feedback, currentIndex+1, len(feedbackList))

	// Создаем клавиатуру навигации
	keyboard := fh.createNavigationKeyboard(currentIndex, len(feedbackList), feedbackType)

	err := fh.base.MessageFactory.EditHTMLWithKeyboard(chatID, messageID, text, &keyboard)

	return err
}

// formatFeedbackText форматирует текст отзыва.
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
	text += "💬 <b>Отзыв:</b>\n" + feedbackText

	// Добавляем контактную информацию если есть
	if contactInfo, ok := feedback["contact_info"].(*string); ok && contactInfo != nil {
		text += "\n\n📞 <b>Контакты:</b> " + *contactInfo
	}

	return text
}

// createNavigationKeyboard создает клавиатуру навигации.
//
//nolint:funlen
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
	if feedbackType == localization.FeedbackTypeActiveLocal {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(
			"📦 В обработанные",
			fmt.Sprintf("archive_feedback_%d", currentIndex),
		))
	}

	// Кнопки для архивных отзывов
	if feedbackType == localization.FeedbackTypeArchiveLocal {
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
		if len(buttons) >= localization.ButtonsPerRow {
			rows = append(rows, []tgbotapi.InlineKeyboardButton{buttons[0], buttons[1]})
			buttons = buttons[localization.ButtonsPerRow:]
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

// ========== Заглушки для интерфейса (будут реализованы позже) ==========

// HandleFeedbackMessage обрабатывает сообщение с отзывом.
func (fh *FeedbackHandlerImpl) HandleFeedbackMessage(message *tgbotapi.Message, user *models.User) error {
	feedbackText := message.Text

	// Проверяем валидность отзыва
	if len([]rune(feedbackText)) < localization.MinFeedbackLength {
		return fh.handleFeedbackTooShort(message, user)
	}

	if len([]rune(feedbackText)) > localization.MaxFeedbackItems {
		return fh.handleFeedbackTooLong(message, user)
	}

	// Проверяем наличие username
	if user.Username == "" {
		return fh.handleFeedbackContactRequest(message, user, feedbackText)
	}

	// Логируем принятие отзыва
	fh.base.Service.LoggingService.Telegram().InfoWithContext(
		"Feedback received",
		"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
		int64(user.ID),
		message.Chat.ID,
		"HandleFeedbackMessage",
		map[string]interface{}{
			"text_length":  len([]rune(feedbackText)),
			"has_username": user.Username != "",
			"username":     user.Username,
		},
	)

	// Сохраняем полный отзыв и отправляем уведомление
	return fh.handleFeedbackComplete(message, user, feedbackText, nil)
}

// handleFeedbackTooShort обрабатывает слишком короткий отзыв.
func (fh *FeedbackHandlerImpl) handleFeedbackTooShort(message *tgbotapi.Message, user *models.User) error {
	feedbackText := message.Text
	count := len([]rune(feedbackText))

	errorText := fmt.Sprintf("%s\n\n%s",
		fh.base.Service.Localizer.Get(user.InterfaceLanguageCode, "feedback_too_short"),
		fh.base.Service.Localizer.GetWithParams(user.InterfaceLanguageCode, "feedback_char_count", map[string]string{
			"count": strconv.Itoa(count),
		}),
	)

	return fh.sendMessage(message.Chat.ID, errorText)
}

// handleFeedbackTooLong обрабатывает слишком длинный отзыв.
func (fh *FeedbackHandlerImpl) handleFeedbackTooLong(message *tgbotapi.Message, user *models.User) error {
	feedbackText := message.Text
	count := len([]rune(feedbackText))

	errorText := fmt.Sprintf("%s\n\n%s",
		fh.base.Service.Localizer.Get(user.InterfaceLanguageCode, "feedback_too_long"),
		fh.base.Service.Localizer.GetWithParams(user.InterfaceLanguageCode, "feedback_char_count", map[string]string{
			"count": strconv.Itoa(count),
		}),
	)

	return fh.sendMessage(message.Chat.ID, errorText)
}

// handleFeedbackContactRequest запрашивает контактные данные при отсутствии username.
func (fh *FeedbackHandlerImpl) handleFeedbackContactRequest(message *tgbotapi.Message, user *models.User, feedbackText string) error {
	// Сохраняем отзыв во временном хранилище (в будущем можно добавить в redis/кэш)
	// Пока просто переходим к следующему состоянию

	// Обновляем состояние для ожидания контактных данных
	err := fh.base.Service.DB.UpdateUserState(user.ID, models.StateWaitingFeedbackContact)
	if err != nil {
		return err
	}

	// Запрашиваем контактные данные
	contactText := fh.base.Service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_request")

	return fh.sendMessage(
		message.Chat.ID,
		contactText,
	)
}

// handleFeedbackComplete завершает процесс обратной связи.
func (fh *FeedbackHandlerImpl) handleFeedbackComplete(message *tgbotapi.Message, user *models.User, feedbackText string, contactInfo *string) error {
	// Используем ID администраторов из обработчика
	adminIDs := fh.adminChatIDs

	// Сохраняем отзыв через сервис
	err := fh.base.Service.SaveUserFeedback(user.ID, feedbackText, contactInfo, adminIDs)
	if err != nil {
		// Используем структурированное логирование
		fh.base.Service.LoggingService.Database().ErrorWithContext(
			"Failed to save feedback",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			message.Chat.ID,
			"SaveUserFeedback",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
		// Используем локализацию для ошибки
		errorText := fh.base.Service.Localizer.Get(user.InterfaceLanguageCode, "feedback_error_generic")
		if errorText == "feedback_error_generic" { // fallback в случае отсутствия перевода
			errorText = "❌ Произошла ошибка при сохранении отзыва. Попробуйте позже."
		}

		return fh.sendMessage(message.Chat.ID, errorText)
	}

	// Отправляем подтверждение пользователю
	successText := fh.base.Service.Localizer.Get(user.InterfaceLanguageCode, "feedback_saved")
	if successText == "feedback_saved" { // fallback в случае отсутствия перевода
		successText = "✅ Спасибо за ваш отзыв! Мы обязательно его рассмотрим."
	}

	// Возвращаем пользователя в активное состояние
	err = fh.base.Service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		// Используем структурированное логирование
		fh.base.Service.LoggingService.Database().ErrorWithContext(
			"Failed to update user state",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			message.Chat.ID,
			"UpdateUserState",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	err = fh.base.Service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		// Используем структурированное логирование
		fh.base.Service.LoggingService.Database().ErrorWithContext(
			"Failed to update user status",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			message.Chat.ID,
			"UpdateUserStatus",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	// Создаем клавиатуру с кнопкой "Главное меню"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			fh.base.KeyboardBuilder.CreateBackToMainButton(user.InterfaceLanguageCode),
		),
	)

	// Используем MessageFactory для отправки сообщения
	if err := fh.base.MessageFactory.SendWithKeyboard(message.Chat.ID, successText, keyboard); err != nil {
		// Используем структурированное логирование
		fh.base.Service.LoggingService.Database().ErrorWithContext(
			"Failed to send success message",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			message.Chat.ID,
			"SendMessage",
			map[string]interface{}{
				"error": err.Error(),
			},
		)

		return err
	}

	return nil
}

// HandleFeedbackContactMessage обрабатывает сообщение с контактными данными.
func (fh *FeedbackHandlerImpl) HandleFeedbackContactMessage(message *tgbotapi.Message, user *models.User) error {
	contactInfo := strings.TrimSpace(message.Text)

	// Валидируем контактные данные
	if contactInfo == "" {
		return fh.sendMessage(message.Chat.ID,
			fh.base.Service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_placeholder"))
	}

	// Подтверждаем получение контактов
	confirmedText := fh.base.Service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_provided")
	if err := fh.sendMessage(message.Chat.ID, confirmedText); err != nil {
		return err
	}

	// Теперь нужно получить сохраненный отзыв пользователя
	// Пока что используем временное решение - просим написать отзыв заново
	// В будущем здесь будет получение из кэша

	feedbackText := "Отзыв был сохранен в предыдущем шаге (требуется интеграция с кэшем)" // временное решение

	return fh.handleFeedbackComplete(message, user, feedbackText, &contactInfo)
}
