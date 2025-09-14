package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramHandler struct {
	bot               *tgbotapi.BotAPI
	service           *core.BotService
	editInterestsTemp map[int64][]int // Временное хранение выбранных интересов для каждого пользователя
}

func NewTelegramHandler(bot *tgbotapi.BotAPI, service *core.BotService) *TelegramHandler {
	return &TelegramHandler{
		bot:               bot,
		service:           service,
		editInterestsTemp: make(map[int64][]int),
	}
}

func (h *TelegramHandler) HandleUpdate(update tgbotapi.Update) error {
	if update.Message != nil {
		return h.handleMessage(update.Message)
	}
	if update.CallbackQuery != nil {
		return h.handleCallbackQuery(update.CallbackQuery)
	}
	return nil
}

func (h *TelegramHandler) handleMessage(message *tgbotapi.Message) error {
	user, err := h.service.HandleUserRegistration(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
		message.From.LanguageCode,
	)
	if err != nil {
		log.Printf("Error handling user registration: %v", err)
		return err
	}

	if message.IsCommand() {
		return h.handleCommand(message, user)
	}
	return h.handleState(message, user)
}

func (h *TelegramHandler) handleCommand(message *tgbotapi.Message, user *models.User) error {
	switch message.Command() {
	case "start":
		return h.handleStartCommand(message, user)
	case "status":
		return h.handleStatusCommand(message, user)
	case "reset":
		return h.handleResetCommand(message, user)
	case "language":
		return h.handleLanguageCommand(message, user)
	case "profile":
		return h.handleProfileCommand(message, user)
	case "feedback":
		return h.handleFeedbackCommand(message, user)
	default:
		return h.sendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
	}
}

func (h *TelegramHandler) handleStartCommand(message *tgbotapi.Message, user *models.User) error {
	// Всегда показываем главное меню, независимо от состояния профиля
	welcomeText := h.service.GetWelcomeMessage(user)
	menuText := welcomeText + "\n\n" + h.service.Localizer.Get(user.InterfaceLanguageCode, "main_menu_title")

	msg := tgbotapi.NewMessage(message.Chat.ID, menuText)
	msg.ReplyMarkup = h.createMainMenuKeyboard(user.InterfaceLanguageCode)
	if _, err := h.bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func (h *TelegramHandler) handleStatusCommand(message *tgbotapi.Message, user *models.User) error {
	statusText := fmt.Sprintf(
		"📊 %s:\n\n"+
			"🆔 ID: %d\n"+
			"📝 %s: %s\n"+
			"🔄 %s: %s\n"+
			"📈 %s: %d%%\n"+
			"🌐 %s: %s",
		h.service.Localizer.Get(user.InterfaceLanguageCode, "your_status"),
		user.ID,
		h.service.Localizer.Get(user.InterfaceLanguageCode, "status"),
		user.Status,
		h.service.Localizer.Get(user.InterfaceLanguageCode, "state"),
		user.State,
		h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_completion"),
		user.ProfileCompletionLevel,
		h.service.Localizer.Get(user.InterfaceLanguageCode, "interface_language"),
		user.InterfaceLanguageCode,
	)
	return h.sendMessage(message.Chat.ID, statusText)
}

func (h *TelegramHandler) handleResetCommand(message *tgbotapi.Message, user *models.User) error {
	return h.sendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset"))
}

func (h *TelegramHandler) handleLanguageCommand(message *tgbotapi.Message, user *models.User) error {
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interface_language")
	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "interface", "", true)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.bot.Send(msg)
	return err
}

func (h *TelegramHandler) handleState(message *tgbotapi.Message, user *models.User) error {
	switch user.State {
	case models.StateWaitingLanguage,
		models.StateWaitingInterests,
		models.StateWaitingTime:
		return h.sendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "use_menu_above"))
	default:
		return h.sendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
	}
}

// Поддержка новых колбэков в роутере
func (h *TelegramHandler) handleCallbackQuery(callback *tgbotapi.CallbackQuery) error {
	user, err := h.service.HandleUserRegistration(
		callback.From.ID,
		callback.From.UserName,
		callback.From.FirstName,
		callback.From.LanguageCode,
	)
	if err != nil {
		log.Printf("Error handling user registration: %v", err)
		return err
	}

	data := callback.Data
	_, _ = h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	switch {
	case strings.HasPrefix(data, "lang_native_"):
		return h.handleNativeLanguageCallback(callback, user)
	case strings.HasPrefix(data, "lang_target_"):
		return h.handleTargetLanguageCallback(callback, user)
	case strings.HasPrefix(data, "lang_edit_native_"):
		return h.handleEditNativeLanguage(callback, user)
	case strings.HasPrefix(data, "lang_edit_target_"):
		return h.handleEditTargetLanguage(callback, user)
	case strings.HasPrefix(data, "lang_interface_"):
		langCode := strings.TrimPrefix(data, "lang_interface_")
		return h.handleInterfaceLanguageSelection(callback, user, langCode)
	case strings.HasPrefix(data, "interest_"):
		interestID := strings.TrimPrefix(data, "interest_")
		return h.handleInterestSelection(callback, user, interestID)
	case strings.HasPrefix(data, "edit_interest_"):
		interestID := strings.TrimPrefix(data, "edit_interest_")
		return h.handleEditInterestSelection(callback, user, interestID)
	case data == "profile_show":
		return h.handleProfileShow(callback, user)
	case data == "profile_reset_ask":
		return h.handleProfileResetAsk(callback, user)
	case data == "profile_reset_yes":
		return h.handleProfileResetYes(callback, user)
	case data == "profile_reset_no":
		return h.handleProfileResetNo(callback, user)
	case data == "interests_continue":
		return h.handleInterestsContinue(callback, user)
	case data == "languages_continue_filling":
		return h.handleLanguagesContinueFilling(callback, user)
	case data == "languages_reselect":
		return h.handleLanguagesReselect(callback, user)
	case strings.HasPrefix(data, "level_"):
		levelCode := strings.TrimPrefix(data, "level_")
		return h.handleLanguageLevelSelection(callback, user, levelCode)
	case strings.HasPrefix(data, "edit_level_"):
		levelCode := strings.TrimPrefix(data, "edit_level_")
		return h.handleEditLevelSelection(callback, user, levelCode)
	case data == "back_to_previous_step":
		return h.handleBackToPreviousStep(callback, user)
	case data == "main_change_language":
		return h.handleMainChangeLanguage(callback, user)
	case data == "main_view_profile":
		return h.handleMainViewProfile(callback, user)
	case data == "main_edit_profile":
		return h.handleMainEditProfile(callback, user)
	case data == "main_feedback":
		return h.handleMainFeedback(callback, user)
	case data == "start_profile_setup":
		return h.startProfileSetup(callback, user)
	case data == "back_to_main_menu":
		return h.handleBackToMainMenu(callback, user)
	case data == "edit_interests":
		return h.handleEditInterests(callback, user)
	case data == "edit_languages":
		return h.handleEditLanguages(callback, user)
	case data == "save_edits":
		return h.handleSaveEdits(callback, user)
	case data == "cancel_edits":
		return h.handleCancelEdits(callback, user)
	case data == "edit_native_lang":
		return h.handleEditNativeLang(callback, user)
	case data == "edit_target_lang":
		return h.handleEditTargetLang(callback, user)
	case data == "edit_level":
		return h.handleEditLevelLang(callback, user)
	default:
		return nil
	}
}

func (h *TelegramHandler) handleInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Проверяем, выбраны ли интересы
	selectedInterests, err := h.service.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		log.Printf("Error getting selected interests: %v", err)
		return err
	}

	// Если не выбрано ни одного интереса, сообщаем пользователю и оставляем клавиатуру
	if len(selectedInterests) == 0 {
		warningMsg := "❗ " + h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_at_least_one_interest")
		if warningMsg == "choose_at_least_one_interest" { // fallback if key doesn't exist
			warningMsg = "❗ Пожалуйста, выберите хотя бы один интерес"
		}

		// Добавляем оригинальный текст с предупреждением
		chooseInterestsText := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")
		fullText := warningMsg + "\n\n" + chooseInterestsText

		// Получаем интересы и оставляем клавиатуру с интересами видимой, обновляя только текст
		interests, _ := h.service.Localizer.GetInterests(user.InterfaceLanguageCode)
		keyboard := h.createInterestsKeyboard(interests, []int{}, user.InterfaceLanguageCode)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			fullText,
			keyboard,
		)
		_, err := h.bot.Request(editMsg)
		return err
	}

	// Если интересы выбраны, завершаем профиль
	completedMsg := h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_completed")
	keyboard := h.createProfileCompletedKeyboard(user.InterfaceLanguageCode)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		completedMsg,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	if err != nil {
		return err
	}

	// Обновляем статус пользователя
	err = h.service.DB.UpdateUserState(user.ID, models.StateActive)
	if err != nil {
		log.Printf("Error updating user state: %v", err)
		return err
	}
	err = h.service.DB.UpdateUserStatus(user.ID, models.StatusActive)
	if err != nil {
		log.Printf("Error updating user status: %v", err)
		return err
	}

	// Увеличиваем уровень завершения профиля до 100%
	err = h.updateProfileCompletionLevel(user.ID, 100)
	if err != nil {
		log.Printf("Error updating profile completion level: %v", err)
		return err
	}

	return nil
}

// updateProfileCompletionLevel обновляет уровень завершения профиля от 0 до 100
func (h *TelegramHandler) updateProfileCompletionLevel(userID int, completionLevel int) error {
	_, err := h.service.DB.GetConnection().Exec(`
		UPDATE users
		SET profile_completion_level = $1, updated_at = NOW()
		WHERE id = $2
	`, completionLevel, userID)
	return err
}

// startProfileSetup начинает настройку профиля сразу с выбора родного языка
func (h *TelegramHandler) startProfileSetup(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)

	// Отправляем новое сообщение для начала настройки профиля
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.bot.Send(msg)
	return err
}

// handleBackToMainMenu возвращает пользователя в главное меню
func (h *TelegramHandler) handleBackToMainMenu(callback *tgbotapi.CallbackQuery, user *models.User) error {
	welcomeText := h.service.GetWelcomeMessage(user)
	menuText := welcomeText + "\n\n" + h.service.Localizer.Get(user.InterfaceLanguageCode, "main_menu_title")

	keyboard := h.createMainMenuKeyboard(user.InterfaceLanguageCode)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		menuText,
		keyboard,
	)
	_, err := h.bot.Request(editMsg)
	return err
}

// Обработчик продолжения заполнения профиля после подтверждения выбора языков
func (h *TelegramHandler) handleLanguagesContinueFilling(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Очищаем старые интересы при переходе к выбору интересов
	err := h.service.DB.ClearUserInterests(user.ID)
	if err != nil {
		log.Printf("Warning: could not clear user interests: %v", err)
	}

	// Предлагаем выбрать уровень владения языком
	langName := h.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
	title := h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := h.createLanguageLevelKeyboard(user.InterfaceLanguageCode, user.TargetLanguageCode)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	return err
}

// Обработчик повторного выбора языков
func (h *TelegramHandler) handleLanguagesReselect(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Сбрасываем выбор языков
	user.NativeLanguageCode = ""
	user.TargetLanguageCode = ""
	user.TargetLanguageLevel = ""

	// Обновляем статус пользователя
	_ = h.service.DB.UpdateUserNativeLanguage(user.ID, "")
	_ = h.service.DB.UpdateUserTargetLanguage(user.ID, "")
	_ = h.service.DB.UpdateUserTargetLanguageLevel(user.ID, "")

	// Предлагаем выбрать родной язык снова
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := h.bot.Request(editMsg)
	return err
}

// Обработчик выбора уровня владения языком
func (h *TelegramHandler) handleLanguageLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error {
	// Сохраняем уровень владения языком
	err := h.service.DB.UpdateUserTargetLanguageLevel(user.ID, levelCode)
	if err != nil {
		return err
	}

	user.TargetLanguageLevel = levelCode

	// Получаем локализованное название уровня
	levelName := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_level_"+levelCode)

	// Подтверждаем выбор уровня
	confirmMsg := fmt.Sprintf("🎯 %s\n\n%s",
		levelName,
		h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests"))

	// Получаем интересы и создаем клавиатуру с пустым списком выбранных
	interests, _ := h.service.Localizer.GetInterests(user.InterfaceLanguageCode)
	keyboard := h.createInterestsKeyboard(interests, []int{}, user.InterfaceLanguageCode)

	// Редактируем сообщение с новой клавиатурой
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		confirmMsg,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	return err
}

// Обработчик кнопки "Назад"
func (h *TelegramHandler) handleBackToPreviousStep(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// В зависимости от текущего состояния пользователя, возвращаемся к предыдущему шагу
	switch user.State {
	case models.StateWaitingTargetLanguage:
		// Возвращаемся к выбору родного языка
		user.NativeLanguageCode = ""
		_ = h.service.DB.UpdateUserNativeLanguage(user.ID, "")

		text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
		keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			keyboard,
		)
		_, err := h.bot.Request(editMsg)
		return err

	case models.StateWaitingLanguageLevel:
		// Возвращаемся к выбору изучаемого языка
		if user.NativeLanguageCode == "ru" {
			// Если родной язык русский, возвращаем к выбору изучаемого языка
			user.TargetLanguageCode = ""
			_ = h.service.DB.UpdateUserTargetLanguage(user.ID, "")

			text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")
			// Исключаем русский из списка изучаемых языков
			keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "target", "ru", true)
			editMsg := tgbotapi.NewEditMessageTextAndMarkup(
				callback.Message.Chat.ID,
				callback.Message.MessageID,
				text,
				keyboard,
			)
			_, err := h.bot.Request(editMsg)
			return err
		} else {
			// Если родной язык не русский, возвращаем к выбору родного языка
			// потому что для не русского родного сразу устанавливается русский как изучаемый
			user.NativeLanguageCode = ""
			user.TargetLanguageCode = ""
			user.TargetLanguageLevel = ""

			_ = h.service.DB.UpdateUserNativeLanguage(user.ID, "")
			_ = h.service.DB.UpdateUserTargetLanguage(user.ID, "")
			_ = h.service.DB.UpdateUserTargetLanguageLevel(user.ID, "")

			text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
			keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)
			editMsg := tgbotapi.NewEditMessageTextAndMarkup(
				callback.Message.Chat.ID,
				callback.Message.MessageID,
				text,
				keyboard,
			)
			_, err := h.bot.Request(editMsg)
			return err
		}

	case models.StateWaitingInterests:
		// Возвращаемся к выбору уровня владения языком
		user.TargetLanguageLevel = ""
		_ = h.service.DB.UpdateUserTargetLanguageLevel(user.ID, "")

		// Предлагаем выбрать уровень владения языком
		langName := h.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
		title := h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
			"language": langName,
		})

		keyboard := h.createLanguageLevelKeyboard(user.InterfaceLanguageCode, user.TargetLanguageCode)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			title,
			keyboard,
		)
		_, err := h.bot.Request(editMsg)
		return err

	default:
		// По умолчанию возвращаем к выбору родного языка
		user.NativeLanguageCode = ""
		user.TargetLanguageCode = ""
		user.TargetLanguageLevel = ""

		_ = h.service.DB.UpdateUserNativeLanguage(user.ID, "")
		_ = h.service.DB.UpdateUserTargetLanguage(user.ID, "")
		_ = h.service.DB.UpdateUserTargetLanguageLevel(user.ID, "")

		text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
		keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			keyboard,
		)
		_, err := h.bot.Request(editMsg)
		return err
	}
}

// ✨ Выбор родного языка
func (h *TelegramHandler) handleNativeLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_native_"):]

	err := h.service.DB.UpdateUserNativeLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	user.NativeLanguageCode = langCode

	// Обновляем статус пользователя
	h.service.DB.UpdateUserState(user.ID, models.StateWaitingLanguage)

	// Переход к следующему шагу онбординга
	return h.proceedToNextOnboardingStep(callback, user, langCode)
}

// proceedToNextOnboardingStep определяет следующий шаг после выбора родного языка
func (h *TelegramHandler) proceedToNextOnboardingStep(callback *tgbotapi.CallbackQuery, user *models.User, nativeLangCode string) error {
	if nativeLangCode == "ru" {
		// Если выбран русский как родной, предлагаем выбрать изучаемый язык
		text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")

		// Исключаем русский из списка изучаемых языков
		keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "target", "ru", true)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, text, keyboard)
		_, err := h.bot.Request(editMsg)
		if err != nil {
			return err
		}

		// Обновляем статус для ожидания выбора изучаемого языка
		h.service.DB.UpdateUserState(user.ID, models.StateWaitingTargetLanguage)
		return nil
	} else {
		// Для всех других языков как родных автоматически устанавливаем русский как изучаемый
		err := h.service.DB.UpdateUserTargetLanguage(user.ID, "ru")
		if err != nil {
			return err
		}
		user.TargetLanguageCode = "ru"

		// Получаем название выбранного языка для сообщения
		nativeLangName := h.service.Localizer.GetLanguageName(nativeLangCode, user.InterfaceLanguageCode)

		// Показываем сообщение о том, что русский язык установлен автоматически
		targetExplanation := h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "target_language_explanation", map[string]string{
			"native_lang": nativeLangName,
		})

		// Предлагаем выбрать уровень владения русским языком
		langName := h.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
		levelTitle := targetExplanation + "\n\n" + h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
			"language": langName,
		})

		keyboard := h.createLanguageLevelKeyboard(user.InterfaceLanguageCode, user.TargetLanguageCode)
		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			levelTitle,
			keyboard,
		)
		_, err = h.bot.Request(editMsg)
		if err != nil {
			return err
		}

		// Обновляем статус для ожидания выбора уровня
		h.service.DB.UpdateUserState(user.ID, models.StateWaitingLanguageLevel)
		return nil
	}
}

func (h *TelegramHandler) handleTargetLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_target_"):]
	err := h.service.DB.UpdateUserTargetLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	// ✅ ОЧИЩАЕМ СТАРЫЕ ИНТЕРЕСЫ при переходе к выбору интересов
	err = h.service.DB.ClearUserInterests(user.ID)
	if err != nil {
		log.Printf("Warning: could not clear user interests: %v", err)
	}

	user.TargetLanguageCode = langCode
	langName := h.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)

	// Предлагаем выбрать уровень владения языком
	title := h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := h.createLanguageLevelKeyboard(user.InterfaceLanguageCode, langCode)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	return err
}

func (h *TelegramHandler) handleInterfaceLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error {
	if err := h.service.DB.UpdateUserInterfaceLanguage(user.ID, langCode); err != nil {
		log.Printf("Error updating interface language: %v", err)
		return err
	}

	// Обновляем язык интерфейса пользователя и получаем новое сообщение
	user.InterfaceLanguageCode = langCode
	langName := h.service.Localizer.GetLanguageName(langCode, langCode)
	text := fmt.Sprintf("%s\n\n%s: %s",
		h.service.Localizer.Get(langCode, "choose_interface_language"),
		h.service.Localizer.Get(langCode, "language_updated"),
		langName,
	)

	// Создаем клавиатуру с языками интерфейса (остальные кнопки остаются)
	keyboard := h.createLanguageKeyboard(langCode, "interface", "", true)

	// Редактируем сообщение, сохраняя клавиатуру
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := h.bot.Request(editMsg)
	return err
}

func (h *TelegramHandler) handleInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		log.Printf("Error parsing interest ID: %v", err)
		return err
	}

	// Получаем текущие выбранные интересы пользователя
	selectedInterests, err := h.service.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		log.Printf("Error getting user interests, using empty list: %v", err)
		selectedInterests = []int{} // fallback
	}

	// Переключаем интерес (toggle)
	isCurrentlySelected := false
	for i, id := range selectedInterests {
		if id == interestID {
			// Убираем из списка
			selectedInterests = append(selectedInterests[:i], selectedInterests[i+1:]...)
			isCurrentlySelected = true
			break
		}
	}

	if !isCurrentlySelected {
		// Добавляем в список
		selectedInterests = append(selectedInterests, interestID)
		err = h.service.DB.SaveUserInterest(user.ID, interestID, false)
	} else {
		// Удаляем интерес из БД
		err = h.service.DB.RemoveUserInterest(user.ID, interestID)
	}

	if err != nil {
		log.Printf("Error updating user interest: %v", err)
		return err
	}

	// ✅ Обновляем только клавиатуру - никаких новых сообщений
	interests, _ := h.service.Localizer.GetInterests(user.InterfaceLanguageCode)
	keyboard := h.createInterestsKeyboard(interests, selectedInterests, user.InterfaceLanguageCode)

	editMsg := tgbotapi.NewEditMessageReplyMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	return err
}

func (h *TelegramHandler) sendTargetLanguageMenu(chatID int64, user *models.User) error {
	// Исключаем родной язык из списка изучаемых
	excludeLang := user.NativeLanguageCode
	if excludeLang == "" {
		excludeLang = "ru" // По умолчанию исключаем русский
	}

	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "target", excludeLang, true)
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.bot.Send(msg)
	return err
}

// В createInterestsKeyboard нужно передать язык интерфейса
func (h *TelegramHandler) sendInterestsMenu(chatID int64, user *models.User) error {
	interests, err := h.service.Localizer.GetInterests(user.InterfaceLanguageCode)
	if err != nil {
		return err
	}

	// Загружаем уже выбранные интересы пользователя
	selectedInterests, err := h.service.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		log.Printf("Error loading user interests: %v", err)
		selectedInterests = []int{} // fallback на пустой список
	}

	keyboard := h.createInterestsKeyboard(interests, selectedInterests, user.InterfaceLanguageCode)
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests")
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err = h.bot.Send(msg)
	return err
}

func (h *TelegramHandler) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	return err
}

func (h *TelegramHandler) createProfileMenuKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	// Кнопки для управления профилем
	editInterests := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "profile_edit_interests"),
		"edit_interests",
	)
	editLanguages := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "profile_edit_languages"),
		"edit_languages",
	)
	changeInterfaceLang := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "main_menu_change_lang"),
		"main_change_language",
	)
	reconfig := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "profile_reconfigure"),
		"profile_reset_ask",
	)
	backToMain := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "back_to_main"),
		"back_to_main_menu",
	)

	// Пять рядов: интересы, языки, язык интерфейса, сброс, главное меню
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{editInterests},
		{editLanguages},
		{changeInterfaceLang},
		{reconfig},
		{backToMain},
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func (h *TelegramHandler) createMainMenuKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	viewProfile := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "main_menu_view_profile"),
		"main_view_profile",
	)
	editProfile := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "main_menu_edit_profile"),
		"main_edit_profile",
	)
	changeLang := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "main_menu_change_lang"),
		"main_change_language",
	)
	feedback := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "main_menu_feedback"),
		"main_feedback",
	)

	// Компонуем меню по 2 кнопки в ряд для лучшей организации
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{viewProfile, editProfile},
		{changeLang, feedback},
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func (h *TelegramHandler) createResetConfirmKeyboard(interfaceLang string) tgbotapi.InlineKeyboardMarkup {
	yes := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "profile_reset_yes"),
		"profile_reset_yes",
	)
	no := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "profile_reset_no"),
		"profile_reset_no",
	)
	return tgbotapi.NewInlineKeyboardMarkup([][]tgbotapi.InlineKeyboardButton{{yes}, {no}}...)
}

// Обработчики главного меню
func (h *TelegramHandler) handleMainChangeLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interface_language")
	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "interface", "", true)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := h.bot.Request(editMsg)
	return err
}

func (h *TelegramHandler) handleMainViewProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Проверяем, заполнен ли профиль по уровню завершения профиля
	if user.ProfileCompletionLevel == 0 {
		// Профиль не заполнен - показываем информационное сообщение и кнопку настройки
		text := h.service.Localizer.Get(user.InterfaceLanguageCode, "empty_profile_message")

		// Создаем клавиатуру с кнопкой настройки профиля
		setupButton := tgbotapi.NewInlineKeyboardButtonData(
			h.service.Localizer.Get(user.InterfaceLanguageCode, "setup_profile_button"),
			"start_profile_setup",
		)

		keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{setupButton})

		// Отправляем новое сообщение вместо редактирования существующего
		newMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
		newMsg.ReplyMarkup = keyboard
		_, err := h.bot.Send(newMsg)
		return err
	}

	// Профиль заполнен - показываем его
	return h.handleProfileShow(callback, user)
}

func (h *TelegramHandler) handleMainEditProfile(callback *tgbotapi.CallbackQuery, user *models.User) error {
	return h.handleProfileResetAsk(callback, user)
}

func (h *TelegramHandler) handleMainFeedback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Создаем message объект для handleFeedbackCommand
	message := &tgbotapi.Message{
		Chat: callback.Message.Chat,
	}
	return h.handleFeedbackCommand(message, user)
}

// Команда /feedback
func (h *TelegramHandler) handleFeedbackCommand(message *tgbotapi.Message, user *models.User) error {
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_text")
	_ = h.service.DB.UpdateUserState(user.ID, models.StateWaitingFeedback)
	return h.sendMessage(message.Chat.ID, text)
}

// Команда /profile — показать профиль в любой момент.
func (h *TelegramHandler) handleProfileCommand(message *tgbotapi.Message, user *models.User) error {
	summary, err := h.service.BuildProfileSummary(user)
	if err != nil {
		return h.sendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
	}
	text := summary + "\n\n" + h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = h.createProfileMenuKeyboard(user.InterfaceLanguageCode)
	_, err = h.bot.Send(msg)
	return err
}

// Колбэки профиля: показать профиль
func (h *TelegramHandler) handleProfileShow(callback *tgbotapi.CallbackQuery, user *models.User) error {
	summary, err := h.service.BuildProfileSummary(user)
	if err != nil {
		return err
	}
	text := summary + "\n\n" + h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		h.createProfileMenuKeyboard(user.InterfaceLanguageCode),
	)
	_, err = h.bot.Request(edit)
	return err
}

// Спросить подтверждение сброса
func (h *TelegramHandler) handleProfileResetAsk(callback *tgbotapi.CallbackQuery, user *models.User) error {
	title := h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_title")
	warn := h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_warning")
	text := fmt.Sprintf("%s\n\n%s", title, warn)
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		h.createResetConfirmKeyboard(user.InterfaceLanguageCode),
	)
	_, err := h.bot.Request(edit)
	return err
}

// Подтверждение сброса
func (h *TelegramHandler) handleProfileResetYes(callback *tgbotapi.CallbackQuery, user *models.User) error {
	if err := h.service.DB.ResetUserProfile(user.ID); err != nil {
		return err
	}
	// Обновляем в памяти базовые поля
	user.NativeLanguageCode = ""
	user.TargetLanguageCode = ""
	user.State = models.StateWaitingLanguage
	user.Status = models.StatusFilling
	user.ProfileCompletionLevel = 0

	done := h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_reset_done")
	// Предложим сразу начать с выбора родного языка
	next := h.service.GetLanguagePrompt(user, "native")
	text := done + "\n\n" + next

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		h.createLanguageKeyboard(user.InterfaceLanguageCode, "native", "", true),
	)
	_, err := h.bot.Request(edit)
	return err
}

// Отмена сброса — вернёмся в главное меню
func (h *TelegramHandler) handleProfileResetNo(callback *tgbotapi.CallbackQuery, user *models.User) error {
	return h.handleBackToMainMenu(callback, user)
}

// handleEditInterests позволяет редактировать интересы пользователя
func (h *TelegramHandler) handleEditInterests(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все доступные интересы
	interests, err := h.service.Localizer.GetInterests(user.InterfaceLanguageCode)
	if err != nil {
		return err
	}

	// Получаем текущие выбранные интересы пользователя и сохраняем в кэше
	selectedInterests, err := h.service.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		log.Printf("Error loading user interests: %v", err)
		selectedInterests = []int{} // fallback
	}

	// Инициализируем временное хранилище для сессии редактирования
	userID := int64(user.ID)
	h.editInterestsTemp[userID] = make([]int, len(selectedInterests))
	copy(h.editInterestsTemp[userID], selectedInterests)

	keyboard := h.createEditInterestsKeyboard(interests, selectedInterests, user.InterfaceLanguageCode)
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests") +
		"\n\n" + h.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + h.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	return err
}

// handleEditLanguages позволяет редактировать языки пользователя
func (h *TelegramHandler) handleEditLanguages(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Показываем текущие настройки языков с кнопками редактирования
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_edit_languages") +
		"\n\n" + h.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + h.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	keyboard := h.createEditLanguagesKeyboard(
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
	_, err := h.bot.Request(editMsg)
	return err
}

// handleSaveEdits сохраняет изменения и возвращается к просмотру профиля
func (h *TelegramHandler) handleSaveEdits(callback *tgbotapi.CallbackQuery, user *models.User) error {
	userID := int64(user.ID)

	// Если есть временное хранилище интересов, применяем изменения в БД
	if tempInterests, exists := h.editInterestsTemp[userID]; exists {
		// Очищаем текущие интересы в БД
		err := h.service.DB.ClearUserInterests(user.ID)
		if err != nil {
			log.Printf("Error clearing user interests: %v", err)
			return err
		}

		// Сохраняем новые интересы
		for _, interestID := range tempInterests {
			err := h.service.DB.SaveUserInterest(user.ID, interestID, false)
			if err != nil {
				log.Printf("Error saving user interest %d: %v", interestID, err)
				return err
			}
		}

		// Удаляем временное хранилище
		delete(h.editInterestsTemp, userID)
	}

	// Показываем профиль с обновленными данными
	return h.handleProfileShow(callback, user)
}

// handleCancelEdits отменяет изменения и возвращается к просмотру профиля
func (h *TelegramHandler) handleCancelEdits(callback *tgbotapi.CallbackQuery, user *models.User) error {
	userID := int64(user.ID)

	// Очищаем временное хранилище без применения изменений
	if _, exists := h.editInterestsTemp[userID]; exists {
		delete(h.editInterestsTemp, userID)
	}

	// Просто показываем профиль без изменений
	return h.handleProfileShow(callback, user)
}

// handleEditNativeLang редактирует родной язык пользователя
func (h *TelegramHandler) handleEditNativeLang(callback *tgbotapi.CallbackQuery, user *models.User) error {
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_native_language")
	// Показываем клавиатуру с сохранением/отменой вместо обычного выбора
	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "edit_native", "", false)

	// Добавляем кнопки сохранить/отменить
	saveRow := h.createSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := h.bot.Request(editMsg)
	return err
}

// handleEditTargetLang редактирует изучаемый язык пользователя (только если родной - русский)
func (h *TelegramHandler) handleEditTargetLang(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Проверяем, что родной язык русский - только в этом случае можно редактировать изучаемый язык
	if user.NativeLanguageCode != "ru" {
		// Не должно происходить по логике, но на всякий случай
		return h.sendMessage(callback.Message.Chat.ID, "Редактирование изучаемого языка недоступно при вашем родном языке.")
	}

	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")
	// Исключаем родной язык из списка изучаемых
	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "edit_target", user.NativeLanguageCode, false)

	// Добавляем кнопки сохранить/отменить
	saveRow := h.createSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err := h.bot.Request(editMsg)
	return err
}

// handleEditNativeLanguage сохраняет выбор родного языка с учетом первоначальной логики
func (h *TelegramHandler) handleEditNativeLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_edit_native_"):]

	// Сохраняем новый родной язык
	err := h.service.DB.UpdateUserNativeLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	user.NativeLanguageCode = langCode

	// Применяем изначальную логику выбора языков
	if langCode == "ru" {
		// Если выбран русский как родной, предлагаем выбрать изучаемый из оставшихся 3
		// Но не меняем существующий изучаемый язык, если он есть
		text := "Выберите изучаемый язык:"
		keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "edit_target", "ru", false)

		// Добавляем кнопки сохранить/отменить
		saveRow := h.createSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
			keyboard,
		)
		_, err := h.bot.Request(editMsg)
		return err
	} else {
		// Если выбран не русский, автоматически устанавливаем русский как изучаемый
		err := h.service.DB.UpdateUserTargetLanguage(user.ID, "ru")
		if err != nil {
			return err
		}
		user.TargetLanguageCode = "ru"

		// Показываем сообщение о том, что русский язык установлен автоматически
		nativeLangName := h.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)
		explanation := h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "target_language_explanation", map[string]string{
			"native_lang": nativeLangName,
		})

		// Предлагаем выбрать уровень владения русским языком
		langName := h.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
		levelTitle := explanation + "\n\n" + h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
			"language": langName,
		})

		keyboard := h.createLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, user.TargetLanguageCode, "edit_level_", false)
		// Добавляем сохранить/отменить
		saveRow := h.createSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			levelTitle,
			keyboard,
		)
		_, err = h.bot.Request(editMsg)
		return err
	}
}

// handleEditTargetLanguage сохраняет выбор изучаемого языка и предлагает выбрать уровень
func (h *TelegramHandler) handleEditTargetLanguage(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_edit_target_"):]

	err := h.service.DB.UpdateUserTargetLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	user.TargetLanguageCode = langCode
	langName := h.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)

	// Предлагаем выбрать уровень владения языком
	title := h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := h.createLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, langCode, "edit_level_", false)
	// Добавляем save/cancel
	saveRow := h.createSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	return err
}

func (h *TelegramHandler) handleEditInterestSelection(callback *tgbotapi.CallbackQuery, user *models.User, interestIDStr string) error {
	interestID, err := strconv.Atoi(interestIDStr)
	if err != nil {
		log.Printf("Error parsing interest ID: %v", err)
		return err
	}

	userID := int64(user.ID)

	// Если временного хранилища нет, инициализируем его
	if _, exists := h.editInterestsTemp[userID]; !exists {
		selectedInterests, err := h.service.DB.GetUserSelectedInterests(user.ID)
		if err != nil {
			log.Printf("Error getting user interests, using empty list: %v", err)
			selectedInterests = []int{}
		}
		h.editInterestsTemp[userID] = make([]int, len(selectedInterests))
		copy(h.editInterestsTemp[userID], selectedInterests)
	}

	// Переключаем интерес в временном хранилище (toggle)
	isCurrentlySelected := false
	for i, id := range h.editInterestsTemp[userID] {
		if id == interestID {
			// Убираем из списка
			h.editInterestsTemp[userID] = append(h.editInterestsTemp[userID][:i], h.editInterestsTemp[userID][i+1:]...)
			isCurrentlySelected = true
			break
		}
	}

	if !isCurrentlySelected {
		// Добавляем в список
		h.editInterestsTemp[userID] = append(h.editInterestsTemp[userID], interestID)
	}

	// ✅ Возвращаем edit клавиатуру с временными изменениями
	interests, _ := h.service.Localizer.GetInterests(user.InterfaceLanguageCode)
	keyboard := h.createEditInterestsKeyboard(interests, h.editInterestsTemp[userID], user.InterfaceLanguageCode)
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests") +
		"\n\n" + h.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + h.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	return err
}

func (h *TelegramHandler) handleEditLevelSelection(callback *tgbotapi.CallbackQuery, user *models.User, levelCode string) error {
	// Сохраняем уровень владения языком
	err := h.service.DB.UpdateUserTargetLanguageLevel(user.ID, levelCode)
	if err != nil {
		return err
	}

	user.TargetLanguageLevel = levelCode

	// Переходим к меню редактирования языков с обновленными данными
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_edit_languages") +
		"\n\n" + h.service.Localizer.Get(user.InterfaceLanguageCode, "save_button") +
		" или " + h.service.Localizer.Get(user.InterfaceLanguageCode, "cancel_button")

	keyboard := h.createEditLanguagesKeyboard(
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
	_, err = h.bot.Request(editMsg)
	return err
}

// handleEditLevelLang редактирует уровень владения языком
func (h *TelegramHandler) handleEditLevelLang(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langName := h.service.Localizer.GetLanguageName(user.TargetLanguageCode, user.InterfaceLanguageCode)
	title := h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "choose_level_title", map[string]string{
		"language": langName,
	})

	keyboard := h.createLanguageLevelKeyboardWithPrefix(user.InterfaceLanguageCode, user.TargetLanguageCode, "edit_level_", false)
	// Добавляем save/cancel
	saveRow := h.createSaveEditsKeyboard(user.InterfaceLanguageCode).InlineKeyboard[0]
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{saveRow[0], saveRow[1]})

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		title,
		keyboard,
	)
	_, err := h.bot.Request(editMsg)
	return err
}
