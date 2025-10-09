package handlers

import (
	"fmt"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для работы с профилем.

// ProfileHandlerImpl обрабатывает все операции с профилем пользователя.
type ProfileHandlerImpl struct {
	base *BaseHandler
}

// NewProfileHandler создает новый экземпляр ProfileHandler.
func NewProfileHandler(base *BaseHandler) *ProfileHandlerImpl {
	return &ProfileHandlerImpl{
		base: base,
	}
}

// HandleProfileCommand обрабатывает команду /profile.
func (ph *ProfileHandlerImpl) HandleProfileCommand(message *tgbotapi.Message, user *models.User) error {
	summary, err := ph.base.service.BuildProfileSummary(user)
	if err != nil {
		// Используем MessageFactory для отправки сообщения об ошибке
		return ph.base.messageFactory.SendText(message.Chat.ID, ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
	}

	text := summary + "\n\n" + ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
	keyboard := ph.base.keyboardBuilder.CreateProfileMenuKeyboard(user.InterfaceLanguageCode)

	// Используем MessageFactory для отправки сообщения с клавиатурой
	return ph.base.messageFactory.SendWithKeyboard(message.Chat.ID, text, keyboard)
}

// HandleProfileShow показывает профиль пользователя.
func (ph *ProfileHandlerImpl) HandleProfileShow(callback *tgbotapi.CallbackQuery, user *models.User) error {
	summary, err := ph.base.service.BuildProfileSummary(user)
	if err != nil {
		return err
	}

	text := summary + "\n\n" + ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
	keyboard := ph.base.keyboardBuilder.CreateProfileMenuKeyboard(user.InterfaceLanguageCode)
	err = ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// HandleProfileResetAsk запрашивает подтверждение сброса профиля.
func (ph *ProfileHandlerImpl) HandleProfileResetAsk(callback *tgbotapi.CallbackQuery, user *models.User) error {
	title := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_title")
	warn := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_warning")
	text := fmt.Sprintf("%s\n\n%s", title, warn)
	keyboard := ph.base.keyboardBuilder.CreateResetConfirmKeyboard(user.InterfaceLanguageCode)
	err := ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// HandleProfileResetYes выполняет сброс профиля.
func (ph *ProfileHandlerImpl) HandleProfileResetYes(callback *tgbotapi.CallbackQuery, user *models.User) error {
	err := ph.base.service.DB.ResetUserProfile(user.ID)
	if err != nil {
		return err
	}
	// Обновляем в памяти базовые поля
	user.NativeLanguageCode = ""
	user.TargetLanguageCode = ""
	user.State = models.StateWaitingLanguage
	user.Status = models.StatusFilling
	user.ProfileCompletionLevel = 0

	done := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_done")
	// Предложим сразу начать с выбора родного языка
	next := ph.base.service.GetLanguagePrompt(user, "native")
	text := done + "\n\n" + next

	keyboard := ph.base.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)
	err = ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// StartProfileSetup начинает настройку профиля с выбора родного языка.
func (ph *ProfileHandlerImpl) StartProfileSetup(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	keyboard := ph.base.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)

	// Редактируем существующее сообщение вместо создания нового
	err := ph.base.messageFactory.EditWithKeyboard(
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
	ph.base.service.LoggingService.Telegram().InfoWithContext(
		"HandleInterestsContinue called",
		generateRequestID("HandleInterestsContinue"),
		user.TelegramID,
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "telegramID": user.TelegramID},
	)

	// Проверяем, выбраны ли интересы
	selectedInterests, err := ph.base.service.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		ph.base.service.LoggingService.Database().ErrorWithContext(
			"Failed to get selected interests",
			generateRequestID("HandleInterestsContinue"),
			user.TelegramID,
			callback.Message.Chat.ID,
			"HandleInterestsContinue",
			map[string]interface{}{"error": err.Error()},
		)

		return err
	}

	ph.base.service.LoggingService.Telegram().InfoWithContext(
		"User has selected interests",
		generateRequestID("HandleInterestsContinue"),
		user.TelegramID,
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "interestCount": len(selectedInterests), "interests": selectedInterests},
	)

	// Если не выбрано ни одного интереса, сообщаем пользователю и оставляем клавиатуру
	if len(selectedInterests) == 0 {
		warningMsg := "❗ " + ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "choose_at_least_one_interest")
		if warningMsg == "choose_at_least_one_interest" { // fallback if key doesn't exist
			warningMsg = "⚠️ Пожалуйста, выберите хотя бы один интерес"
		}

		// Добавляем оригинальный текст с предупреждением
		chooseInterestsText := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")
		fullText := warningMsg + "\n\n" + chooseInterestsText

		// Получаем интересы и оставляем клавиатуру с интересами видимой, обновляя только текст
		interests, _ := ph.base.service.GetCachedInterests(user.InterfaceLanguageCode)
		keyboard := ph.base.keyboardBuilder.CreateInterestsKeyboard(interests, []int{}, user.InterfaceLanguageCode)
		err = ph.base.messageFactory.EditWithKeyboard(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			fullText,
			&keyboard,
		)

		return err
	}

	// Если интересы выбраны, завершаем профиль
	completedMsg := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_completed")
	keyboard := ph.base.keyboardBuilder.CreateProfileCompletedKeyboard(user.InterfaceLanguageCode)

	err = ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		completedMsg,
		&keyboard,
	)
	if err != nil {
		return err
	}

	// Обновляем статус пользователя
	ph.base.service.LoggingService.Database().InfoWithContext(
		"Updating user state to active",
		generateRequestID("HandleInterestsContinue"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "state": "active"},
	)

	err = ph.base.service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		ph.base.service.LoggingService.Database().ErrorWithContext(
			"Error updating user state",
			generateRequestID("HandleInterestsContinue"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleInterestsContinue",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)

		return err
	}

	ph.base.service.LoggingService.Database().InfoWithContext(
		"Updating user status to active",
		generateRequestID("HandleInterestsContinue"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "status": "active"},
	)

	err = ph.base.service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		ph.base.service.LoggingService.Database().ErrorWithContext(
			"Error updating user status",
			generateRequestID("HandleInterestsContinue"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleInterestsContinue",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)

		return err
	}

	// Увеличиваем уровень завершения профиля до 100%
	ph.base.service.LoggingService.Database().InfoWithContext(
		"Updating user profile completion level to 100%",
		generateRequestID("HandleInterestsContinue"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "completionLevel": localization.ProfileCompletionLevelComplete},
	)

	err = ph.updateProfileCompletionLevel(user.ID, localization.ProfileCompletionLevelComplete)
	if err != nil {
		ph.base.service.LoggingService.Database().ErrorWithContext(
			"Error updating profile completion level",
			generateRequestID("HandleInterestsContinue"),
			int64(user.ID),
			callback.Message.Chat.ID,
			"HandleInterestsContinue",
			map[string]interface{}{"userID": user.ID, "error": err.Error()},
		)

		return err
	}

	ph.base.service.LoggingService.Telegram().InfoWithContext(
		"Successfully completed profile",
		generateRequestID("HandleInterestsContinue"),
		int64(user.ID),
		callback.Message.Chat.ID,
		"HandleInterestsContinue",
		map[string]interface{}{"userID": user.ID, "result": "profile_completed"},
	)

	return nil
}

// updateProfileCompletionLevel обновляет уровень завершения профиля от 0 до 100.
func (ph *ProfileHandlerImpl) updateProfileCompletionLevel(userID int, completionLevel int) error {
	ph.base.service.LoggingService.Database().InfoWithContext(
		"Executing updateProfileCompletionLevel",
		generateRequestID("updateProfileCompletionLevel"),
		int64(userID),
		0, // нет chatID в этой функции
		"updateProfileCompletionLevel",
		map[string]interface{}{"userID": userID, "completionLevel": completionLevel},
	)

	result, err := ph.base.service.DB.GetConnection().Exec(`
		UPDATE users
		SET profile_completion_level = $1, updated_at = NOW()
		WHERE id = $2
	`, completionLevel, userID)
	if err != nil {
		ph.base.service.LoggingService.Database().ErrorWithContext(
			"Error in updateProfileCompletionLevel",
			generateRequestID("updateProfileCompletionLevel"),
			int64(userID),
			0, // нет chatID в этой функции
			"updateProfileCompletionLevel",
			map[string]interface{}{"userID": userID, "completionLevel": completionLevel, "error": err.Error()},
		)

		return err
	}

	rowsAffected, _ := result.RowsAffected()
	ph.base.service.LoggingService.Database().InfoWithContext(
		"updateProfileCompletionLevel completed",
		generateRequestID("updateProfileCompletionLevel"),
		int64(userID),
		0, // нет chatID в этой функции
		"updateProfileCompletionLevel",
		map[string]interface{}{"userID": userID, "rowsAffected": rowsAffected},
	)

	return nil
}

// HandleEditLanguages позволяет редактировать языки пользователя.
func (ph *ProfileHandlerImpl) HandleEditLanguages(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Показываем текущие настройки языков с кнопками редактирования
	text := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_edit_languages") +
		"\n\n" + ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	keyboard := ph.base.keyboardBuilder.CreateEditLanguagesKeyboard(
		user.InterfaceLanguageCode,
		user.NativeLanguageCode,
		user.TargetLanguageCode,
		user.TargetLanguageLevel,
	)

	err := ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// HandleEditNativeLang редактирует родной язык пользователя.
func (ph *ProfileHandlerImpl) HandleEditNativeLang(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	// Показываем клавиатуру с сохранением/отменой вместо обычного выбора
	keyboard := ph.base.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "edit_native", "", false)

	// Добавляем кнопки сохранить/отменить
	saveRow := ph.base.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	err := ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// HandleEditTargetLang редактирует изучаемый язык пользователя (только если родной - русский).
func (ph *ProfileHandlerImpl) HandleEditTargetLang(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Проверяем, что родной язык русский - только в этом случае можно редактировать изучаемый язык
	if user.NativeLanguageCode != "ru" {
		// Не должно происходить по логике, но на всякий случай
		// Используем MessageFactory для отправки сообщения об ошибке
		return ph.base.messageFactory.SendText(callback.Message.Chat.ID, "Редактирование изучаемого языка недоступно при вашем родном языке.")
	}

	text := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")
	// Исключаем родной язык из списка изучаемых
	keyboard := ph.base.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "edit_target", user.NativeLanguageCode, false)

	// Добавляем кнопки сохранить/отменить
	saveRow := ph.base.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	err := ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// HandleEditNativeLanguage сохраняет выбор родного языка с учетом первоначальной логики.
func (ph *ProfileHandlerImpl) HandleEditNativeLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_edit_native_"):]

	// Сохраняем новый родной язык
	err := ph.base.service.DB.UpdateUserNativeLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	user.NativeLanguageCode = langCode

	// Применяем изначальную логику выбора языков
	if langCode == "ru" {
		// Если выбран русский как родной, предлагаем выбрать изучаемый из оставшихся 3
		// Но не меняем существующий изучаемый язык, если он есть
		text := "Выберите изучаемый язык:"
		keyboard := ph.base.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "edit_target", "ru", false)

		// Добавляем кнопки сохранить/отменить
		saveRow := ph.base.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

		err = ph.base.messageFactory.EditWithKeyboard(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			&keyboard,
		)

		return err
	} else {
		// Если выбран не русский, автоматически устанавливаем русский как изучаемый
		err := ph.base.service.DB.UpdateUserTargetLanguage(user.ID, "ru")
		if err != nil {
			return err
		}

		user.TargetLanguageCode = "ru"

		// Показываем подтверждение и предлагаем выбрать уровень
		nativeLangName := ph.base.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)
		text := fmt.Sprintf("Родной язык: %s\nИзучаемый язык: Русский\n\nВыберите уровень владения русским языком:",
			nativeLangName)

		keyboard := ph.base.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, "ru", "edit_level", false)

		// Добавляем кнопки сохранить/отменить
		saveRow := ph.base.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

		err = ph.base.messageFactory.EditWithKeyboard(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			&keyboard,
		)

		return err
	}
}

// HandleEditTargetLanguage сохраняет выбор изучаемого языка.
func (ph *ProfileHandlerImpl) HandleEditTargetLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_edit_target_"):]

	err := ph.base.service.DB.UpdateUserTargetLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	user.TargetLanguageCode = langCode
	langName := ph.base.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)

	// Предлагаем выбрать уровень владения языком
	title := ph.base.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := ph.base.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, langCode, "edit_level_", false)
	// Добавляем save/cancel
	saveRow := ph.base.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	err = ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		&keyboard,
	)

	return err
}

// HandleEditLevelSelection обрабатывает выбор уровня владения языком при редактировании.
func (ph *ProfileHandlerImpl) HandleEditLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error {
	// Сохраняем уровень владения языком
	err := ph.base.service.DB.UpdateUserTargetLanguageLevel(user.ID, levelCode)
	if err != nil {
		return err
	}

	user.TargetLanguageLevel = levelCode

	// Переходим к меню редактирования языков с обновленными данными
	text := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_edit_languages") +
		"\n\n" + ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	keyboard := ph.base.keyboardBuilder.CreateEditLanguagesKeyboard(
		user.InterfaceLanguageCode,
		user.NativeLanguageCode,
		user.TargetLanguageCode,
		user.TargetLanguageLevel,
	)

	err = ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)

	return err
}

// HandleEditLevelLang редактирует уровень владения языком.
func (ph *ProfileHandlerImpl) HandleEditLevelLang(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langName := ph.base.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
	title := ph.base.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := ph.base.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, user.TargetLanguageCode, "edit_level_", false)
	// Добавляем save/cancel
	saveRow := ph.base.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	err := ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		&keyboard,
	)

	return err
}

// ShowProfileSetupFeatures показывает новые возможности заполнения профиля.
func (ph *ProfileHandlerImpl) ShowProfileSetupFeatures(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Создаем сообщение с новыми возможностями
	featuresText := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_features")
	isolatedText := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_isolated_editing")
	detailedText := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_detailed_changes")
	safeText := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_safe_editing")
	massText := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_mass_operations")
	undoText := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_undo")
	navText := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_enhanced_navigation")
	realtimeText := ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "profile_setup_real_time_updates")

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
	continueButton := tgbotapi.NewInlineKeyboardButtonData(
		ph.base.service.Localizer.Get(user.InterfaceLanguageCode, "continue_button"),
		"profile_setup_continue",
	)
	keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{continueButton})

	// Обновляем сообщение
	err := ph.base.messageFactory.EditWithKeyboard(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		fullText,
		&keyboard,
	)

	return err
}
