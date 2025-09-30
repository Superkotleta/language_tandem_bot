package handlers

import (
	"fmt"
	"log"
	"strconv"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ProfileHandlerImpl обрабатывает все операции с профилем пользователя
type ProfileHandlerImpl struct {
	bot               *tgbotapi.BotAPI
	service           *core.BotService
	keyboardBuilder   *KeyboardBuilder
	editInterestsTemp map[int64][]int // Временное хранение выбранных интересов для каждого пользователя
	errorHandler      *errors.ErrorHandler
}

// NewProfileHandler создает новый экземпляр ProfileHandler
func NewProfileHandler(bot *tgbotapi.BotAPI, service *core.BotService, keyboardBuilder *KeyboardBuilder, errorHandler *errors.ErrorHandler) *ProfileHandlerImpl {
	return &ProfileHandlerImpl{
		bot:               bot,
		service:           service,
		keyboardBuilder:   keyboardBuilder,
		editInterestsTemp: make(map[int64][]int),
		errorHandler:      errorHandler,
	}
}

// HandleProfileCommand обрабатывает команду /profile
func (ph *ProfileHandlerImpl) HandleProfileCommand(message *tgbotapi.Message, user *models.User) error {
	summary, err := ph.service.BuildProfileSummary(user)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, ph.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
		_, err := ph.bot.Request(msg)
		return err
	}
	text := summary + "\n\n" + ph.service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = ph.keyboardBuilder.CreateProfileMenuKeyboard(user.InterfaceLanguageCode)
	_, err = ph.bot.Send(msg)
	return err
}

// HandleProfileShow показывает профиль пользователя
func (ph *ProfileHandlerImpl) HandleProfileShow(callback *tgbotapi.CallbackQuery, user *models.User) error {
	summary, err := ph.service.BuildProfileSummary(user)
	if err != nil {
		return err
	}
	text := summary + "\n\n" + ph.service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		ph.keyboardBuilder.CreateProfileMenuKeyboard(user.InterfaceLanguageCode),
	)
	_, err = ph.bot.Request(edit)
	return err
}

// HandleProfileResetAsk запрашивает подтверждение сброса профиля
func (ph *ProfileHandlerImpl) HandleProfileResetAsk(callback *tgbotapi.CallbackQuery, user *models.User) error {
	title := ph.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_title")
	warn := ph.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_warning")
	text := fmt.Sprintf("%s\n\n%s", title, warn)
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		ph.keyboardBuilder.CreateResetConfirmKeyboard(user.InterfaceLanguageCode),
	)
	_, err := ph.bot.Request(edit)
	return err
}

// HandleProfileResetYes выполняет сброс профиля
func (ph *ProfileHandlerImpl) HandleProfileResetYes(callback *tgbotapi.CallbackQuery, user *models.User) error {
	if err := ph.service.DB.ResetUserProfile(user.ID); err != nil {
		return err
	}
	// Обновляем в памяти базовые поля
	user.NativeLanguageCode = ""
	user.TargetLanguageCode = ""
	user.State = models.StateWaitingLanguage
	user.Status = models.StatusFilling
	user.ProfileCompletionLevel = 0

	done := ph.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_done")
	// Предложим сразу начать с выбора родного языка
	next := ph.service.GetLanguagePrompt(user, "native")
	text := done + "\n\n" + next

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		ph.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true),
	)
	_, err := ph.bot.Request(edit)
	return err
}

// StartProfileSetup начинает настройку профиля с выбора родного языка
func (ph *ProfileHandlerImpl) StartProfileSetup(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ph.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	keyboard := ph.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)

	// Редактируем существующее сообщение вместо создания нового
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := ph.bot.Request(editMsg)
	return err
}

// HandleInterestsContinue обрабатывает продолжение после выбора интересов
func (ph *ProfileHandlerImpl) HandleInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error {
	log.Printf("HandleInterestsContinue called for user ID: %d, Telegram ID: %d", user.ID, user.TelegramID)

	// Проверяем, выбраны ли интересы
	selectedInterests, err := ph.service.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		log.Printf("Error getting selected interests: %v", err)
		return err
	}

	log.Printf("User %d has %d selected interests: %v", user.ID, len(selectedInterests), selectedInterests)

	// Если не выбрано ни одного интереса, сообщаем пользователю и оставляем клавиатуру
	if len(selectedInterests) == 0 {
		warningMsg := "❗ " + ph.service.Localizer.Get(user.InterfaceLanguageCode, "choose_at_least_one_interest")
		if warningMsg == "choose_at_least_one_interest" { // fallback if key doesn't exist
			warningMsg = "⚠️ Пожалуйста, выберите хотя бы один интерес"
		}

		// Добавляем оригинальный текст с предупреждением
		chooseInterestsText := ph.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")
		fullText := warningMsg + "\n\n" + chooseInterestsText

		// Получаем интересы и оставляем клавиатуру с интересами видимой, обновляя только текст
		interests, _ := ph.service.GetCachedInterests(user.InterfaceLanguageCode)
		keyboard := ph.keyboardBuilder.CreateInterestsKeyboard(interests, []int{}, user.InterfaceLanguageCode)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			fullText,
			keyboard,
		)
		_, err := ph.bot.Request(editMsg)
		return err
	}

	// Если интересы выбраны, завершаем профиль
	completedMsg := ph.service.Localizer.Get(user.InterfaceLanguageCode, "profile_completed")
	keyboard := ph.keyboardBuilder.CreateProfileCompletedKeyboard(user.InterfaceLanguageCode)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		completedMsg,
		keyboard,
	)
	_, err = ph.bot.Request(editMsg)
	if err != nil {
		return err
	}

	// Обновляем статус пользователя
	log.Printf("Updating user %d state to active", user.ID)
	err = ph.service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		log.Printf("Error updating user state: %v", err)
		return err
	}

	log.Printf("Updating user %d status to active", user.ID)
	err = ph.service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		log.Printf("Error updating user status: %v", err)
		return err
	}

	// Увеличиваем уровень завершения профиля до 100%
	log.Printf("Updating user %d profile completion level to 100%%", user.ID)
	err = ph.updateProfileCompletionLevel(user.ID, 100)
	if err != nil {
		log.Printf("Error updating profile completion level: %v", err)
		return err
	}

	log.Printf("Successfully completed profile for user %d", user.ID)

	return nil
}

// updateProfileCompletionLevel обновляет уровень завершения профиля от 0 до 100
func (ph *ProfileHandlerImpl) updateProfileCompletionLevel(userID int, completionLevel int) error {
	log.Printf("Executing updateProfileCompletionLevel: userID=%d, level=%d", userID, completionLevel)

	result, err := ph.service.DB.GetConnection().Exec(`
		UPDATE users
		SET profile_completion_level = $1, updated_at = NOW()
		WHERE id = $2
	`, completionLevel, userID)
	if err != nil {
		log.Printf("Error in updateProfileCompletionLevel: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("updateProfileCompletionLevel: %d rows affected for user %d", rowsAffected, userID)

	return nil
}

func (ph *ProfileHandlerImpl) HandleEditInterests(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все доступные интересы через кэш
	interests, err := ph.service.GetCachedInterests(user.InterfaceLanguageCode)
	if err != nil {
		return err
	}

	// Получаем текущие выбранные интересы пользователя через Batch Loading
	selectedInterestsMap, err := ph.service.BatchLoadUserInterests([]int{user.ID})
	var selectedInterests []int
	if err != nil {
		log.Printf("Error loading user interests: %v", err)
		selectedInterests = []int{} // fallback
	} else {
		selectedInterests = selectedInterestsMap[user.ID]
	}

	// Инициализируем временное хранилище для сессии редактирования
	userID := int64(user.ID)
	ph.editInterestsTemp[userID] = make([]int, len(selectedInterests))
	copy(ph.editInterestsTemp[userID], selectedInterests)

	keyboard := ph.keyboardBuilder.CreateEditInterestsKeyboard(interests, selectedInterests, user.InterfaceLanguageCode)
	text := ph.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests") +
		"\n\n" + ph.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + ph.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err = ph.bot.Request(editMsg)
	return err
}

// HandleEditLanguages позволяет редактировать языки пользователя
func (ph *ProfileHandlerImpl) HandleEditLanguages(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Показываем текущие настройки языков с кнопками редактирования
	text := ph.service.Localizer.Get(user.InterfaceLanguageCode, "profile_edit_languages") +
		"\n\n" + ph.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + ph.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	keyboard := ph.keyboardBuilder.CreateEditLanguagesKeyboard(
		user.InterfaceLanguageCode,
		user.NativeLanguageCode,
		user.TargetLanguageCode,
		user.TargetLanguageLevel,
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := ph.bot.Request(editMsg)
	return err
}

// HandleSaveEdits сохраняет изменения и возвращается к просмотру профиля
func (ph *ProfileHandlerImpl) HandleSaveEdits(callback *tgbotapi.CallbackQuery, user *models.User) error {
	userID := int64(user.ID)

	// Если есть временное хранилище интересов, применяем изменения в БД
	if tempInterests, exists := ph.editInterestsTemp[userID]; exists {
		// Очищаем текущие интересы в БД
		err := ph.service.DB.ClearUserInterests(user.ID)
		if err != nil {
			log.Printf("Error clearing user interests: %v", err)
			return err
		}

		// Сохраняем новые интересы
		for _, interestID := range tempInterests {
			err := ph.service.DB.SaveUserInterest(user.ID, interestID, false)
			if err != nil {
				log.Printf("Error saving user interest %d: %v", interestID, err)
				return err
			}
		}

		// Удаляем временное хранилище
		delete(ph.editInterestsTemp, userID)
	}

	// Показываем профиль с обновленными данными
	return ph.HandleProfileShow(callback, user)
}

// HandleCancelEdits отменяет изменения и возвращается к просмотру профиля
func (ph *ProfileHandlerImpl) HandleCancelEdits(callback *tgbotapi.CallbackQuery, user *models.User) error {
	userID := int64(user.ID)
	// Удаляем временное хранилище без сохранения
	delete(ph.editInterestsTemp, userID)
	return ph.HandleProfileShow(callback, user)
}

// HandleEditNativeLang редактирует родной язык пользователя
func (ph *ProfileHandlerImpl) HandleEditNativeLang(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := ph.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	// Показываем клавиатуру с сохранением/отменой вместо обычного выбора
	keyboard := ph.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "edit_native", "", false)

	// Добавляем кнопки сохранить/отменить
	saveRow := ph.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := ph.bot.Request(editMsg)
	return err
}

// HandleEditTargetLang редактирует изучаемый язык пользователя (только если родной - русский)
func (ph *ProfileHandlerImpl) HandleEditTargetLang(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Проверяем, что родной язык русский - только в этом случае можно редактировать изучаемый язык
	if user.NativeLanguageCode != "ru" {
		// Не должно происходить по логике, но на всякий случай
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Редактирование изучаемого языка недоступно при вашем родном языке.")
		_, err := ph.bot.Request(msg)
		return err
	}

	text := ph.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")
	// Исключаем родной язык из списка изучаемых
	keyboard := ph.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "edit_target", user.NativeLanguageCode, false)

	// Добавляем кнопки сохранить/отменить
	saveRow := ph.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := ph.bot.Request(editMsg)
	return err
}

// HandleEditNativeLanguage сохраняет выбор родного языка с учетом первоначальной логики
func (ph *ProfileHandlerImpl) HandleEditNativeLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_edit_native_"):]

	// Сохраняем новый родной язык
	err := ph.service.DB.UpdateUserNativeLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	user.NativeLanguageCode = langCode

	// Применяем изначальную логику выбора языков
	if langCode == "ru" {
		// Если выбран русский как родной, предлагаем выбрать изучаемый из оставшихся 3
		// Но не меняем существующий изучаемый язык, если он есть
		text := "Выберите изучаемый язык:"
		keyboard := ph.keyboardBuilder.CreateLanguageKeyboard(user.InterfaceLanguageCode, "edit_target", "ru", false)

		// Добавляем кнопки сохранить/отменить
		saveRow := ph.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			keyboard,
		)
		_, err := ph.bot.Request(editMsg)
		return err
	} else {
		// Если выбран не русский, автоматически устанавливаем русский как изучаемый
		err := ph.service.DB.UpdateUserTargetLanguage(user.ID, "ru")
		if err != nil {
			return err
		}
		user.TargetLanguageCode = "ru"

		// Показываем подтверждение и предлагаем выбрать уровень
		nativeLangName := ph.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)
		text := fmt.Sprintf("Родной язык: %s\nИзучаемый язык: Русский\n\nВыберите уровень владения русским языком:",
			nativeLangName)

		keyboard := ph.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, "ru", "edit_level", false)

		// Добавляем кнопки сохранить/отменить
		saveRow := ph.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			keyboard,
		)
		_, err = ph.bot.Request(editMsg)
		return err
	}
}

// HandleEditTargetLanguage сохраняет выбор изучаемого языка
func (ph *ProfileHandlerImpl) HandleEditTargetLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_edit_target_"):]

	err := ph.service.DB.UpdateUserTargetLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	user.TargetLanguageCode = langCode
	langName := ph.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)

	// Предлагаем выбрать уровень владения языком
	title := ph.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := ph.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, langCode, "edit_level_", false)
	// Добавляем save/cancel
	saveRow := ph.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err = ph.bot.Request(editMsg)
	return err
}

// HandleEditInterestSelection обрабатывает выбор/отмену интереса при редактировании
func (ph *ProfileHandlerImpl) HandleEditInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		log.Printf("Error parsing interest ID: %v", err)
		return err
	}

	userID := int64(user.ID)

	// Если временного хранилища нет, инициализируем его через Batch Loading
	if _, exists := ph.editInterestsTemp[userID]; !exists {
		selectedInterestsMap, err := ph.service.BatchLoadUserInterests([]int{user.ID})
		var selectedInterests []int
		if err != nil {
			log.Printf("Error getting user interests, using empty list: %v", err)
			selectedInterests = []int{}
		} else {
			selectedInterests = selectedInterestsMap[user.ID]
		}
		ph.editInterestsTemp[userID] = make([]int, len(selectedInterests))
		copy(ph.editInterestsTemp[userID], selectedInterests)
	}

	// Переключаем интерес в временном хранилище (toggle)
	isCurrentlySelected := false
	for i, id := range ph.editInterestsTemp[userID] {
		if id == interestID {
			// Убираем из списка
			ph.editInterestsTemp[userID] = append(ph.editInterestsTemp[userID][:i], ph.editInterestsTemp[userID][i+1:]...)
			isCurrentlySelected = true
			break
		}
	}

	if !isCurrentlySelected {
		// Добавляем в список
		ph.editInterestsTemp[userID] = append(ph.editInterestsTemp[userID], interestID)
	}

	// Обновляем клавиатуру с новым состоянием через кэш
	interests, err := ph.service.GetCachedInterests(user.InterfaceLanguageCode)
	if err != nil {
		return err
	}

	keyboard := ph.keyboardBuilder.CreateEditInterestsKeyboard(interests, ph.editInterestsTemp[userID], user.InterfaceLanguageCode)
	text := ph.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests") +
		"\n\n" + ph.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + ph.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err = ph.bot.Request(editMsg)
	return err
}

// HandleEditLevelSelection обрабатывает выбор уровня владения языком при редактировании
func (ph *ProfileHandlerImpl) HandleEditLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error {
	// Сохраняем уровень владения языком
	err := ph.service.DB.UpdateUserTargetLanguageLevel(user.ID, levelCode)
	if err != nil {
		return err
	}

	user.TargetLanguageLevel = levelCode

	// Переходим к меню редактирования языков с обновленными данными
	text := ph.service.Localizer.Get(user.InterfaceLanguageCode, "profile_edit_languages") +
		"\n\n" + ph.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + ph.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	keyboard := ph.keyboardBuilder.CreateEditLanguagesKeyboard(
		user.InterfaceLanguageCode,
		user.NativeLanguageCode,
		user.TargetLanguageCode,
		user.TargetLanguageLevel,
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err = ph.bot.Request(editMsg)
	return err
}

// HandleEditLevelLang редактирует уровень владения языком
func (ph *ProfileHandlerImpl) HandleEditLevelLang(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langName := ph.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
	title := ph.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := ph.keyboardBuilder.CreateLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, user.TargetLanguageCode, "edit_level_", false)
	// Добавляем save/cancel
	saveRow := ph.keyboardBuilder.CreateSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err := ph.bot.Request(editMsg)
	return err
}
