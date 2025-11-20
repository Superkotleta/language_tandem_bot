package mocks

import (
	"fmt"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramHandlerWrapper обертка для тестирования обработчика Telegram.
type TelegramHandlerWrapper struct {
	Service        *core.BotService
	SentMessages   []tgbotapi.MessageConfig
	SentCallbacks  []tgbotapi.CallbackConfig
	EditedMessages []tgbotapi.EditMessageTextConfig
	LastError      error
}

// HandleUpdate обрабатывает update и записывает отправленные сообщения.
func (w *TelegramHandlerWrapper) HandleUpdate(update tgbotapi.Update) error {
	if update.Message != nil {
		return w.handleMessage(update.Message)
	}

	if update.CallbackQuery != nil {
		return w.handleCallbackQuery(update.CallbackQuery)
	}

	return nil
}

// handleMessage имитирует обработку сообщения.
func (w *TelegramHandlerWrapper) handleMessage(message *tgbotapi.Message) error {
	user, err := w.Service.HandleUserRegistration(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
		message.From.LanguageCode,
	)
	if err != nil {
		w.LastError = err

		return err
	}

	if message.IsCommand() {
		return w.handleCommand(message, user)
	}

	return w.handleState(message, user)
}

// handleCommand имитирует обработку команд.
func (w *TelegramHandlerWrapper) handleCommand(message *tgbotapi.Message, user *models.User) error {
	switch message.Command() {
	case "start":
		return w.handleStartCommand(message, user)
	case "status":
		return w.handleStatusCommand(message, user)
	case "profile":
		return w.handleProfileCommand(message, user)
	case "feedback":
		return w.handleFeedbackCommand(message, user)
	default:
		// Неизвестная команда
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command")
		w.SentMessages = append(w.SentMessages, msg)

		return nil
	}
}

// handleStartCommand имитирует обработку команды /start.
func (w *TelegramHandlerWrapper) handleStartCommand(message *tgbotapi.Message, user *models.User) error {
	welcomeText := w.Service.GetWelcomeMessage(user)

	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	// Создаем простую клавиатуру для теста
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 My Profile", "profile_show"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Edit Profile", "profile_edit"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌐 Interface Language", "language_interface"),
			tgbotapi.NewInlineKeyboardButtonData("💬 Feedback", "feedback_create"),
		),
	)
	msg.ReplyMarkup = keyboard

	w.SentMessages = append(w.SentMessages, msg)

	return nil
}

// handleStatusCommand имитирует обработку команды /status.
func (w *TelegramHandlerWrapper) handleStatusCommand(message *tgbotapi.Message, user *models.User) error {
	// Создаем простое сообщение со статусом пользователя
	statusText := fmt.Sprintf("User ID: %d\nStatus: %s\nProfile completion: %d%%",
		user.TelegramID, user.Status, user.ProfileCompletionLevel)

	msg := tgbotapi.NewMessage(message.Chat.ID, statusText)
	w.SentMessages = append(w.SentMessages, msg)

	return nil
}

// handleProfileCommand имитирует обработку команды /profile.
func (w *TelegramHandlerWrapper) handleProfileCommand(message *tgbotapi.Message, user *models.User) error {
	profileText, err := w.Service.BuildProfileSummary(user)
	if err != nil {
		profileText = "Error loading profile"
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, profileText)
	w.SentMessages = append(w.SentMessages, msg)

	return nil
}

// handleFeedbackCommand имитирует обработку команды /feedback.
func (w *TelegramHandlerWrapper) handleFeedbackCommand(message *tgbotapi.Message, user *models.User) error {
	// Проверяем, является ли пользователь администратором
	isAdmin := false

	for _, adminID := range []int64{123456789, 987654321} {
		if user.TelegramID == adminID {
			isAdmin = true

			break
		}
	}

	var msg tgbotapi.MessageConfig
	if isAdmin {
		msg = tgbotapi.NewMessage(message.Chat.ID, "Admin feedback interface")
	} else {
		msg = tgbotapi.NewMessage(message.Chat.ID, "Please send your feedback:")
	}

	w.SentMessages = append(w.SentMessages, msg)

	return nil
}

// handleState имитирует обработку состояний пользователя.
func (w *TelegramHandlerWrapper) handleState(message *tgbotapi.Message, _ *models.User) error {
	// Простая имитация обработки состояний
	msg := tgbotapi.NewMessage(message.Chat.ID, "Processing your message...")
	w.SentMessages = append(w.SentMessages, msg)

	return nil
}

// handleCallbackQuery имитирует обработку callback запросов.
func (w *TelegramHandlerWrapper) handleCallbackQuery(callback *tgbotapi.CallbackQuery) error {
	user, err := w.Service.HandleUserRegistration(
		callback.From.ID,
		callback.From.UserName,
		callback.From.FirstName,
		callback.From.LanguageCode,
	)
	if err != nil {
		w.LastError = err

		return err
	}

	// Имитируем обработку разных callback'ов
	switch callback.Data {
	case "profile_show":
		profileText, err := w.Service.BuildProfileSummary(user)
		if err != nil {
			profileText = "Error loading profile"
		}

		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, profileText)
		w.EditedMessages = append(w.EditedMessages, edit)

	case "profile_edit":
		editText := "Choose what to edit:"
		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, editText)
		w.EditedMessages = append(w.EditedMessages, edit)

	default:
		// Неизвестный callback
		callbackResponse := tgbotapi.NewCallback(callback.ID, "Unknown action")
		w.SentCallbacks = append(w.SentCallbacks, callbackResponse)
	}

	// Отправляем ответ на callback
	callbackResponse := tgbotapi.NewCallback(callback.ID, "")
	w.SentCallbacks = append(w.SentCallbacks, callbackResponse)

	return nil
}

// GetSentMessagesCount возвращает количество отправленных сообщений.
func (w *TelegramHandlerWrapper) GetSentMessagesCount() int {
	return len(w.SentMessages)
}

// GetLastSentMessage возвращает последнее отправленное сообщение.
func (w *TelegramHandlerWrapper) GetLastSentMessage() *tgbotapi.MessageConfig {
	if len(w.SentMessages) == 0 {
		return nil
	}

	return &w.SentMessages[len(w.SentMessages)-1]
}

// Reset очищает все записанные сообщения.
func (w *TelegramHandlerWrapper) Reset() {
	w.SentMessages = make([]tgbotapi.MessageConfig, 0)
	w.SentCallbacks = make([]tgbotapi.CallbackConfig, 0)
	w.EditedMessages = make([]tgbotapi.EditMessageTextConfig, 0)
	w.LastError = nil
}
