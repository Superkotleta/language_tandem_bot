package profile

import (
	"fmt"

	"language-exchange-bot/internal/adapters/telegram/handlers/base"
	"language-exchange-bot/internal/adapters/telegram/handlers/language"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для работы с профилем.

// ProfileHandlerImpl обрабатывает все операции с профилем пользователя.
type ProfileHandlerImpl struct {
	base *base.BaseHandler
}

// NewProfileHandler создает новый экземпляр ProfileHandler.
func NewProfileHandler(baseHandler *base.BaseHandler) *ProfileHandlerImpl {
	return &ProfileHandlerImpl{
		base: baseHandler,
	}
}

// HandleProfileCommand обрабатывает команду /profile.
func (ph *ProfileHandlerImpl) HandleProfileCommand(message *tgbotapi.Message, user *models.User) error {
	summary, err := ph.base.Service.BuildProfileSummary(user)
	if err != nil {
		// Используем MessageFactory для отправки сообщения об ошибке
		return ph.base.MessageFactory.SendText(message.Chat.ID, ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
	}

	text := summary + "\n\n" + ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
	keyboard := ph.base.KeyboardBuilder.CreateProfileMenuKeyboard(user.InterfaceLanguageCode)

	// Используем MessageFactory для отправки сообщения с клавиатурой
	return ph.base.MessageFactory.SendWithKeyboard(message.Chat.ID, text, keyboard)
}

// HandleProfileShow показывает профиль пользователя.
func (ph *ProfileHandlerImpl) HandleProfileShow(callback *tgbotapi.CallbackQuery, user *models.User) error {
	summary, err := ph.base.Service.BuildProfileSummary(user)
	if err != nil {
		return err
	}

	text := summary + "\n\n" + ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
	keyboard := ph.base.KeyboardBuilder.CreateProfileMenuKeyboard(user.InterfaceLanguageCode)
	err = ph.base.MessageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// HandleProfileResetAsk запрашивает подтверждение сброса профиля.
func (ph *ProfileHandlerImpl) HandleProfileResetAsk(callback *tgbotapi.CallbackQuery, user *models.User) error {
	title := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_title")
	warn := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_warning")
	text := fmt.Sprintf("%s\n\n%s", title, warn)
	keyboard := ph.base.KeyboardBuilder.CreateResetConfirmKeyboard(user.InterfaceLanguageCode)
	err := ph.base.MessageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// HandleProfileResetYes выполняет сброс профиля.
func (ph *ProfileHandlerImpl) HandleProfileResetYes(callback *tgbotapi.CallbackQuery, user *models.User) error {
	err := ph.base.Service.DB.ResetUserProfile(user.ID)
	if err != nil {
		return err
	}
	// Обновляем в памяти базовые поля
	user.NativeLanguageCode = ""
	user.TargetLanguageCode = ""
	user.State = models.StateWaitingLanguage
	user.Status = models.StatusFilling
	user.ProfileCompletionLevel = 0

	done := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_done")
	// Предложим сразу начать с выбора родного языка
	next := ph.base.Service.GetLanguagePrompt(user, "native")
	text := done + "\n\n" + next

	keyboard := ph.base.KeyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)
	err = ph.base.MessageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// StartProfileSetup начинает настройку профиля с выбора родного языка.
func (ph *ProfileHandlerImpl) StartProfileSetup(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	keyboard := ph.base.KeyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)

	// Редактируем существующее сообщение вместо создания нового
	err := ph.base.MessageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// HandleInterestsContinue обрабатывает продолжение после выбора интересов.
//
//nolint:funlen
func (ph *ProfileHandlerImpl) HandleInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error {
	ph.base.Service.LoggingService.Telegram().InfoWithContext(
		"HandleInterestsContinue called",
		base.GenerateRequestID("HandleInterestsContinue"),
		user.TelegramID,
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "telegramID": user.TelegramID},
	)

	// Проверяем, выбраны ли интересы
	selectedInterests, err := ph.base.Service.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		ph.base.Service.LoggingService.Database().ErrorWithContext(
			"Failed to get selected interests",
			base.GenerateRequestID("HandleInterestsContinue"),
			user.TelegramID,
			callback.Message.Chat.ID,
			"HandleInterestsContinue",
			map[string]interface{}{"error": err.Error()},
		)

		return err
	}

	ph.base.Service.LoggingService.Telegram().InfoWithContext(
		"User has selected interests",
		base.GenerateRequestID("HandleInterestsContinue"),
		user.TelegramID,
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "interestCount": len(selectedInterests), "interests": selectedInterests},
	)

	// Если не выбрано ни одного интереса, сообщаем пользователю и оставляем клавиатуру
	if len(selectedInterests) == 0 {
		warningMsg := "❗ " + ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "choose_at_least_one_interest")
		if warningMsg == "choose_at_least_one_interest" { // fallback if key doesn't exist
			warningMsg = "⚠️ Пожалуйста, выберите хотя бы один интерес"
		}

		// Добавляем оригинальный текст с предупреждением
		chooseInterestsText := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")
		fullText := warningMsg + "\n\n" + chooseInterestsText

		// Получаем интересы и оставляем клавиатуру с интересами видимой, обновляя только текст
		interests, _ := ph.base.Service.GetCachedInterests(user.InterfaceLanguageCode)
		keyboard := ph.base.KeyboardBuilder.CreateInterestsKeyboard(interests, []int{}, user.InterfaceLanguageCode)
		err = ph.base.MessageFactory.EditWithKeyboard(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			fullText,
			&keyboard,
		)

		return err
	}

	// Если интересы выбраны, завершаем профиль
	completedMsg := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_completed")
	keyboard := ph.base.KeyboardBuilder.CreateProfileCompletedKeyboard(user.InterfaceLanguageCode)

	err = ph.base.MessageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		completedMsg,
		&keyboard,
	)
	if err != nil {
		return err
	}

	// Обновляем статус пользователя
	ph.base.Service.LoggingService.Database().InfoWithContext(
		"Updating user state to active",
		base.GenerateRequestID("HandleInterestsContinue"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "state": "active"},
	)

	err = ph.base.Service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		ph.base.Service.LoggingService.Database().ErrorWithContext(
			"Error updating user state",
			base.GenerateRequestID("HandleInterestsContinue"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleInterestsContinue",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)

		return err
	}

	ph.base.Service.LoggingService.Database().InfoWithContext(
		"Updating user status to active",
		base.GenerateRequestID("HandleInterestsContinue"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "status": "active"},
	)

	err = ph.base.Service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		ph.base.Service.LoggingService.Database().ErrorWithContext(
			"Error updating user status",
			base.GenerateRequestID("HandleInterestsContinue"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleInterestsContinue",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)

		return err
	}

	// Увеличиваем уровень завершения профиля до 100%
	ph.base.Service.LoggingService.Database().InfoWithContext(
		"Updating user profile completion level to 100%",
		base.GenerateRequestID("HandleInterestsContinue"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "completionLevel": localization.ProfileCompletionLevelComplete},
	)

	err = ph.updateProfileCompletionLevel(user.ID, localization.ProfileCompletionLevelComplete)
	if err != nil {
		ph.base.Service.LoggingService.Database().ErrorWithContext(
			"Error updating profile completion level",
			base.GenerateRequestID("HandleInterestsContinue"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleInterestsContinue",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)

		return err
	}

	ph.base.Service.LoggingService.Telegram().InfoWithContext(
		"Successfully completed profile",
		base.GenerateRequestID("HandleInterestsContinue"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "result": "profile_completed"},
	)

	return nil
}

// updateProfileCompletionLevel обновляет уровень завершения профиля от 0 до 100.
func (ph *ProfileHandlerImpl) updateProfileCompletionLevel(userID int, completionLevel int) error {
	ph.base.Service.LoggingService.Database().InfoWithContext(
		"Executing updateProfileCompletionLevel",
		base.GenerateRequestID("updateProfileCompletionLevel"),
		int64(userID),
		0, // нет chatID в этой функции
		"updateProfileCompletionLevel",
		map[string]interface{}{"userID": userID, "completionLevel": completionLevel},
	)

	result, err := ph.base.Service.DB.GetConnection().Exec(`
		UPDATE users
		SET profile_completion_level = $1, updated_at = NOW()
		WHERE id = $2
	`, completionLevel, userID)
	if err != nil {
		ph.base.Service.LoggingService.Database().ErrorWithContext(
			"Error in updateProfileCompletionLevel",
			base.GenerateRequestID("updateProfileCompletionLevel"),
			int64(userID),
			0, // нет chatID в этой функции
			"updateProfileCompletionLevel",
			map[string]interface{}{"userID": userID, "completionLevel": completionLevel, "error": err.Error()},
		)

		return err
	}

	rowsAffected, _ := result.RowsAffected()
	ph.base.Service.LoggingService.Database().InfoWithContext(
		"updateProfileCompletionLevel completed",
		base.GenerateRequestID("updateProfileCompletionLevel"),
		int64(userID),
		0, // нет chatID в этой функции
		"updateProfileCompletionLevel",
		map[string]interface{}{"userID": userID, "rowsAffected": rowsAffected},
	)

	return nil
}

// HandleEditLanguages запускает изолированный редактор языков.
// Старая логика заменена на IsolatedLanguageEditor.
func (ph *ProfileHandlerImpl) HandleEditLanguages(callback *tgbotapi.CallbackQuery, user *models.User) error {
	editor := language.NewIsolatedLanguageEditor(ph.base)
	return editor.StartEditSession(callback, user)
}

// ShowProfileSetupFeatures показывает новые возможности заполнения профиля.
func (ph *ProfileHandlerImpl) ShowProfileSetupFeatures(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Создаем сообщение с новыми возможностями
	featuresText := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_features")
	isolatedText := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_isolated_editing")
	detailedText := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_detailed_changes")
	safeText := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_safe_editing")
	massText := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_mass_operations")
	undoText := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_undo")
	navText := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_enhanced_navigation")
	realtimeText := ph.base.Service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_real_time_updates")

	fullText := fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s\n%s\n%s\n%s",
		featuresText,
		isolatedText,
		detailedText,
		safeText,
		massText,
		undoText,
		navText,
		realtimeText,
	)

	// Создаем клавиатуру с кнопкой "Продолжить"
	continueButton := ph.base.KeyboardBuilder.CreateContinueButton(
		user.InterfaceLanguageCode,
		"profile_setup_continue",
	)
	keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{continueButton})

	// Обновляем сообщение
	err := ph.base.MessageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		fullText,
		&keyboard,
	)

	return err
}
