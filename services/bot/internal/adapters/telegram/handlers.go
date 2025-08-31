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
	default:
		return h.sendMessage(message.Chat.ID, h.service.Localizer.Get(user.InterfaceLanguageCode, "unknown_command"))
	}
}

func (h *TelegramHandler) handleStartCommand(message *tgbotapi.Message, user *models.User) error {
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
	callbackResponse := tgbotapi.NewCallback(callback.ID, "")
	_, _ = h.bot.Request(callbackResponse)

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
	default:
		return nil
	}
}

// ✨ Выбор родного языка
func (h *TelegramHandler) handleNativeLanguageCallback(cb *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := strings.TrimPrefix(cb.Data, "lang_native_")

	if err := h.service.DB.UpdateUserNativeLanguage(user.ID, langCode); err != nil {
		return err
	}
	user.NativeLanguageCode = langCode

	// ВАЖНО: интерфейс НЕ меняем автоматически
	langName := h.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)
	confirm := fmt.Sprintf("✅ %s: %s\n\n%s",
		h.service.Localizer.Get(user.InterfaceLanguageCode, "native_language_confirmed"),
		langName,
		h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_target_language"),
	)
	edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, confirm)
	if _, err := h.bot.Request(edit); err != nil {
		return err
	}

	return h.sendTargetLanguageMenu(cb.Message.Chat.ID, user)
}

func (h *TelegramHandler) handleTargetLanguageCallback(callback *tgbotapi.CallbackQuery, user *models.User) error {
	langCode := strings.TrimPrefix(callback.Data, "lang_target_")
	if err := h.service.DB.UpdateUserTargetLanguage(user.ID, langCode); err != nil {
		return err
	}
	user.TargetLanguageCode = langCode

	langName := h.service.Localizer.GetLanguageName(langCode, user.InterfaceLanguageCode)
	confirmMsg := fmt.Sprintf("✅ %s: %s\n\n%s",
		h.service.Localizer.Get(user.InterfaceLanguageCode, "target_language_confirmed"),
		langName,
		h.service.Localizer.Get(user.InterfaceLanguageCode, "choose_interests"),
	)

	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, confirmMsg)
	if _, err := h.bot.Request(editMsg); err != nil {
		return err
	}
	return h.sendInterestsMenu(callback.Message.Chat.ID, user)
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
	if err := h.service.DB.SaveUserInterest(user.ID, interestID, false); err != nil {
		log.Printf("Error saving user interest: %v", err)
		return err
	}
	text := h.service.Localizer.Get(user.InterfaceLanguageCode, "interest_added")
	return h.sendMessage(callback.Message.Chat.ID, text)
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
	keyboard := h.createInterestsKeyboard(interests)
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
