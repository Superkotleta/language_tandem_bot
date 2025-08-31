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
	bot     *tgbotapi.BotAPI
	service *core.BotService
}

func NewTelegramHandler(bot *tgbotapi.BotAPI, service *core.BotService) *TelegramHandler {
	return &TelegramHandler{
		bot:     bot,
		service: service,
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
	default:
		return h.sendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
	}
}

func (h *TelegramHandler) handleStartCommand(message *tgbotapi.Message, user *models.User) error {

	completed, err := h.service.IsProfileCompleted(user)
	if err == nil && completed {
		summary, serr := h.service.BuildProfileSummary(user)
		if serr != nil {
			log.Printf("profile summary error: %v", serr)
		}
		text := summary + "\n\n" + h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_actions")
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ReplyMarkup = h.createProfileMenuKeyboard(user.InterfaceLanguageCode)
		_, sendErr := h.bot.Send(msg)
		return sendErr
	}

	welcomeText := h.service.GetWelcomeMessage(user)
	languagePrompt := h.service.GetLanguagePrompt(user, "native")
	fullText := welcomeText + "\n\n" + languagePrompt

	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "native")
	msg := tgbotapi.NewMessage(message.Chat.ID, fullText)
	msg.ReplyMarkup = keyboard
	if _, err := h.bot.Send(msg); err != nil {
		return err
	}

	_ = h.service.DB.UpdateUserState(user.ID, models.StateWaitingLanguage)
	_ = h.service.DB.UpdateUserStatus(user.ID, models.StatusFilling)
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
	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "interface")
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
	case strings.HasPrefix(data, "lang_interface_"):
		langCode := strings.TrimPrefix(data, "lang_interface_")
		return h.handleInterfaceLanguageSelection(callback, user, langCode)
	case strings.HasPrefix(data, "interest_"):
		interestID := strings.TrimPrefix(data, "interest_")
		return h.handleInterestSelection(callback, user, interestID)
	case data == "profile_show":
		return h.handleProfileShow(callback, user)
	case data == "profile_reset_ask":
		return h.handleProfileResetAsk(callback, user)
	case data == "profile_reset_yes":
		return h.handleProfileResetYes(callback, user)
	case data == "profile_reset_no":
		return h.handleProfileResetNo(callback, user)
	default:
		return nil
	}
}

func (h *TelegramHandler) handleInterestsContinue(callback *tgbotapi.CallbackQuery, user *models.User) error {
	// Завершаем онбординг
	completedMsg := h.service.Localizer.Get(user.InterfaceLanguageCode, "profile_completed")

	editMsg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		completedMsg,
	)
	_, err := h.bot.Request(editMsg)
	if err != nil {
		return err
	}

	// Обновляем статус пользователя
	h.service.DB.UpdateUserState(user.ID, models.StateActive)
	h.service.DB.UpdateUserStatus(user.ID, models.StatusActive)

	return nil
}

// ✨ Выбор родного языка
func (h *TelegramHandler) handleNativeLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := callback.Data[len("lang_native_"):]

	err := h.service.DB.UpdateUserNativeLanguage(user.ID, langCode)
	if err != nil {
		return err
	}

	user.NativeLanguageCode = langCode
	langName := h.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)

	// ✅ Одно сообщение с новой клавиатурой
	confirmMsg := fmt.Sprintf("✅ %s: %s\n\n%s",
		h.service.Localizer.Get(user.InterfaceLanguageCode, "native_language_confirmed"),
		langName,
		h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language"))

	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "target")
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(callback.Message.Chat.ID, callback.Message.MessageID, confirmMsg, keyboard)
	_, err = h.bot.Request(editMsg)
	return err
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

	// ✅ Одно сообщение с клавиатурой интересов - НЕТ ДУБЛИРОВАНИЯ
	confirmMsg := fmt.Sprintf("✅ %s: %s\n\n%s",
		h.service.Localizer.Get(user.InterfaceLanguageCode, "target_language_confirmed"),
		langName,
		h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests"))

	// Получаем интересы и создаем клавиатуру с пустым списком выбранных
	interests, _ := h.service.Localizer.GetInterests(user.InterfaceLanguageCode)
	keyboard := h.createInterestsKeyboard(interests, []int{}) // ✅ Пустой массив

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

func (h *TelegramHandler) handleInterfaceLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, langCode string) error {
	if err := h.service.DB.UpdateUserInterfaceLanguage(user.ID, langCode); err != nil {
		log.Printf("Error updating interface language: %v", err)
		return err
	}
	langName := h.service.Localizer.GetLanguageName(langCode, langCode)
	text := fmt.Sprintf("✅ %s: %s",
		h.service.Localizer.Get(langCode, "language_updated"),
		langName,
	)
	return h.sendMessage(callback.Message.Chat.ID, text)
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
	keyboard := h.createInterestsKeyboard(interests, selectedInterests)

	editMsg := tgbotapi.NewEditMessageReplyMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		keyboard,
	)
	_, err = h.bot.Request(editMsg)
	return err
}

func (h *TelegramHandler) sendTargetLanguageMenu(chatID int64, user *models.User) error {
	keyboard := h.createLanguageKeyboard(user.InterfaceLanguageCode, "target")
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language")
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err := h.bot.Send(msg)
	return err
}

func (h *TelegramHandler) sendInterestsMenu(chatID int64, user *models.User) error {
	interests, err := h.service.Localizer.GetInterests(user.InterfaceLanguageCode)
	if err != nil {
		return err
	}

	// ✅ Загружаем уже выбранные интересы пользователя
	selectedInterests, err := h.service.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		log.Printf("Error loading user interests: %v", err)
		selectedInterests = []int{} // fallback на пустой список
	}

	keyboard := h.createInterestsKeyboard(interests, selectedInterests)
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
	// Лейблы можно локализовать через Localizer при желании
	show := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "profile_show"),
		"profile_show",
	)
	reconfig := tgbotapi.NewInlineKeyboardButtonData(
		h.service.Localizer.Get(interfaceLang, "profile_reconfigure"),
		"profile_reset_ask",
	)
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{show},
		{reconfig},
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
		h.createLanguageKeyboard(user.InterfaceLanguageCode, "native"),
	)
	_, err := h.bot.Request(edit)
	return err
}

// Отмена сброса — вернёмся в меню профиля
func (h *TelegramHandler) handleProfileResetNo(callback *tgbotapi.CallbackQuery, user *models.User) error {
	return h.handleProfileShow(callback, user)
}
