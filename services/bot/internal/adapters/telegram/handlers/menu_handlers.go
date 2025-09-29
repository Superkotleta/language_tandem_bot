package handlers

import (
	"fmt"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MenuHandler обрабатывает команды и действия главного меню
type MenuHandler struct {
	bot             *tgbotapi.BotAPI
	service         *core.BotService
	keyboardBuilder *KeyboardBuilder
	errorHandler    *errors.ErrorHandler
}

// NewMenuHandler создает новый экземпляр MenuHandler
func NewMenuHandler(bot *tgbotapi.BotAPI, service *core.BotService, keyboardBuilder *KeyboardBuilder, errorHandler *errors.ErrorHandler) *MenuHandler {
	return &MenuHandler{
		bot:             bot,
		service:         service,
		keyboardBuilder: keyboardBuilder,
		errorHandler:    errorHandler,
	}
}

// HandleStartCommand обрабатывает команду /start
func (mh *MenuHandler) HandleStartCommand(message *tgbotapi.Message, user *models.User) error {
	// Всегда показываем главное меню, независимо от состояния профиля
	welcomeText := mh.service.GetWelcomeMessage(user)
	menuText := welcomeText + "\n\n" + mh.service.Localizer.Get(user.InterfaceLanguageCode, "main_menu_title")

	msg := tgbotapi.NewMessage(message.Chat.ID, menuText)
	hasProfile := user.ProfileCompletionLevel > 0
	msg.ReplyMarkup = mh.keyboardBuilder.CreateMainMenuKeyboard(user.InterfaceLanguageCode, hasProfile)

	if _, err := mh.bot.Send(msg); err != nil {
		// Используем новую систему обработки ошибок
		return mh.errorHandler.HandleTelegramError(
			err,
			message.Chat.ID,
			int64(user.ID),
			"HandleStartCommand",
		)
	}

	return nil
}

// HandleStatusCommand обрабатывает команду /status
func (mh *MenuHandler) HandleStatusCommand(message *tgbotapi.Message, user *models.User) error {
	statusText := fmt.Sprintf(
		"📊 %s:\n\n"+
			"🆔 ID: %d\n"+
			"📝 %s: %s\n"+
			"🔄 %s: %s\n"+
			"📈 %s: %d%%\n"+
			"🌐 %s: %s",
		mh.service.Localizer.Get(user.InterfaceLanguageCode, "your_status"),
		user.ID,
		mh.service.Localizer.Get(user.InterfaceLanguageCode, "status"),
		user.Status,
		mh.service.Localizer.Get(user.InterfaceLanguageCode, "state"),
		user.State,
		mh.service.Localizer.Get(user.InterfaceLanguageCode, "profile_completion"),
		user.ProfileCompletionLevel,
		mh.service.Localizer.Get(user.InterfaceLanguageCode, "interface_language"),
		user.InterfaceLanguageCode,
	)
	return mh.sendMessage(message.Chat.ID, statusText)
}

// HandleResetCommand обрабатывает команду /reset
func (mh *MenuHandler) HandleResetCommand(message *tgbotapi.Message, user *models.User) error {
	return mh.sendMessage(message.Chat.ID, mh.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset"))
}

// HandleLanguageCommand обрабатывает команду /language
func (mh *MenuHandler) HandleLanguageCommand(message *tgbotapi.Message, user *models.User) error {
	text := mh.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interface_language")
	keyboard := mh.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "interface", "", true)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	_, err := mh.bot.Send(msg)
	return err
}

// HandleBackToMainMenu возвращает пользователя в главное меню
func (mh *MenuHandler) HandleBackToMainMenu(callback *tgbotapi.CallbackQuery, user *models.User) error {
	welcomeText := mh.service.GetWelcomeMessage(user)
	menuText := welcomeText + "\n\n" + mh.service.Localizer.Get(user.InterfaceLanguageCode, "main_menu_title")

	hasProfile := user.ProfileCompletionLevel > 0
	keyboard := mh.keyboardBuilder.CreateMainMenuKeyboard(user.InterfaceLanguageCode, hasProfile)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		menuText,
		keyboard,
	)
	_, err := mh.bot.Request(editMsg)
	return err
}

// HandleMainChangeLanguage обрабатывает смену языка интерфейса
func (mh *MenuHandler) HandleMainChangeLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := mh.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interface_language")
	keyboard := mh.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "interface", "", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := mh.bot.Request(editMsg)
	return err
}

// HandleMainViewProfile обрабатывает просмотр профиля
func (mh *MenuHandler) HandleMainViewProfile(callback *tgbotapi.CallbackQuery, user *models.User, profileHandler *ProfileHandlerImpl) error {
	// Проверяем, заполнен ли профиль по уровню завершения профиля
	if user.ProfileCompletionLevel == 0 {
		// Профиль не заполнен - показываем информационное сообщение и кнопку настройки
		text := mh.service.Localizer.Get(user.InterfaceLanguageCode, "empty_profile_message")

		// Создаем клавиатуру с кнопкой настройки профиля
		setupButton := tgbotapi.NewInlineKeyboardButtonData(
			mh.service.Localizer.Get(user.InterfaceLanguageCode, "setup_profile_button"),
			"start_profile_setup",
		)

		keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{setupButton})

		// Редактируем существующее сообщение вместо создания нового
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			keyboard,
		)
		_, err := mh.bot.Request(editMsg)
		return err
	}

	// Профиль заполнен - показываем его
	return profileHandler.HandleProfileShow(callback, user)
}

// HandleMainEditProfile обрабатывает редактирование профиля
func (mh *MenuHandler) HandleMainEditProfile(callback *tgbotapi.CallbackQuery, user *models.User, profileHandler *ProfileHandlerImpl) error {
	return profileHandler.HandleProfileResetAsk(callback, user)
}

// HandleMainFeedback обрабатывает переход к отзывам
func (mh *MenuHandler) HandleMainFeedback(callback *tgbotapi.CallbackQuery, user *models.User, feedbackHandler FeedbackHandler) error {
	// Получаем текст обратной связи
	text := mh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_text")

	// Создаем клавиатуру для обратной связи
	keyboard := mh.createFeedbackKeyboard(user.InterfaceLanguageCode)

	// Редактируем сообщение вместо отправки нового
	return mh.editMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, &keyboard)
}

// HandleFeedbackHelp обрабатывает помощь по обратной связи
func (mh *MenuHandler) HandleFeedbackHelp(callback *tgbotapi.CallbackQuery, user *models.User) error {
	helpTitle := mh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_help_title")
	helpContent := mh.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_help_content")
	helpText := helpTitle + "\n\n" + helpContent

	keyboard := mh.createFeedbackKeyboard(user.InterfaceLanguageCode)

	// Редактируем сообщение с помощью
	return mh.editMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, helpText, &keyboard)
}

// createFeedbackKeyboard создает клавиатуру для обратной связи
func (mh *MenuHandler) createFeedbackKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	keyboard := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData(mh.service.Localizer.Get(lang, "feedback_back_to_main"), "back_to_main_menu"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData(mh.service.Localizer.Get(lang, "feedback_help"), "feedback_help"),
		},
	}
	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// editMessageTextAndMarkup редактирует сообщение с клавиатурой
func (mh *MenuHandler) editMessageTextAndMarkup(chatID int64, messageID int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	if keyboard != nil {
		edit.ReplyMarkup = keyboard
	}
	_, err := mh.bot.Send(edit)
	if err != nil {
		return mh.errorHandler.HandleTelegramError(
			err,
			chatID,
			0, // UserID неизвестен в этом контексте
			"EditFeedbackMessage",
		)
	}
	return nil
}

// sendMessage отправляет простое текстовое сообщение
func (mh *MenuHandler) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := mh.bot.Send(msg)
	return err
}

// ProfileHandler интерфейс для работы с профилем
type ProfileHandler interface {
	HandleProfileShow(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleProfileResetAsk(callback *tgbotapi.CallbackQuery, user *models.User) error
}
