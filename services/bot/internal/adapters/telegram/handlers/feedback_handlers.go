package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

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
	base           *BaseHandler
	adminChatIDs   []int64
	adminUsernames []string
}

// NewFeedbackHandler создает новый экземпляр FeedbackHandler.
func NewFeedbackHandler(
	base *BaseHandler,
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
	text := fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_text")
	if err := fh.base.service.DB.UpdateUserState(user.ID, models.StateWaitingFeedback); err != nil {
		log.Printf("Failed to update user state to waiting feedback for user %d: %v", user.ID, err)
	}

	return fh.sendMessage(message.Chat.ID, text)
}

// HandleFeedbacksCommand обрабатывает команду /feedbacks (только для администраторов).
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
		return fh.sendMessage(message.Chat.ID, fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "access_denied"))
	}

	// Показываем статистику отзывов и меню управления
	return fh.showFeedbackStatistics(message.Chat.ID, user)
}

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
	return fh.base.messageFactory.SendText(chatID, text)
}

// editFeedbackStatistics редактирует сообщение со статистикой отзывов.
func (fh *FeedbackHandlerImpl) editFeedbackStatistics(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
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
	err = fh.base.messageFactory.EditWithKeyboard(chatID, messageID, text, &keyboard)

	return err
}

// showFeedbackStatistics показывает статистику отзывов.
func (fh *FeedbackHandlerImpl) showFeedbackStatistics(chatID int64, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
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
	keyboard := fh.base.keyboardBuilder.CreateFeedbackAdminKeyboard(user.InterfaceLanguageCode)

	// Используем MessageFactory для отправки сообщения
	return fh.base.messageFactory.SendWithKeyboard(chatID, text, keyboard)
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

	err := fh.base.messageFactory.EditHTMLWithKeyboard(chatID, messageID, text, &keyboard)

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
	fh.base.service.LoggingService.Telegram().InfoWithContext(
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
		fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_too_short"),
		fh.base.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "feedback_char_count", map[string]string{
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
		fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_too_long"),
		fh.base.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "feedback_char_count", map[string]string{
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
	err := fh.base.service.DB.UpdateUserState(user.ID, models.StateWaitingFeedbackContact)
	if err != nil {
		return err
	}

	// Запрашиваем контактные данные
	contactText := fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_request")

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
	err := fh.base.service.SaveUserFeedback(user.ID, feedbackText, contactInfo, adminIDs)
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
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
		errorText := fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_error_generic")
		if errorText == "feedback_error_generic" { // fallback в случае отсутствия перевода
			errorText = "❌ Произошла ошибка при сохранении отзыва. Попробуйте позже."
		}

		return fh.sendMessage(message.Chat.ID, errorText)
	}

	// Отправляем подтверждение пользователю
	successText := fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_saved")
	if successText == "feedback_saved" { // fallback в случае отсутствия перевода
		successText = "✅ Спасибо за ваш отзыв! Мы обязательно его рассмотрим."
	}

	// Возвращаем пользователя в активное состояние
	err = fh.base.service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
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

	err = fh.base.service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
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
			tgbotapi.NewInlineKeyboardButtonData(
				fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "main_menu_title"),
				"back_to_main_menu",
			),
		),
	)

	// Используем MessageFactory для отправки сообщения
	if err := fh.base.messageFactory.SendWithKeyboard(message.Chat.ID, successText, keyboard); err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
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
			fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_placeholder"))
	}

	// Подтверждаем получение контактов
	confirmedText := fh.base.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_provided")
	if err := fh.sendMessage(message.Chat.ID, confirmedText); err != nil {
		return err
	}

	// Теперь нужно получить сохраненный отзыв пользователя
	// Пока что используем временное решение - просим написать отзыв заново
	// В будущем здесь будет получение из кэша

	feedbackText := "Отзыв был сохранен в предыдущем шаге (требуется интеграция с кэшем)" // временное решение

	return fh.handleFeedbackComplete(message, user, feedbackText, &contactInfo)
}

// changeFeedbackStatus изменяет статус отзыва и отправляет подтверждение.
func (fh *FeedbackHandlerImpl) changeFeedbackStatus(callback *tgbotapi.CallbackQuery, user *models.User, feedbackID int, processed bool, confirmMsg string) error {
	// Обновляем статус отзыва
	err := fh.base.service.UpdateFeedbackStatus(feedbackID, processed)
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
			"Failed to update feedback status",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"UpdateFeedbackStatus",
			map[string]interface{}{
				"feedback_id": feedbackID,
				"error":       err.Error(),
			},
		)

		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка обновления статуса")
	}

	// Используем MessageFactory для отправки HTML сообщения
	if err := fh.base.messageFactory.SendHTML(callback.Message.Chat.ID, confirmMsg); err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Telegram().ErrorWithContext(
			"Failed to send status change confirmation",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"SendStatusChangeConfirmation",
			map[string]interface{}{
				"feedback_id": feedbackID,
				"error":       err.Error(),
			},
		)
	}

	return nil
}

// processArchiveFeedbackAction обрабатывает действия над архивными отзывами.
//
//nolint:cyclop // функция содержит последовательную логику обработки, сложность оправдана
func (fh *FeedbackHandlerImpl) processArchiveFeedbackAction(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, actionFunc func(int) error, successMessage string) error {
	// Получаем все обработанные отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
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

	// Получаем ID отзыва
	feedbackID := archiveFeedbacks[index]["id"].(int)

	// Выполняем действие
	err = actionFunc(feedbackID)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ "+err.Error())
	}

	// Обновляем список (удаляем обработанный отзыв)
	archiveFeedbacks = append(archiveFeedbacks[:index], archiveFeedbacks[index+1:]...)

	// Показываем следующий отзыв или сообщение об отсутствии отзывов
	if len(archiveFeedbacks) == 0 {
		// Редактируем сообщение
		text := successMessage
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 К статистике", "back_to_feedback_stats"),
			),
		)

		err = fh.base.messageFactory.EditWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, text, &keyboard)

		return err
	}

	// Показываем следующий отзыв (или предыдущий, если это был последний)
	nextIndex := index
	if nextIndex >= len(archiveFeedbacks) {
		nextIndex = len(archiveFeedbacks) - 1
	}

	return fh.editFeedbackWithNavigation(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		archiveFeedbacks,
		nextIndex,
		"archive",
	)
}

// HandleFeedbackProcess обрабатывает отметку отзыва как обработанного.
func (fh *FeedbackHandlerImpl) HandleFeedbackProcess(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error {
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка идентификатора отзыва")
	}

	confirmMsg := fmt.Sprintf("✅ Отзыв #%d отмечен как <b>обработанный</b>", feedbackID)

	return fh.changeFeedbackStatus(callback, user, feedbackID, true, confirmMsg)
}

// HandleFeedbackUnprocess возвращает отзыв в необработанные.
func (fh *FeedbackHandlerImpl) HandleFeedbackUnprocess(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error {
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка идентификатора отзыва")
	}

	confirmMsg := fmt.Sprintf("🔄 Отзыв #%d возвращен в <b>обработку</b>", feedbackID)

	return fh.changeFeedbackStatus(callback, user, feedbackID, false, confirmMsg)
}

// HandleFeedbackDelete удаляет отзыв.
func (fh *FeedbackHandlerImpl) HandleFeedbackDelete(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error {
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка идентификатора отзыва")
	}

	// Удаляем отзыв
	err = fh.base.service.DeleteFeedback(feedbackID)
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
			"Failed to delete feedback",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"DeleteFeedback",
			map[string]interface{}{
				"feedback_id": feedbackID,
				"error":       err.Error(),
			},
		)

		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка удаления отзыва")
	}

	// Используем MessageFactory для отправки HTML сообщения
	deleteMsg := fmt.Sprintf("🗑️ Отзыв #%d <b>удален</b>", feedbackID)
	if err := fh.base.messageFactory.SendHTML(callback.Message.Chat.ID, deleteMsg); err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Telegram().ErrorWithContext(
			"Failed to send deletion confirmation",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"SendDeletionConfirmation",
			map[string]interface{}{
				"feedback_id": feedbackID,
				"error":       err.Error(),
			},
		)
	}

	return nil
}

// HandleShowActiveFeedbacks показывает активные отзывы.
func (fh *FeedbackHandlerImpl) HandleShowActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все отзывы
	feedbacks, err := fh.base.service.GetAllFeedback()
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
			"Failed to get feedbacks",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"GetAllFeedback",
			map[string]interface{}{
				"error": err.Error(),
			},
		)

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

// HandleShowArchiveFeedbacks показывает архивные отзывы.
func (fh *FeedbackHandlerImpl) HandleShowArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все отзывы
	feedbacks, err := fh.base.service.GetAllFeedback()
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
			"Failed to get feedbacks",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"GetAllFeedback",
			map[string]interface{}{
				"error": err.Error(),
			},
		)

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

// HandleShowAllFeedbacks показывает все отзывы.
func (fh *FeedbackHandlerImpl) HandleShowAllFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все отзывы
	feedbacks, err := fh.base.service.GetAllFeedback()
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
			"Failed to get feedbacks",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"GetAllFeedback",
			map[string]interface{}{
				"error": err.Error(),
			},
		)

		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		return fh.sendMessage(callback.Message.Chat.ID, "📝 Отзывов пока нет")
	}

	// Показываем первый отзыв с навигацией (редактируем существующее сообщение)
	return fh.editFeedbackWithNavigation(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		feedbacks,
		0,
		"all",
	)
}

// HandleBrowseActiveFeedbacks просматривает активные отзывы.
func (fh *FeedbackHandlerImpl) HandleBrowseActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	return fh.handleBrowseFeedbacks(callback, user, indexStr, "active")
}

// HandleBrowseArchiveFeedbacks просматривает архивные отзывы.
func (fh *FeedbackHandlerImpl) HandleBrowseArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	return fh.handleBrowseFeedbacks(callback, user, indexStr, "archive")
}

// HandleBrowseAllFeedbacks просматривает все отзывы.
func (fh *FeedbackHandlerImpl) HandleBrowseAllFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	return fh.handleBrowseFeedbacks(callback, user, indexStr, "all")
}

// handleBrowseFeedbacks общая функция для навигации по отзывам.
func (fh *FeedbackHandlerImpl) handleBrowseFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error {
	// Парсим индекс
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fh.sendMessage(callback.Message.Chat.ID, "❌ Ошибка в параметрах")
	}

	// Получаем все отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
			"Failed to get feedbacks",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"GetAllFeedback",
			map[string]interface{}{
				"error": err.Error(),
			},
		)

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
	return fh.editFeedbackWithNavigation(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		feedbacks,
		index,
		feedbackType,
	)
}

// HandleNavigateFeedback обрабатывает навигацию по отзывам.
func (fh *FeedbackHandlerImpl) HandleNavigateFeedback(callback *tgbotapi.CallbackQuery, user *models.User, feedbackType string, indexStr string) error {
	return fh.handleBrowseFeedbacks(callback, user, indexStr, feedbackType)
}

// HandleArchiveFeedback архивирует отзыв.
//
//nolint:cyclop,funlen // функция содержит последовательную логику архивирования, длина оправдана
func (fh *FeedbackHandlerImpl) HandleArchiveFeedback(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	// Получаем все активные отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
			"Failed to get feedbacks",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"GetAllFeedback",
			map[string]interface{}{
				"error": err.Error(),
			},
		)

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
	err = fh.base.service.ArchiveFeedback(feedbackID)
	if err != nil {
		// Используем структурированное логирование
		fh.base.service.LoggingService.Database().ErrorWithContext(
			"Failed to archive feedback",
			"req_"+strconv.FormatInt(time.Now().UnixNano(), 10),
			int64(user.ID),
			callback.Message.Chat.ID,
			"ArchiveFeedback",
			map[string]interface{}{
				"feedback_id": feedbackID,
				"error":       err.Error(),
			},
		)

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

		err = fh.base.messageFactory.EditWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, text, &keyboard)

		return err
	}

	// Показываем следующий отзыв (или предыдущий, если это был последний)
	nextIndex := index
	if nextIndex >= len(activeFeedbacks) {
		nextIndex = len(activeFeedbacks) - 1
	}

	return fh.editFeedbackWithNavigation(callback.Message.Chat.ID, callback.Message.MessageID, activeFeedbacks, nextIndex, "active")
}

// HandleBackToFeedbacks возвращает к списку отзывов.
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

// HandleBackToFeedbackStats возвращает к статистике отзывов.
func (fh *FeedbackHandlerImpl) HandleBackToFeedbackStats(callback *tgbotapi.CallbackQuery, user *models.User) error {
	return fh.editFeedbackStatistics(callback.Message.Chat.ID, callback.Message.MessageID, user)
}

// editActiveFeedbacks редактирует сообщение со списком активных отзывов.
func (fh *FeedbackHandlerImpl) editActiveFeedbacks(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
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
		_, err := fh.base.bot.Send(editMsg)

		return err
	}

	// Показываем первый отзыв с навигацией
	return fh.editFeedbackWithNavigation(chatID, messageID, activeFeedbacks, 0, "active")
}

// editArchiveFeedbacks редактирует сообщение со списком обработанных отзывов.
func (fh *FeedbackHandlerImpl) editArchiveFeedbacks(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
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
		_, err := fh.base.bot.Send(editMsg)

		return err
	}

	// Показываем первый отзыв с навигацией
	return fh.editFeedbackWithNavigation(chatID, messageID, archiveFeedbacks, 0, "archive")
}

// editAllFeedbacks редактирует сообщение со списком всех отзывов.
func (fh *FeedbackHandlerImpl) editAllFeedbacks(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
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
		_, err := fh.base.bot.Send(editMsg)

		return err
	}

	// Показываем первый отзыв с навигацией
	return fh.editFeedbackWithNavigation(chatID, messageID, allFeedbacks, 0, "all")
}

// editActiveFeedbacksList редактирует сообщение со списком активных отзывов (заголовок).
func (fh *FeedbackHandlerImpl) editActiveFeedbacksList(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
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

		sendErr := fh.base.messageFactory.EditWithKeyboard(chatID, messageID, text, &keyboard)

		return sendErr
	}

	// Показываем заголовок списка активных отзывов
	text := fmt.Sprintf("🔥 <b>Активные отзывы (%d):</b>\n\nВыберите отзыв для просмотра:", len(activeFeedbacks))

	// Создаем клавиатуру с кнопками для каждого отзыва
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(activeFeedbacks))

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

	err = fh.base.messageFactory.EditHTMLWithKeyboard(chatID, messageID, text, &keyboard)

	return err
}

// editArchiveFeedbacksList редактирует сообщение со списком обработанных отзывов (заголовок).
func (fh *FeedbackHandlerImpl) editArchiveFeedbacksList(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
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

		sendErr := fh.base.messageFactory.EditWithKeyboard(chatID, messageID, text, &keyboard)

		return sendErr
	}

	// Показываем заголовок списка обработанных отзывов
	text := fmt.Sprintf("📦 <b>Обработанные отзывы (%d):</b>\n\nВыберите отзыв для просмотра:", len(archiveFeedbacks))

	// Создаем клавиатуру с кнопками для каждого отзыва
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(archiveFeedbacks))

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

	err = fh.base.messageFactory.EditHTMLWithKeyboard(chatID, messageID, text, &keyboard)

	return err
}

// editAllFeedbacksList редактирует сообщение со списком всех отзывов (заголовок).
func (fh *FeedbackHandlerImpl) editAllFeedbacksList(chatID int64, messageID int, user *models.User) error {
	// Получаем все отзывы
	allFeedbacks, err := fh.base.service.GetAllFeedback()
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

		sendErr := fh.base.messageFactory.EditWithKeyboard(chatID, messageID, text, &keyboard)

		return sendErr
	}

	// Показываем заголовок списка всех отзывов
	text := fmt.Sprintf("📋 <b>Все отзывы (%d):</b>\n\nВыберите отзыв для просмотра:", len(allFeedbacks))

	// Создаем клавиатуру с кнопками для каждого отзыва
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(allFeedbacks))

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

	err = fh.base.messageFactory.EditHTMLWithKeyboard(chatID, messageID, text, &keyboard)

	return err
}

// HandleDeleteCurrentFeedback удаляет текущий отзыв.
func (fh *FeedbackHandlerImpl) HandleDeleteCurrentFeedback(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	return fh.processArchiveFeedbackAction(
		callback,
		user,
		indexStr,
		fh.base.service.DeleteFeedback,
		"✅ Отзыв удален!\n\n🎉 Все обработанные отзывы удалены!",
	)
}

// HandleDeleteAllArchiveFeedbacks показывает подтверждение удаления всех обработанных отзывов.
func (fh *FeedbackHandlerImpl) HandleDeleteAllArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем количество обработанных отзывов
	allFeedbacks, err := fh.base.service.GetAllFeedback()
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

	err = fh.base.messageFactory.EditHTMLWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, text, &keyboard)

	return err
}

// HandleConfirmDeleteAllArchive подтверждает и выполняет удаление всех обработанных отзывов.
func (fh *FeedbackHandlerImpl) HandleConfirmDeleteAllArchive(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Удаляем все обработанные отзывы
	deletedCount, err := fh.base.service.DeleteAllProcessedFeedbacks()
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

	err = fh.base.messageFactory.EditHTMLWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, text, &keyboard)

	return err
}

// HandleUnarchiveFeedback возвращает отзыв в активные.
func (fh *FeedbackHandlerImpl) HandleUnarchiveFeedback(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	return fh.processArchiveFeedbackAction(
		callback,
		user,
		indexStr,
		fh.base.service.UnarchiveFeedback,
		"✅ Отзыв возвращен в активные!\n\n🎉 Все обработанные отзывы возвращены!",
	)
}

// HandleFeedbackPrev переходит к предыдущему отзыву.
func (fh *FeedbackHandlerImpl) HandleFeedbackPrev(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error {
	// TODO: Реализовать позже - навигация назад
	return fh.sendMessage(callback.Message.Chat.ID, "⬅️ Предыдущий отзыв (в разработке)")
}

// HandleFeedbackNext переходит к следующему отзыву.
func (fh *FeedbackHandlerImpl) HandleFeedbackNext(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error {
	// TODO: Реализовать позже - навигация вперед
	return fh.sendMessage(callback.Message.Chat.ID, "➡️ Следующий отзыв (в разработке)")
}

// HandleFeedbackBack возвращается к списку отзывов.
func (fh *FeedbackHandlerImpl) HandleFeedbackBack(callback *tgbotapi.CallbackQuery, user *models.User, feedbackType string) error {
	// TODO: Реализовать позже - возврат к списку
	return fh.sendMessage(callback.Message.Chat.ID, "🔙 Назад к списку (в разработке)")
}
