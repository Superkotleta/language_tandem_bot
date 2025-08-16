package handlers

import (
	"fmt"
	"log"
	"strings"

	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramHandler struct {
	bot *tgbotapi.BotAPI
	db  *database.DB
}

func NewTelegramHandler(bot *tgbotapi.BotAPI, db *database.DB) *TelegramHandler {
	return &TelegramHandler{
		bot: bot,
		db:  db,
	}
}

func (h *TelegramHandler) HandleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message != nil {
			h.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			h.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

func (h *TelegramHandler) handleMessage(message *tgbotapi.Message) {
	// Находим или создаем пользователя
	user, err := h.db.FindOrCreateUser(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
	)
	if err != nil {
		log.Printf("Error finding/creating user: %v", err)
		return
	}

	// Обрабатываем команды
	if message.IsCommand() {
		h.handleCommand(message, user)
		return
	}

	// Обрабатываем состояния
	h.handleState(message, user)
}

func (h *TelegramHandler) handleCommand(message *tgbotapi.Message, user *models.User) {
	switch message.Command() {
	case "start":
		h.handleStartCommand(message, user)
	case "status":
		h.handleStatusCommand(message, user)
	case "reset":
		h.handleResetCommand(message, user)
	default:
		h.sendMessage(message.Chat.ID, "Неизвестная команда. Используйте /start")
	}
}

func (h *TelegramHandler) handleStartCommand(message *tgbotapi.Message, user *models.User) {
	welcomeText := fmt.Sprintf(
		"🎉 Привет, %s! Добро пожаловать в Language Exchange Bot!\n\n"+
			"Я помогу найти тебе идеального языкового партнера для практики.\n\n"+
			"Давай заполним твой профиль! 📝\n\n"+
			"Шаг 1: Выбери свой родной язык:",
		user.FirstName,
	)

	// Создаем клавиатуру с языками
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_native_ru"),
			tgbotapi.NewInlineKeyboardButtonData("🇺🇸 English", "lang_native_en"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇪🇸 Español", "lang_native_es"),
			tgbotapi.NewInlineKeyboardButtonData("🇫🇷 Français", "lang_native_fr"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)

	// Обновляем состояние
	h.db.UpdateUserState(user.ID, models.StateWaitingLanguage)
	h.db.UpdateUserStatus(user.ID, models.StatusFilling)
}

func (h *TelegramHandler) handleStatusCommand(message *tgbotapi.Message, user *models.User) {
	statusText := fmt.Sprintf(
		"📊 Твой статус:\n\n"+
			"🆔 ID: %d\n"+
			"📝 Статус: %s\n"+
			"🔄 Состояние: %s\n"+
			"📈 Уровень заполнения профиля: %d%%",
		user.ID,
		h.getStatusEmoji(user.Status),
		h.getStateDescription(user.State),
		user.ProfileCompletionLevel,
	)

	h.sendMessage(message.Chat.ID, statusText)
}

func (h *TelegramHandler) handleResetCommand(message *tgbotapi.Message, user *models.User) {
	h.db.UpdateUserState(user.ID, models.StateStart)
	h.db.UpdateUserStatus(user.ID, models.StatusNotStarted)

	h.sendMessage(message.Chat.ID, "✅ Профиль сброшен! Используйте /start для начала заново.")
}

func (h *TelegramHandler) handleState(message *tgbotapi.Message, user *models.User) {
	switch user.State {
	case models.StateWaitingLanguage:
		h.sendMessage(message.Chat.ID, "👆 Пожалуйста, выберите язык из меню выше")
	case models.StateWaitingInterests:
		h.sendMessage(message.Chat.ID, "👆 Пожалуйста, выберите интересы из меню выше")
	case models.StateWaitingTime:
		h.sendMessage(message.Chat.ID, "👆 Пожалуйста, выберите время из меню выше")
	default:
		h.sendMessage(message.Chat.ID, "Используйте /start для начала работы с ботом")
	}
}

func (h *TelegramHandler) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	// Получаем пользователя
	user, err := h.db.FindOrCreateUser(
		callback.From.ID,
		callback.From.UserName,
		callback.From.FirstName,
	)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		return
	}

	data := callback.Data

	// Обрабатываем выбор языка
	if strings.HasPrefix(data, "lang_native_") {
		lang := strings.TrimPrefix(data, "lang_native_")
		h.handleNativeLanguageSelection(callback, user, lang)
	}
}

func (h *TelegramHandler) handleNativeLanguageSelection(callback *tgbotapi.CallbackQuery, user *models.User, lang string) {
	langNames := map[string]string{
		"ru": "Русский 🇷🇺",
		"en": "English 🇺🇸",
		"es": "Español 🇪🇸",
		"fr": "Français 🇫🇷",
	}

	langName := langNames[lang]

	// Отвечаем на callback
	callbackResponse := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("Выбран родной язык: %s", langName))
	h.bot.Request(callbackResponse)

	// Отправляем следующий вопрос
	text := fmt.Sprintf("✅ Родной язык: %s\n\nТеперь выбери язык, который хочешь изучать:", langName)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang_target_ru"),
			tgbotapi.NewInlineKeyboardButtonData("🇺🇸 English", "lang_target_en"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇪🇸 Español", "lang_target_es"),
			tgbotapi.NewInlineKeyboardButtonData("🇫🇷 Français", "lang_target_fr"),
		),
	)

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)

	// TODO: Сохранить выбранный язык в БД
	// h.db.SaveNativeLanguage(user.ID, lang)
}

// Вспомогательные функции
func (h *TelegramHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	h.bot.Send(msg)
}

func (h *TelegramHandler) getStatusEmoji(status string) string {
	switch status {
	case models.StatusNotStarted:
		return "🔴 Не начат"
	case models.StatusFilling:
		return "🟡 Заполняется"
	case models.StatusReady:
		return "🟢 Готов к подбору"
	case models.StatusMatched:
		return "💙 Найден партнер"
	case models.StatusWaiting:
		return "⏳ В ожидании"
	default:
		return status
	}
}

func (h *TelegramHandler) getStateDescription(state string) string {
	switch state {
	case models.StateStart:
		return "Начальное"
	case models.StateWaitingLanguage:
		return "Выбор языка"
	case models.StateWaitingInterests:
		return "Выбор интересов"
	case models.StateWaitingTime:
		return "Выбор времени"
	case models.StateComplete:
		return "Завершено"
	default:
		return "Неизвестно"
	}
}
