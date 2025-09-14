package telegram

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

type TelegramHandler struct {
	bot               *tgbotapi.BotAPI
	service           *core.BotService
	editInterestsTemp map[int64][]int // Временное хранение выбранных интересов для каждого пользователя
	adminChatIDs      []int64         // Chat ID администраторов
	adminUsernames    []string        // Usernames администраторов для проверки доступа
}

func NewTelegramHandler(bot *tgbotapi.BotAPI, service *core.BotService, adminChatIDs []int64) *TelegramHandler {
	return &TelegramHandler{
		bot:               bot,
		service:           service,
		editInterestsTemp: make(map[int64][]int),
		adminChatIDs:      adminChatIDs,
		adminUsernames:    make([]string, 0), // пустой список нет хардкода
	}
}

func NewTelegramHandlerWithAdmins(bot *tgbotapi.BotAPI, service *core.BotService, adminChatIDs []int64, adminUsernames []string) *TelegramHandler {
	return &TelegramHandler{
		bot:               bot,
		service:           service,
		editInterestsTemp: make(map[int64][]int),
		adminChatIDs:      adminChatIDs,
		adminUsernames:    adminUsernames,
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
	case "feedbacks":
		return h.handleFeedbacksCommand(message, user)
	default:
		log.Printf("Unknown command: %s", message.Command())
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
	case models.StateWaitingFeedback:
		return h.handleFeedbackMessage(message, user)
	case models.StateWaitingFeedbackContact:
		return h.handleFeedbackContactMessage(message, user)
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
	case strings.HasPrefix(data, "fb_process_"):
		feedbackIDStr := strings.TrimPrefix(data, "fb_process_")
		return h.handleFeedbackProcess(callback, user, feedbackIDStr)
	case strings.HasPrefix(data, "fb_unprocess_"):
		feedbackIDStr := strings.TrimPrefix(data, "fb_unprocess_")
		return h.handleFeedbackUnprocess(callback, user, feedbackIDStr)
	case strings.HasPrefix(data, "fb_delete_"):
		feedbackIDStr := strings.TrimPrefix(data, "fb_delete_")
		return h.handleFeedbackDelete(callback, user, feedbackIDStr)
	case strings.HasPrefix(data, "browse_active_feedbacks_"):
		indexStr := strings.TrimPrefix(data, "browse_active_feedbacks_")
		return h.handleBrowseActiveFeedbacks(callback, user, indexStr)
	case strings.HasPrefix(data, "browse_archive_feedbacks_"):
		indexStr := strings.TrimPrefix(data, "browse_archive_feedbacks_")
		return h.handleBrowseArchiveFeedbacks(callback, user, indexStr)
	case strings.HasPrefix(data, "browse_all_feedbacks_"):
		indexStr := strings.TrimPrefix(data, "browse_all_feedbacks_")
		return h.handleBrowseAllFeedbacks(callback, user, indexStr)
	case strings.HasPrefix(data, "feedback_prev_"):
		parts := strings.TrimPrefix(data, "feedback_prev_")
		indexAndType := strings.Split(parts, "_")
		if len(indexAndType) == 2 {
			return h.handleFeedbackPrev(callback, user, indexAndType[0], indexAndType[1])
		}
		return nil
	case strings.HasPrefix(data, "feedback_next_"):
		parts := strings.TrimPrefix(data, "feedback_next_")
		indexAndType := strings.Split(parts, "_")
		if len(indexAndType) == 2 {
			return h.handleFeedbackNext(callback, user, indexAndType[0], indexAndType[1])
		}
		return nil
	case strings.HasPrefix(data, "feedback_back_"):
		feedbackType := strings.TrimPrefix(data, "feedback_back_")
		return h.handleFeedbackBack(callback, user, feedbackType)
	case data == "show_active_feedbacks":
		return h.handleShowActiveFeedbacks(callback, user)
	case data == "show_archive_feedbacks":
		return h.handleShowArchiveFeedbacks(callback, user)
	case data == "show_all_feedbacks":
		return h.handleShowAllFeedbacks(callback, user)
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

// Команда /feedbacks — просмотр отзывов (доступна только администраторам)
func (h *TelegramHandler) handleFeedbacksCommand(message *tgbotapi.Message, user *models.User) error {
	// Проверяем права администратора по Chat ID и username
	isAdminByID := false
	isAdminByUsername := false

	// Проверяем по Chat ID
	for _, adminID := range h.adminChatIDs {
		if message.Chat.ID == adminID {
			isAdminByID = true
			break
		}
	}

	// Проверяем по username
	if user.Username != "" {
		for _, adminUsername := range h.adminUsernames {
			cleanUsername := strings.TrimPrefix(adminUsername, "@")
			if user.Username == cleanUsername {
				isAdminByUsername = true
				break
			}
		}
	}

	if !isAdminByID && !isAdminByUsername {
		log.Printf("❌ Отказано в доступе: пользователь %s (ID: %d, ChatID: %d) пытается использовать /feedbacks",
			user.Username, user.ID, message.Chat.ID)
		return h.sendMessage(message.Chat.ID, "❌ Данная команда доступна только администраторам бота.")
	}

	// Получаем все отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return h.sendMessage(message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		return h.sendMessage(message.Chat.ID, "📝 Отзывов пока нет")
	}

	// Группируем отзывы по содержимому и пользователю, чтобы убрать дубли
	type feedbackKey struct {
		userID       int64
		feedbackText string
	}
	seen := make(map[feedbackKey][]map[string]interface{})
	for _, fb := range feedbacks {
		key := feedbackKey{
			userID:       fb["telegram_id"].(int64),
			feedbackText: fb["feedback_text"].(string),
		}
		seen[key] = append(seen[key], fb)
	}

	// Разделяем отзывы на обработанные и необработанные
	var processedFeedbacks []map[string]interface{}
	var unprocessedFeedbacks []map[string]interface{}

	totalFeedbacks := len(seen)
	processedCount := 0
	unprocessedCount := 0
	shortCount := 0
	longCount := 0
	contactCount := 0

	for _, group := range seen {
		// Берем наиболее свежий отзыв из группы
		latest := group[0]
		for _, fb := range group {
			if fb["created_at"].(time.Time).After(latest["created_at"].(time.Time)) {
				latest = fb
			}
		}

		charCount := len([]rune(strings.ReplaceAll(latest["feedback_text"].(string), "\n", " ")))

		// Определяем характеристики отзыва для статистики
		if charCount < 50 {
			shortCount++
		} else if charCount > 200 {
			longCount++
		}

		if latest["is_processed"].(bool) {
			processedCount++
			processedFeedbacks = append(processedFeedbacks, latest)
		} else {
			unprocessedCount++
			unprocessedFeedbacks = append(unprocessedFeedbacks, latest)
		}

		// Подсчет контактов
		if latest["contact_info"] != nil && latest["contact_info"].(string) != "" {
			contactCount++
		}
	}

	// Отправляем компактную статистику с кнопками управления
	mediumCount := totalFeedbacks - shortCount - longCount

	statsMessage := fmt.Sprintf(
		"📊 Отзывы - Статистика:\n\n"+
			"⏳ Обработка:\n"+
			"- Всего отзывов: %d\n"+
			"- 🆕 Активных (необработанных): %d\n"+
			"- ✅ В архиве (обработанных): %d\n\n"+
			"📏 По длине:\n"+
			"- 📝 Короткие (< 50 симв.): %d\n"+
			"- 📊 Средние (50-200 симв.): %d\n"+
			"- 📖 Длинные (> 200 симв.): %d\n\n"+
			"📞 С контактными данными: %d",
		totalFeedbacks, unprocessedCount, processedCount,
		shortCount, mediumCount, longCount, contactCount,
	)

	// Кнопки для интерактивного управления отзывами
	var buttons [][]tgbotapi.InlineKeyboardButton
	if unprocessedCount > 0 {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("🆕 Просмотреть активные "+fmt.Sprintf("(%d)", unprocessedCount), "browse_active_feedbacks_0"),
		})
	}
	if processedCount > 0 {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("📚 Просмотреть архив "+fmt.Sprintf("(%d)", processedCount), "browse_archive_feedbacks_0"),
		})
	}
	if len(seen) > 0 {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("📋 Просмотреть все "+fmt.Sprintf("(%d)", totalFeedbacks), "browse_all_feedbacks_0"),
		})
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(message.Chat.ID, statsMessage)
	msg.ReplyMarkup = keyboard
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки статистики: %v", err)
		return err
	}

	return nil
}

// Вспомогательная функция для отправки списка отзывов
func (h *TelegramHandler) sendFeedbackList(chatID int64, feedbackList []map[string]interface{}) error {
	for _, feedback := range feedbackList {
		if err := h.sendFeedbackItem(chatID, feedback); err != nil {
			return err
		}
	}
	return nil
}

// Вспомогательная функция для отправки одного отзыва
func (h *TelegramHandler) sendFeedbackItem(chatID int64, fb map[string]interface{}) error {
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

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки отзыва ID %d: %v", feedbackID, err)
		// Fallback без HTML
		plainMessage := fmt.Sprintf(
			"%s %s %s\n"+
				"Автор: %s\n"+
				"Статус: %s (%d символов)\n"+
				"Дата: %s%s\n\n"+
				"Содержание отзыва:\n%s",
			statusIcon, firstName, username,
			statusText,
			charIcon,
			charCount,
			dateStr,
			contactStr,
			feedbackTextContent,
		)
		plainMsg := tgbotapi.NewMessage(chatID, plainMessage)
		plainMsg.ReplyMarkup = keyboard
		if _, plainErr := h.bot.Send(plainMsg); plainErr != nil {
			log.Printf("Критичная ошибка отправки отзыва ID %d без HTML: %v", feedbackID, plainErr)
			return plainErr
		}
	}
	return nil
}

// === ОБРАБОТЧИКИ ВИДОВ ОТЗЫВОВ ===

// handleShowActiveFeedbacks показывает только необработанные отзывы
func (h *TelegramHandler) handleShowActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		return h.sendMessage(callback.Message.Chat.ID, "📝 Отзывов пока нет")
	}

	// Группируем и фильтруем только необработанные отзывы
	type feedbackKey struct {
		userID       int64
		feedbackText string
	}
	seen := make(map[feedbackKey][]map[string]interface{})
	for _, fb := range feedbacks {
		if !fb["is_processed"].(bool) { // Только необработанные
			key := feedbackKey{
				userID:       fb["telegram_id"].(int64),
				feedbackText: fb["feedback_text"].(string),
			}
			seen[key] = append(seen[key], fb)
		}
	}

	if len(seen) == 0 {
		return h.sendMessage(callback.Message.Chat.ID, "🎉 Все отзывы обработаны!")
	}

	// Отправляем заголовок
	headerMsg := tgbotapi.NewMessage(callback.Message.Chat.ID,
		fmt.Sprintf("🏷️ <b>Необработанные отзывы (%d):</b>", len(seen)))
	headerMsg.ParseMode = tgbotapi.ModeHTML
	if _, err := h.bot.Send(headerMsg); err != nil {
		log.Printf("Ошибка отправки заголовка активных отзывов: %v", err)
	}

	var activeFeedbacks []map[string]interface{}
	for _, group := range seen {
		latest := group[0]
		for _, fb := range group {
			if fb["created_at"].(time.Time).After(latest["created_at"].(time.Time)) {
				latest = fb
			}
		}
		activeFeedbacks = append(activeFeedbacks, latest)
	}

	return h.sendFeedbackList(callback.Message.Chat.ID, activeFeedbacks)
}

// handleShowArchiveFeedbacks показывает только обработанные отзывы
func (h *TelegramHandler) handleShowArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		return h.sendMessage(callback.Message.Chat.ID, "📝 Отзывов пока нет")
	}

	// Группируем и фильтруем только обработанные отзывы
	type feedbackKey struct {
		userID       int64
		feedbackText string
	}
	seen := make(map[feedbackKey][]map[string]interface{})
	for _, fb := range feedbacks {
		if fb["is_processed"].(bool) { // Только обработанные
			key := feedbackKey{
				userID:       fb["telegram_id"].(int64),
				feedbackText: fb["feedback_text"].(string),
			}
			seen[key] = append(seen[key], fb)
		}
	}

	if len(seen) == 0 {
		return h.sendMessage(callback.Message.Chat.ID, "📚 Архив пуст - нет обработанных отзывов")
	}

	// Отправляем заголовок
	headerMsg := tgbotapi.NewMessage(callback.Message.Chat.ID,
		fmt.Sprintf("📚 <b>Архив обработанных отзывов (%d):</b>", len(seen)))
	headerMsg.ParseMode = tgbotapi.ModeHTML
	if _, err := h.bot.Send(headerMsg); err != nil {
		log.Printf("Ошибка отправки заголовка архива отзывов: %v", err)
	}

	var archivedFeedbacks []map[string]interface{}
	for _, group := range seen {
		latest := group[0]
		for _, fb := range group {
			if fb["created_at"].(time.Time).After(latest["created_at"].(time.Time)) {
				latest = fb
			}
		}
		archivedFeedbacks = append(archivedFeedbacks, latest)
	}

	return h.sendFeedbackList(callback.Message.Chat.ID, archivedFeedbacks)
}

// handleShowAllFeedbacks показывает все отзывы
func (h *TelegramHandler) handleShowAllFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Получаем все отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		return h.sendMessage(callback.Message.Chat.ID, "📝 Отзывов пока нет")
	}

	// Группируем все отзывы
	type feedbackKey struct {
		userID       int64
		feedbackText string
	}
	seen := make(map[feedbackKey][]map[string]interface{})
	for _, fb := range feedbacks {
		key := feedbackKey{
			userID:       fb["telegram_id"].(int64),
			feedbackText: fb["feedback_text"].(string),
		}
		seen[key] = append(seen[key], fb)
	}

	// Отправляем заголовок
	totalCount := len(seen)
	headerMsg := tgbotapi.NewMessage(callback.Message.Chat.ID,
		fmt.Sprintf("📋 <b>Все отзывы (%d):</b>", totalCount))
	headerMsg.ParseMode = tgbotapi.ModeHTML
	if _, err := h.bot.Send(headerMsg); err != nil {
		log.Printf("Ошибка отправки заголовка всех отзывов: %v", err)
	}

	var allFeedbacks []map[string]interface{}
	for _, group := range seen {
		latest := group[0]
		for _, fb := range group {
			if fb["created_at"].(time.Time).After(latest["created_at"].(time.Time)) {
				latest = fb
			}
		}
		allFeedbacks = append(allFeedbacks, latest)
	}

	return h.sendFeedbackList(callback.Message.Chat.ID, allFeedbacks)
}

// вспомогательная функция для сохранения состояния навигации отзывов
func (h *TelegramHandler) getFeedbackNavigationState(userID int64, feedbackType string, currentIndex int) string {
	return fmt.Sprintf("fb_nav_%d_%s_%d", userID, feedbackType, currentIndex)
}

// вспомогательная функция для извлечения состояния навигации
func (h *TelegramHandler) parseFeedbackNavigationState(stateStr string) (userID int64, feedbackType string, currentIndex int) {
	parts := strings.Split(stateStr, "_")
	if len(parts) >= 4 && parts[0] == "fb" && parts[1] == "nav" {
		userID, _ = strconv.ParseInt(parts[2], 10, 64)
		feedbackType = parts[3]
		if len(parts) >= 5 {
			currentIndex, _ = strconv.Atoi(parts[4])
		}
	}
	return
}

// Вспомогательная функция для расчета процентов
func (h *TelegramHandler) calculatePercentage(part, total int) int {
	if total == 0 {
		return 0
	}
	return (part * 100) / total
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

// === ОБРАБОТЧИКИ СИСТЕМЫ ОБРАТНОЙ СВЯЗИ ===

// handleFeedbackMessage обрабатывает отзыв пользователя
func (h *TelegramHandler) handleFeedbackMessage(message *tgbotapi.Message, user *models.User) error {
	feedbackText := message.Text

	// Проверяем валидность отзыва
	if len([]rune(feedbackText)) < 10 {
		return h.handleFeedbackTooShort(message, user)
	}
	if len([]rune(feedbackText)) > 1000 {
		return h.handleFeedbackTooLong(message, user)
	}

	// Проверяем наличие username
	if user.Username == "" {
		return h.handleFeedbackContactRequest(message, user, feedbackText)
	}

	// Логируем принятие отзыва
	log.Printf("Отзыв принят: len=%d, has_username=%v", len([]rune(feedbackText)), user.Username != "")

	// Сохраняем полный отзыв и отправляем уведомление
	return h.handleFeedbackComplete(message, user, feedbackText, nil)
}

// handleFeedbackTooShort обрабатывает слишком короткий отзыв
func (h *TelegramHandler) handleFeedbackTooShort(message *tgbotapi.Message, user *models.User) error {
	feedbackText := message.Text
	count := len([]rune(feedbackText))

	errorText := fmt.Sprintf("%s\n\n%s",
		h.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_too_short"),
		h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "feedback_char_count", map[string]string{
			"count": strconv.Itoa(count),
		}),
	)

	return h.sendMessage(message.Chat.ID, errorText)
}

// handleFeedbackTooLong обрабатывает слишком длинный отзыв
func (h *TelegramHandler) handleFeedbackTooLong(message *tgbotapi.Message, user *models.User) error {
	feedbackText := message.Text
	count := len([]rune(feedbackText))

	errorText := fmt.Sprintf("%s\n\n%s",
		h.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_too_long"),
		h.service.Localizer.GetWithParams(user.InterfaceLanguageCode, "feedback_char_count", map[string]string{
			"count": strconv.Itoa(count),
		}),
	)

	return h.sendMessage(message.Chat.ID, errorText)
}

// handleFeedbackContactRequest запрашивает контактные данные при отсутствии username
func (h *TelegramHandler) handleFeedbackContactRequest(message *tgbotapi.Message, user *models.User, feedbackText string) error {
	// Сохраняем отзыв во временном хранилище (в будущем можно добавить в redis/кэш)
	// Пока просто переходим к следующему состоянию

	// Обновляем состояние для ожидания контактных данных
	err := h.service.DB.UpdateUserState(user.ID, models.StateWaitingFeedbackContact)
	if err != nil {
		return err
	}

	// Запрашиваем контактные данные
	contactText := h.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_request")
	return h.sendMessage(message.Chat.ID, contactText)
}

// handleFeedbackContactMessage обрабатывает контактные данные пользователя
func (h *TelegramHandler) handleFeedbackContactMessage(message *tgbotapi.Message, user *models.User) error {
	contactInfo := strings.TrimSpace(message.Text)

	// Валидируем контактные данные
	if contactInfo == "" {
		return h.sendMessage(message.Chat.ID,
			h.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_placeholder"))
	}

	// Подтверждаем получение контактов
	confirmedText := h.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_contact_provided")
	h.sendMessage(message.Chat.ID, confirmedText)

	// Теперь нужно получить сохраненный отзыв пользователя
	// Пока что используем временное решение - просим написать отзыв заново
	// В будущем здесь будет получение из кэша

	feedbackText := "Отзыв был сохранен в предыдущем шаге (требуется интеграция с кэшем)" // временное решение

	return h.handleFeedbackComplete(message, user, feedbackText, &contactInfo)
}

// handleFeedbackComplete завершает процесс обратной связи
func (h *TelegramHandler) handleFeedbackComplete(message *tgbotapi.Message, user *models.User, feedbackText string, contactInfo *string) error {
	// Используем ID администраторов из обработчика
	adminIDs := h.adminChatIDs

	// Сохраняем отзыв через сервис
	err := h.service.SaveUserFeedback(user.ID, feedbackText, contactInfo, adminIDs)
	if err != nil {
		log.Printf("Ошибка сохранения отзыва: %v", err)
		// Используем локализацию для ошибки
		errorText := h.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_error_generic")
		if errorText == "feedback_error_generic" { // fallback в случае отсутствия перевода
			errorText = "Произошла ошибка при сохранении отзыва. Попробуйте позже."
		}
		return h.sendMessage(message.Chat.ID, errorText)
	}

	// Отправляем подтверждение пользователю
	successText := h.service.Localizer.Get(user.InterfaceLanguageCode, "feedback_saved")
	h.sendMessage(message.Chat.ID, successText)

	// Обновляем состояние пользователя на активное
	return h.service.DB.UpdateUserState(user.ID, models.StateActive)
}

// === ОБРАБОТЧИКИ КОНТРОЛЯ ОТЗЫВОВ ===

// handleFeedbackProcess помечает отзыв как обработанный
func (h *TelegramHandler) handleFeedbackProcess(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error {
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка идентификатора отзыва")
	}

	// Обновляем статус отзыва как обработанный
	err = h.service.UpdateFeedbackStatus(feedbackID, true)
	if err != nil {
		log.Printf("Ошибка обновления статуса отзыва %d: %v", feedbackID, err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка обновления статуса")
	}

	// Отправляем обновление администратору
	confirmMsg := fmt.Sprintf("✅ Отзыв #%d отмечен как <b>обработанный</b>", feedbackID)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, confirmMsg)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки подтверждения обработки: %v", err)
	}

	return nil
}

// handleFeedbackUnprocess возвращает отзыв в необработанный статус
func (h *TelegramHandler) handleFeedbackUnprocess(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error {
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка идентификатора отзыва")
	}

	// Возвращаем отзыв в необработанный статус
	err = h.service.UpdateFeedbackStatus(feedbackID, false)
	if err != nil {
		log.Printf("Ошибка возврата отзыва в обработку %d: %v", feedbackID, err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка возврата статуса")
	}

	// Отправляем обновление администратору
	confirmMsg := fmt.Sprintf("🔄 Отзыв #%d возвращен в <b>обработку</b>", feedbackID)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, confirmMsg)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки подтверждения возврата: %v", err)
	}

	return nil
}

// handleFeedbackDelete удаляет отзыв из базы данных
func (h *TelegramHandler) handleFeedbackDelete(callback *tgbotapi.CallbackQuery, user *models.User, feedbackIDStr string) error {
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка идентификатора отзыва")
	}

	// Удаляем отзыв
	err = h.service.DeleteFeedback(feedbackID)
	if err != nil {
		log.Printf("Ошибка удаления отзыва %d: %v", feedbackID, err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка удаления отзыва")
	}

	// Отправляем подтверждение удаления
	deleteMsg := fmt.Sprintf("🗑️ Отзыв #%d <b>удален</b>", feedbackID)
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, deleteMsg)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки подтверждения удаления: %v", err)
	}

	return nil
}

// showFeedbackStatisticsEdit показывает статистику отзывов с редактированием текущего сообщения
func (h *TelegramHandler) showFeedbackStatisticsEdit(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Проверяем права администратора
	isAdminByID := false
	isAdminByUsername := false

	for _, adminID := range h.adminChatIDs {
		if callback.Message.Chat.ID == adminID {
			isAdminByID = true
			break
		}
	}

	if user.Username != "" {
		for _, adminUsername := range h.adminUsernames {
			cleanUsername := strings.TrimPrefix(adminUsername, "@")
			if user.Username == cleanUsername {
				isAdminByUsername = true
				break
			}
		}
	}

	if !isAdminByID && !isAdminByUsername {
		return h.sendMessage(callback.Message.Chat.ID, "❌ Данная команда доступна только администраторам бота.")
	}

	// Получаем все отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		editMsg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"📝 Отзывов пока нет",
		)
		_, err := h.bot.Request(editMsg)
		return err
	}

	// Статистика та же, что и в handleFeedbacksCommand
	type feedbackKey struct {
		userID       int64
		feedbackText string
	}
	seen := make(map[feedbackKey][]map[string]interface{})
	for _, fb := range feedbacks {
		key := feedbackKey{
			userID:       fb["telegram_id"].(int64),
			feedbackText: fb["feedback_text"].(string),
		}
		seen[key] = append(seen[key], fb)
	}

	var processedFeedbacks []map[string]interface{}
	var unprocessedFeedbacks []map[string]interface{}

	totalFeedbacks := len(seen)
	processedCount := 0
	unprocessedCount := 0
	shortCount := 0
	longCount := 0
	contactCount := 0

	for _, group := range seen {
		latest := group[0]
		for _, fb := range group {
			if fb["created_at"].(time.Time).After(latest["created_at"].(time.Time)) {
				latest = fb
			}
		}

		charCount := len([]rune(strings.ReplaceAll(latest["feedback_text"].(string), "\n", " ")))

		if charCount < 50 {
			shortCount++
		} else if charCount > 200 {
			longCount++
		}

		if latest["is_processed"].(bool) {
			processedCount++
			processedFeedbacks = append(processedFeedbacks, latest)
		} else {
			unprocessedCount++
			unprocessedFeedbacks = append(unprocessedFeedbacks, latest)
		}

		if latest["contact_info"] != nil && latest["contact_info"].(string) != "" {
			contactCount++
		}
	}

	mediumCount := totalFeedbacks - shortCount - longCount

	statsMessage := fmt.Sprintf(
		"📊 Отзывы - Статистика:\n\n"+
			"⏳ Обработка:\n"+
			"- Всего отзывов: %d\n"+
			"- 🆕 Активных (необработанных): %d\n"+
			"- ✅ В архиве (обработанных): %d\n\n"+
			"📏 По длине:\n"+
			"- 📝 Короткие (< 50 симв.): %d\n"+
			"- 📊 Средние (50-200 симв.): %d\n"+
			"- 📖 Длинные (> 200 симв.): %d\n\n"+
			"📞 С контактными данными: %d",
		totalFeedbacks, unprocessedCount, processedCount,
		shortCount, mediumCount, longCount, contactCount,
	)

	// Кнопки управления отзывами
	var buttons [][]tgbotapi.InlineKeyboardButton
	if unprocessedCount > 0 {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("🆕 Просмотреть активные "+fmt.Sprintf("(%d)", unprocessedCount), "browse_active_feedbacks_0"),
		})
	}
	if processedCount > 0 {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("📚 Просмотреть архив "+fmt.Sprintf("(%d)", processedCount), "browse_archive_feedbacks_0"),
		})
	}
	if len(seen) > 0 {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("📋 Просмотреть все "+fmt.Sprintf("(%d)", totalFeedbacks), "browse_all_feedbacks_0"),
		})
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	// Редактируем текушее сообщение
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		statsMessage,
		keyboard,
	)

	_, err = h.bot.Request(editMsg)
	return err
}

// === ОБРАБОТЧИКИ ИНТЕРАКТИВНОГО ПРОСМОТРА ОТЗЫВОВ ===

// handleBrowseActiveFeedbacks показывает активные отзывы в интерактивном режиме с редактированием
func (h *TelegramHandler) handleBrowseActiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		log.Printf("Ошибка парсинга индекса: %v", err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка индекса")
	}

	// Получаем активные отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	// Фильтруем только необработанные отзывы
	var activeFeedbacks []map[string]interface{}
	type feedbackKey struct {
		userID       int64
		feedbackText string
	}
	seen := make(map[feedbackKey][]map[string]interface{})
	for _, fb := range feedbacks {
		if !fb["is_processed"].(bool) {
			key := feedbackKey{
				userID:       fb["telegram_id"].(int64),
				feedbackText: fb["feedback_text"].(string),
			}
			seen[key] = append(seen[key], fb)
		}
	}

	for _, group := range seen {
		for _, fb := range group {
			activeFeedbacks = append(activeFeedbacks, fb)
			break
		}
	}

	if len(activeFeedbacks) == 0 {
		return h.sendMessage(callback.Message.Chat.ID, "🎉 Все отзывы обработаны!")
	}

	// Проверяем границы
	if index < 0 || index >= len(activeFeedbacks) {
		index = 0
	}

	// Показываем текущий отзыв с редактированием текущего сообщения
	return h.showFeedbackItemWithNavigationEdit(callback, activeFeedbacks[index], index, len(activeFeedbacks), "active")
}

// handleBrowseArchiveFeedbacks показывает обработанные отзывы в интерактивном режиме
func (h *TelegramHandler) handleBrowseArchiveFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		log.Printf("Ошибка парсинга индекса: %v", err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка индекса")
	}

	// Получаем обработанные отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	// Фильтруем только обработанные отзывы
	var archivedFeedbacks []map[string]interface{}
	type feedbackKey struct {
		userID       int64
		feedbackText string
	}
	seen := make(map[feedbackKey][]map[string]interface{})
	for _, fb := range feedbacks {
		if fb["is_processed"].(bool) {
			key := feedbackKey{
				userID:       fb["telegram_id"].(int64),
				feedbackText: fb["feedback_text"].(string),
			}
			seen[key] = append(seen[key], fb)
		}
	}

	for _, group := range seen {
		for _, fb := range group {
			archivedFeedbacks = append(archivedFeedbacks, fb)
			break
		}
	}

	if len(archivedFeedbacks) == 0 {
		return h.sendMessage(callback.Message.Chat.ID, "📚 Архив пуст - нет обработанных отзывов")
	}

	// Проверяем границы
	if index < 0 || index >= len(archivedFeedbacks) {
		index = 0
	}

	// Показываем текущий отзыв с редактированием текущего сообщения
	return h.showFeedbackItemWithNavigationEdit(callback, archivedFeedbacks[index], index, len(archivedFeedbacks), "archive")
}

// handleBrowseAllFeedbacks показывает все отзывы в интерактивном режиме
func (h *TelegramHandler) handleBrowseAllFeedbacks(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string) error {
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		log.Printf("Ошибка парсинга индекса: %v", err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка индекса")
	}

	// Получаем все отзывы
	feedbacks, err := h.service.GetAllFeedback()
	if err != nil {
		log.Printf("Ошибка получения отзывов: %v", err)
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка получения отзывов")
	}

	if len(feedbacks) == 0 {
		return h.sendMessage(callback.Message.Chat.ID, "📝 Отзывов пока нет")
	}

	// Проверяем границы
	if index < 0 || index >= len(feedbacks) {
		index = 0
	}

	// Показываем текущий отзыв с редактированием текущего сообщения
	return h.showFeedbackItemWithNavigationEdit(callback, feedbacks[index], index, len(feedbacks), "all")
}

// showFeedbackItemWithNavigation показывает отзыв с кнопками навигации
func (h *TelegramHandler) showFeedbackItemWithNavigation(chatID int64, fb map[string]interface{}, currentIndex int, totalCount int, feedbackType string) error {
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

	// Определим тип списка для заголовка
	headerText := ""
	switch feedbackType {
	case "active":
		headerText = fmt.Sprintf("🆕 <b>Активные отзывы (%d/%d)</b>", currentIndex+1, totalCount)
	case "archive":
		headerText = fmt.Sprintf("📚 <b>Архив (%d/%d)</b>", currentIndex+1, totalCount)
	case "all":
		headerText = fmt.Sprintf("📋 <b>Все отзывы (%d/%d)</b>", currentIndex+1, totalCount)
	}

	// Формируем полное объединенное сообщение
	fullMessage := fmt.Sprintf("%s\n\n%s <b>%s</b> %s\n"+
		"👤 <b>Автор:</b> %s\n"+
		"📊 <b>Статус:</b> %s (%d символов)\n"+
		"⏰ <b>Дата:</b> %s%s\n\n"+
		"<b>📨 Содержание отзыва:</b>\n"+
		"<i>%s</i>",
		headerText, statusIcon, firstName, username,
		statusText,
		charIcon,
		charCount,
		dateStr,
		contactStr,
		feedbackTextContent,
	)

	// Создаем клавиатуру навигации
	var buttons [][]tgbotapi.InlineKeyboardButton

	// Кнопки управления отзывом
	actionRow := []tgbotapi.InlineKeyboardButton{}
	if fb["is_processed"].(bool) {
		actionRow = append(actionRow,
			tgbotapi.NewInlineKeyboardButtonData("🔄 Вернуть в обработку", fmt.Sprintf("fb_unprocess_%d", feedbackID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("fb_delete_%d", feedbackID)),
		)
	} else {
		actionRow = append(actionRow,
			tgbotapi.NewInlineKeyboardButtonData("✅ Обработан", fmt.Sprintf("fb_process_%d", feedbackID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("fb_delete_%d", feedbackID)),
		)
	}
	buttons = append(buttons, actionRow)

	// Кнопки навигации
	navRow := []tgbotapi.InlineKeyboardButton{}
	if currentIndex > 0 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ Предыдущий", fmt.Sprintf("feedback_prev_%d_%s", currentIndex, feedbackType)))
	}
	navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("🏠 К стат-тике", fmt.Sprintf("feedback_back_%s", feedbackType)))
	if currentIndex < totalCount-1 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("Следующий ➡️", fmt.Sprintf("feedback_next_%d_%s", currentIndex, feedbackType)))
	}
	buttons = append(buttons, navRow)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, fullMessage)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}

// showFeedbackItemWithNavigationEdit показывает отзыв с кнопками навигации с редактированием текущего сообщения
func (h *TelegramHandler) showFeedbackItemWithNavigationEdit(callback *tgbotapi.CallbackQuery, fb map[string]interface{}, currentIndex int, totalCount int, feedbackType string) error {
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

	// Определим тип списка для заголовка
	headerText := ""
	switch feedbackType {
	case "active":
		headerText = fmt.Sprintf("🆕 <b>Активные отзывы (%d/%d)</b>", currentIndex+1, totalCount)
	case "archive":
		headerText = fmt.Sprintf("📚 <b>Архив (%d/%d)</b>", currentIndex+1, totalCount)
	case "all":
		headerText = fmt.Sprintf("📋 <b>Все отзывы (%d/%d)</b>", currentIndex+1, totalCount)
	}

	// Формируем полное объединенное сообщение
	fullMessage := fmt.Sprintf("%s\n\n%s <b>%s</b> %s\n"+
		"👤 <b>Автор:</b> %s\n"+
		"📊 <b>Статус:</b> %s (%d символов)\n"+
		"⏰ <b>Дата:</b> %s%s\n\n"+
		"<b>📨 Содержание отзыва:</b>\n"+
		"<i>%s</i>",
		headerText, statusIcon, firstName, username,
		statusText,
		charIcon,
		charCount,
		dateStr,
		contactStr,
		feedbackTextContent,
	)

	// Создаем клавиатуру навигации
	var buttons [][]tgbotapi.InlineKeyboardButton

	// Кнопки управления отзывом
	actionRow := []tgbotapi.InlineKeyboardButton{}
	if fb["is_processed"].(bool) {
		actionRow = append(actionRow,
			tgbotapi.NewInlineKeyboardButtonData("🔄 Вернуть в обработку", fmt.Sprintf("fb_unprocess_%d", feedbackID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("fb_delete_%d", feedbackID)),
		)
	} else {
		actionRow = append(actionRow,
			tgbotapi.NewInlineKeyboardButtonData("✅ Обработан", fmt.Sprintf("fb_process_%d", feedbackID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("fb_delete_%d", feedbackID)),
		)
	}
	buttons = append(buttons, actionRow)

	// Кнопки навигации
	navRow := []tgbotapi.InlineKeyboardButton{}
	if currentIndex > 0 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ Предыдущий", fmt.Sprintf("feedback_prev_%d_%s", currentIndex, feedbackType)))
	}
	navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("🏠 К стат-тике", fmt.Sprintf("feedback_back_%s", feedbackType)))
	if currentIndex < totalCount-1 {
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("Следующий ➡️", fmt.Sprintf("feedback_next_%d_%s", currentIndex, feedbackType)))
	}
	buttons = append(buttons, navRow)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		fullMessage,
		keyboard,
	)
	editMsg.ParseMode = tgbotapi.ModeHTML

	_, err := h.bot.Request(editMsg)
	return err
}

// handleFeedbackPrev переходит к предыдущему отзыву
func (h *TelegramHandler) handleFeedbackPrev(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error {
	currentIndex, err := strconv.Atoi(indexStr)
	if err != nil {
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка индекса")
	}

	newIndex := currentIndex - 1
	if newIndex < 0 {
		newIndex = 0
	}

	switch feedbackType {
	case "active":
		return h.handleBrowseActiveFeedbacks(callback, user, strconv.Itoa(newIndex))
	case "archive":
		return h.handleBrowseArchiveFeedbacks(callback, user, strconv.Itoa(newIndex))
	case "all":
		return h.handleBrowseAllFeedbacks(callback, user, strconv.Itoa(newIndex))
	default:
		return h.sendMessage(callback.Message.Chat.ID, "❌ Неизвестный тип отзывов")
	}
}

// handleFeedbackNext переходит к следующему отзыву
func (h *TelegramHandler) handleFeedbackNext(callback *tgbotapi.CallbackQuery, user *models.User, indexStr string, feedbackType string) error {
	currentIndex, err := strconv.Atoi(indexStr)
	if err != nil {
		return h.sendMessage(callback.Message.Chat.ID, "❌ Ошибка индекса")
	}

	newIndex := currentIndex + 1

	switch feedbackType {
	case "active":
		return h.handleBrowseActiveFeedbacks(callback, user, strconv.Itoa(newIndex))
	case "archive":
		return h.handleBrowseArchiveFeedbacks(callback, user, strconv.Itoa(newIndex))
	case "all":
		return h.handleBrowseAllFeedbacks(callback, user, strconv.Itoa(newIndex))
	default:
		return h.sendMessage(callback.Message.Chat.ID, "❌ Неизвестный тип отзывов")
	}
}

// handleFeedbackBack возвращает к статистике отзывов с редактированием текущего сообщения
func (h *TelegramHandler) handleFeedbackBack(callback *tgbotapi.CallbackQuery, user *models.User, feedbackType string) error {
	return h.showFeedbackStatisticsEdit(callback, user)
}
