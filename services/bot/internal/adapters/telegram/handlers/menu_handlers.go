package handlers

import (
	"fmt"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MenuHandler обрабатывает команды и действия главного меню.
type MenuHandler struct {
	base *BaseHandler
}

// NewMenuHandler создает новый экземпляр MenuHandler.
func NewMenuHandler(base *BaseHandler) *MenuHandler {
	return &MenuHandler{
		base: base,
	}
}

// HandleStartCommand обрабатывает команду /start.
func (mh *MenuHandler) HandleStartCommand(message *tgbotapi.Message, user *models.User) error {
	// Всегда показываем главное меню, независимо от состояния профиля
	welcomeText := mh.base.service.GetWelcomeMessage(user)
	menuText := welcomeText + "\n\n" + mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleMainMenuTitle)

	hasProfile := user.ProfileCompletionLevel > 0
	keyboard := mh.base.keyboardBuilder.CreateMainMenuKeyboard(user.InterfaceLanguageCode, hasProfile)

	// Используем MessageFactory для отправки сообщения
	return mh.base.messageFactory.SendWithKeyboard(message.Chat.ID, menuText, keyboard)
}

// HandleStatusCommand обрабатывает команду /status.
func (mh *MenuHandler) HandleStatusCommand(message *tgbotapi.Message, user *models.User) error {
	statusText := fmt.Sprintf(
		"📊 %s:\n\n"+
			"🆔 ID: %d\n"+
			"📝 %s: %s\n"+
			"🔄 %s: %s\n"+
			"📈 %s: %d%%\n"+
			"🌐 %s: %s",
		mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleYourStatus),
		user.ID,
		mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleStatus),
		user.Status,
		mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleState),
		user.State,
		mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleProfileCompletion),
		user.ProfileCompletionLevel,
		mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleInterfaceLanguage),
		user.InterfaceLanguageCode,
	)

	return mh.base.messageFactory.SendText(message.Chat.ID, statusText)
}

// HandleResetCommand обрабатывает команду /reset.
func (mh *MenuHandler) HandleResetCommand(message *tgbotapi.Message, user *models.User) error {
	return mh.base.messageFactory.SendText(message.Chat.ID, mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleProfileReset))
}

// HandleLanguageCommand обрабатывает команду /language.
func (mh *MenuHandler) HandleLanguageCommand(message *tgbotapi.Message, user *models.User) error {
	text := mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleChooseInterfaceLanguage)
	keyboard := mh.base.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "interface", "", true)

	// Используем MessageFactory для отправки сообщения
	return mh.base.messageFactory.SendWithKeyboard(message.Chat.ID, text, keyboard)
}

// HandleBackToMainMenu возвращает пользователя в главное меню.
func (mh *MenuHandler) HandleBackToMainMenu(callback *tgbotapi.CallbackQuery, user *models.User) error {
	welcomeText := mh.base.service.GetWelcomeMessage(user)
	menuText := welcomeText + "\n\n" + mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleMainMenuTitle)

	hasProfile := user.ProfileCompletionLevel > 0
	keyboard := mh.base.keyboardBuilder.CreateMainMenuKeyboard(user.InterfaceLanguageCode, hasProfile)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		menuText,
		keyboard,
	)
	_, err := mh.base.bot.Request(editMsg)

	return err
}

// HandleMainChangeLanguage обрабатывает смену языка интерфейса.
func (mh *MenuHandler) HandleMainChangeLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleChooseInterfaceLanguage)
	keyboard := mh.base.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "interface", "", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := mh.base.bot.Request(editMsg)

	return err
}

// HandleMainViewProfile обрабатывает просмотр профиля.
func (mh *MenuHandler) HandleMainViewProfile(callback *tgbotapi.CallbackQuery, user *models.User, profileHandler *ProfileHandlerImpl) error {
	// Проверяем, заполнен ли профиль по уровню завершения профиля
	if user.ProfileCompletionLevel == 0 {
		// Профиль не заполнен - показываем информационное сообщение и кнопку настройки
		text := mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleEmptyProfileMessage)

		// Создаем клавиатуру с кнопками настройки профиля
		setupButton := tgbotapi.NewInlineKeyboardButtonData(
			mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleSetupProfileButton),
			"show_profile_setup_features",
		)

		keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{setupButton})

		// Редактируем существующее сообщение вместо создания нового
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			keyboard,
		)
		_, err := mh.base.bot.Request(editMsg)

		return err
	}

	// Профиль заполнен - показываем его
	return profileHandler.HandleProfileShow(callback, user)
}

// HandleMainEditProfile обрабатывает редактирование профиля.
func (mh *MenuHandler) HandleMainEditProfile(callback *tgbotapi.CallbackQuery, user *models.User, profileHandler *ProfileHandlerImpl) error {
	return profileHandler.HandleProfileResetAsk(callback, user)
}

// HandleMainFeedback обрабатывает переход к отзывам.
func (mh *MenuHandler) HandleMainFeedback(callback *tgbotapi.CallbackQuery, user *models.User, feedbackHandler FeedbackHandler) error {
	// Переводим пользователя в состояние ожидания отзыва
	_ = mh.base.service.DB.UpdateUserState(user.ID, models.StateWaitingFeedback)

	// Получаем текст обратной связи
	text := mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleFeedbackText)

	// Создаем клавиатуру для обратной связи
	keyboard := mh.createFeedbackKeyboard(user.InterfaceLanguageCode)

	// Редактируем сообщение вместо отправки нового
	return mh.editMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, &keyboard)
}

// HandleFeedbackHelp обрабатывает помощь по обратной связи.
func (mh *MenuHandler) HandleFeedbackHelp(callback *tgbotapi.CallbackQuery, user *models.User) error {
	helpTitle := mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleFeedbackHelpTitle)
	helpContent := mh.base.service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleFeedbackHelpContent)
	helpText := helpTitle + "\n\n" + helpContent

	keyboard := mh.createFeedbackKeyboard(user.InterfaceLanguageCode)

	// Редактируем сообщение с помощью
	return mh.editMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, helpText, &keyboard)
}

// createFeedbackKeyboard создает клавиатуру для обратной связи.
func (mh *MenuHandler) createFeedbackKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	keyboard := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData(mh.base.service.Localizer.Get(lang, localization.LocaleFeedbackBackToMain), "back_to_main_menu"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData(mh.base.service.Localizer.Get(lang, localization.LocaleFeedbackHelp), "feedback_help"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// editMessageTextAndMarkup редактирует сообщение с клавиатурой.
func (mh *MenuHandler) editMessageTextAndMarkup(chatID int64, messageID int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	// Используем MessageFactory для редактирования сообщения
	if keyboard != nil {
		return mh.base.messageFactory.EditWithKeyboard(chatID, messageID, text, keyboard)
	}
	return mh.base.messageFactory.EditText(chatID, messageID, text)
}

// ProfileHandler интерфейс для работы с профилем.
type ProfileHandler interface {
	HandleProfileShow(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleProfileResetAsk(callback *tgbotapi.CallbackQuery, user *models.User) error
}
