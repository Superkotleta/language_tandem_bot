package menu

import (
	"fmt"
	"log"

	"language-exchange-bot/internal/adapters/telegram/handlers/base"
	"language-exchange-bot/internal/adapters/telegram/handlers/feedback"
	"language-exchange-bot/internal/adapters/telegram/handlers/profile"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MenuHandler обрабатывает команды и действия главного меню.
type MenuHandler struct {
	base *base.BaseHandler
}

// NewMenuHandler создает новый экземпляр MenuHandler.
func NewMenuHandler(baseHandler *base.BaseHandler) *MenuHandler {
	return &MenuHandler{
		base: baseHandler,
	}
}

// HandleStartCommand обрабатывает команду /start.
func (mh *MenuHandler) HandleStartCommand(message *tgbotapi.Message, user *models.User) error {
	// Всегда показываем главное меню, независимо от состояния профиля
	welcomeText := mh.base.Service.GetWelcomeMessage(user)
	menuText := welcomeText + "\n\n" + mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleMainMenuTitle)

	hasProfile := user.ProfileCompletionLevel > 0
	keyboard := mh.base.KeyboardBuilder.CreateMainMenuKeyboard(user.InterfaceLanguageCode, hasProfile)

	// Используем MessageFactory для отправки сообщения
	return mh.base.MessageFactory.SendWithKeyboard(message.Chat.ID, menuText, keyboard)
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
		mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleYourStatus),
		user.ID,
		mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleStatus),
		user.Status,
		mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleState),
		user.State,
		mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleProfileCompletion),
		user.ProfileCompletionLevel,
		mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleInterfaceLanguage),
		user.InterfaceLanguageCode,
	)

	return mh.base.MessageFactory.SendText(message.Chat.ID, statusText)
}

// HandleResetCommand обрабатывает команду /reset.
func (mh *MenuHandler) HandleResetCommand(message *tgbotapi.Message, user *models.User) error {
	return mh.base.MessageFactory.SendText(message.Chat.ID, mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleProfileReset))
}

// HandleLanguageCommand обрабатывает команду /language.
func (mh *MenuHandler) HandleLanguageCommand(message *tgbotapi.Message, user *models.User) error {
	text := mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleChooseInterfaceLanguage)
	keyboard := mh.base.KeyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "interface", "", true)

	// Используем MessageFactory для отправки сообщения
	return mh.base.MessageFactory.SendWithKeyboard(message.Chat.ID, text, keyboard)
}

// HandleBackToMainMenu возвращает пользователя в главное меню.
func (mh *MenuHandler) HandleBackToMainMenu(callback *tgbotapi.CallbackQuery, user *models.User) error {
	welcomeText := mh.base.Service.GetWelcomeMessage(user)
	menuText := welcomeText + "\n\n" + mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleMainMenuTitle)

	hasProfile := user.ProfileCompletionLevel > 0
	keyboard := mh.base.KeyboardBuilder.CreateMainMenuKeyboard(user.InterfaceLanguageCode, hasProfile)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		menuText,
		keyboard,
	)
	_, err := mh.base.Bot.Request(editMsg)

	return err
}

// HandleMainChangeLanguage обрабатывает смену языка интерфейса.
func (mh *MenuHandler) HandleMainChangeLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleChooseInterfaceLanguage)
	keyboard := mh.base.KeyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "interface", "", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := mh.base.Bot.Request(editMsg)

	return err
}

// HandleMainViewProfile обрабатывает просмотр профиля.
func (mh *MenuHandler) HandleMainViewProfile(callback *tgbotapi.CallbackQuery, user *models.User, profileHandler *profile.ProfileHandlerImpl) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	// Получаем свежие данные пользователя для проверки актуального статуса профиля
	freshUser, err := mh.base.Service.GetCachedUser(user.TelegramID)
	if err != nil {
		log.Printf("Failed to get fresh user data for profile view: %v", err)
		// В случае ошибки используем переданные данные
		freshUser = user
	}

	if freshUser == nil {
		return fmt.Errorf("freshUser is nil after GetCachedUser")
	}

	// Проверяем, заполнен ли профиль по уровню завершения профиля
	if freshUser.ProfileCompletionLevel == 0 {
		// Профиль не заполнен - показываем информационное сообщение и кнопку настройки
		text := mh.base.Service.Localizer.Get(freshUser.InterfaceLanguageCode, localization.LocaleEmptyProfileMessage)

		// Создаем клавиатуру с кнопками настройки профиля
		setupButton := tgbotapi.NewInlineKeyboardButtonData(
			mh.base.Service.Localizer.Get(freshUser.InterfaceLanguageCode, localization.LocaleSetupProfileButton),
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
		_, err := mh.base.Bot.Request(editMsg)

		return err
	}

	// Профиль заполнен - показываем его
	return profileHandler.HandleProfileShow(callback, freshUser)
}

// HandleMainEditProfile обрабатывает редактирование профиля.
func (mh *MenuHandler) HandleMainEditProfile(callback *tgbotapi.CallbackQuery, user *models.User, profileHandler *profile.ProfileHandlerImpl) error {
	return profileHandler.HandleProfileResetAsk(callback, user)
}

// HandleMainFeedback обрабатывает переход к отзывам.
func (mh *MenuHandler) HandleMainFeedback(callback *tgbotapi.CallbackQuery, user *models.User, feedbackHandler feedback.FeedbackHandler) error {
	// Переводим пользователя в состояние ожидания отзыва
	if err := mh.base.Service.DB.UpdateUserState(user.ID, models.StateWaitingFeedback); err != nil {
		log.Printf("Failed to update user state to waiting feedback for user %d: %v", user.ID, err)
	}

	// Получаем текст обратной связи
	text := mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleFeedbackText)

	// Создаем клавиатуру для обратной связи
	keyboard := mh.createFeedbackKeyboard(user.InterfaceLanguageCode)

	// Редактируем сообщение вместо отправки нового
	return mh.editMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, &keyboard)
}

// HandleFeedbackHelp обрабатывает помощь по обратной связи.
func (mh *MenuHandler) HandleFeedbackHelp(callback *tgbotapi.CallbackQuery, user *models.User) error {
	helpTitle := mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleFeedbackHelpTitle)
	helpContent := mh.base.Service.Localizer.Get(user.InterfaceLanguageCode, localization.LocaleFeedbackHelpContent)
	helpText := helpTitle + "\n\n" + helpContent

	keyboard := mh.createFeedbackKeyboard(user.InterfaceLanguageCode)

	// Редактируем сообщение с помощью
	return mh.editMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, helpText, &keyboard)
}

// createFeedbackKeyboard создает клавиатуру для обратной связи.
func (mh *MenuHandler) createFeedbackKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	keyboard := [][]tgbotapi.InlineKeyboardButton{
		{
			mh.base.KeyboardBuilder.CreateBackToMainButton(lang),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData(mh.base.Service.Localizer.Get(lang, localization.LocaleFeedbackHelp), "feedback_help"),
		},
	}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// editMessageTextAndMarkup редактирует сообщение с клавиатурой.
func (mh *MenuHandler) editMessageTextAndMarkup(chatID int64, messageID int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	// Используем MessageFactory для редактирования сообщения
	if keyboard != nil {
		return mh.base.MessageFactory.EditWithKeyboard(chatID, messageID, text, keyboard)
	}

	return mh.base.MessageFactory.EditText(chatID, messageID, text)
}

// ProfileHandler интерфейс для работы с профилем.
type ProfileHandler interface {
	HandleProfileShow(callback *tgbotapi.CallbackQuery, user *models.User) error
	HandleProfileResetAsk(callback *tgbotapi.CallbackQuery, user *models.User) error
}
