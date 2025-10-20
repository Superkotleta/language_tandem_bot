package telegram

import (
	"fmt"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/logging"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Parse modes are now defined in localization/constants.go

// MessageFactory предоставляет гибридный API для отправки Telegram сообщений.
// Поддерживает как простые методы для частых случаев, так и Builder паттерн для сложных сценариев.
type MessageFactory struct {
	bot          *tgbotapi.BotAPI
	errorHandler *errors.ErrorHandler
	logger       *logging.LoggingService
}

// MessageBuilder предоставляет Builder API для создания новых сообщений.
type MessageBuilder struct {
	factory   *MessageFactory
	config    tgbotapi.MessageConfig
	chatID    int64
	userID    int64
	operation string
}

// EditMessageBuilder предоставляет Builder API для редактирования сообщений.
type EditMessageBuilder struct {
	factory   *MessageFactory
	config    tgbotapi.EditMessageTextConfig
	chatID    int64
	messageID int
	userID    int64
	operation string
}

// NewMessageFactory создает новый экземпляр MessageFactory.
func NewMessageFactory(
	bot *tgbotapi.BotAPI,
	errorHandler *errors.ErrorHandler,
	logger *logging.LoggingService,
) *MessageFactory {
	return &MessageFactory{
		bot:          bot,
		errorHandler: errorHandler,
		logger:       logger,
	}
}

// =============================================================================
// ПРОСТЫЕ МЕТОДЫ (Quick API) - для 80% случаев использования
// =============================================================================

// SendText отправляет простое текстовое сообщение.
func (f *MessageFactory) SendText(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)

	return f.sendWithLogging(msg, chatID, 0, "SendText", "text")
}

// SendWithKeyboard отправляет сообщение с клавиатурой.
func (f *MessageFactory) SendWithKeyboard(chatID int64, text string, keyboard interface{}) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	return f.sendWithLogging(msg, chatID, 0, "SendWithKeyboard", "text_with_keyboard")
}

// SendHTML отправляет HTML-форматированное сообщение.
func (f *MessageFactory) SendHTML(chatID int64, htmlText string) error {
	msg := tgbotapi.NewMessage(chatID, htmlText)
	msg.ParseMode = localization.ParseModeHTML

	return f.sendWithLogging(msg, chatID, 0, "SendHTML", "html")
}

// SendHTMLWithKeyboard отправляет HTML-форматированное сообщение с клавиатурой.
func (f *MessageFactory) SendHTMLWithKeyboard(chatID int64, htmlText string, keyboard interface{}) error {
	msg := tgbotapi.NewMessage(chatID, htmlText)
	msg.ParseMode = localization.ParseModeHTML
	msg.ReplyMarkup = keyboard

	return f.sendWithLogging(msg, chatID, 0, "SendHTMLWithKeyboard", "html_with_keyboard")
}

// EditText редактирует текстовое сообщение.
func (f *MessageFactory) EditText(chatID int64, messageID int, text string) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)

	return f.sendWithLogging(edit, chatID, 0, "EditText", "edit_text")
}

// EditWithKeyboard редактирует сообщение с клавиатурой.
func (f *MessageFactory) EditWithKeyboard(chatID int64, messageID int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ReplyMarkup = keyboard

	return f.sendWithLogging(edit, chatID, 0, "EditWithKeyboard", "edit_with_keyboard")
}

// EditHTML редактирует HTML-форматированное сообщение.
func (f *MessageFactory) EditHTML(chatID int64, messageID int, htmlText string) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, htmlText)
	edit.ParseMode = localization.ParseModeHTML

	return f.sendWithLogging(edit, chatID, 0, "EditHTML", "edit_html")
}

// =============================================================================
// BUILDER API - для сложных случаев (20% использования)
// =============================================================================

// NewMessage создает новый MessageBuilder для отправки сообщения.
func (f *MessageFactory) NewMessage(chatID int64) *MessageBuilder {
	return &MessageBuilder{
		factory:   f,
		config:    tgbotapi.MessageConfig{},
		chatID:    chatID,
		userID:    0,
		operation: "NewMessage",
	}
}

// NewEditMessage создает новый EditMessageBuilder для редактирования сообщения.
func (f *MessageFactory) NewEditMessage(chatID int64, messageID int) *EditMessageBuilder {
	return &EditMessageBuilder{
		factory:   f,
		config:    tgbotapi.EditMessageTextConfig{},
		chatID:    chatID,
		messageID: messageID,
		userID:    0,
		operation: "NewEditMessage",
	}
}

// =============================================================================
// MESSAGE BUILDER METHODS
// =============================================================================

// WithText устанавливает текст сообщения.
func (b *MessageBuilder) WithText(text string) *MessageBuilder {
	b.config = tgbotapi.NewMessage(b.chatID, text)

	return b
}

// WithKeyboard устанавливает клавиатуру.
func (b *MessageBuilder) WithKeyboard(keyboard interface{}) *MessageBuilder {
	b.config.ReplyMarkup = keyboard

	return b
}

// WithParseMode устанавливает режим парсинга (HTML, Markdown, MarkdownV2).
func (b *MessageBuilder) WithParseMode(mode string) *MessageBuilder {
	b.config.ParseMode = mode

	return b
}

// WithHTML устанавливает HTML режим парсинга.
func (b *MessageBuilder) WithHTML() *MessageBuilder {
	b.config.ParseMode = localization.ParseModeHTML

	return b
}

// WithMarkdown устанавливает Markdown режим парсинга.
func (b *MessageBuilder) WithMarkdown() *MessageBuilder {
	b.config.ParseMode = "Markdown"

	return b
}

// DisableWebPagePreview отключает превью веб-страниц.
func (b *MessageBuilder) DisableWebPagePreview() *MessageBuilder {
	b.config.DisableWebPagePreview = true

	return b
}

// DisableNotification отправляет сообщение без уведомления.
func (b *MessageBuilder) DisableNotification() *MessageBuilder {
	b.config.DisableNotification = true

	return b
}

// ReplyTo устанавливает сообщение для ответа.
func (b *MessageBuilder) ReplyTo(messageID int) *MessageBuilder {
	b.config.ReplyToMessageID = messageID

	return b
}

// WithUserID устанавливает ID пользователя для логирования.
func (b *MessageBuilder) WithUserID(userID int64) *MessageBuilder {
	b.userID = userID

	return b
}

// WithOperation устанавливает название операции для логирования.
func (b *MessageBuilder) WithOperation(operation string) *MessageBuilder {
	b.operation = operation

	return b
}

// Send отправляет сообщение.
func (b *MessageBuilder) Send() error {
	return b.factory.sendWithLogging(b.config, b.chatID, b.userID, b.operation, "builder_message")
}

// =============================================================================
// EDIT MESSAGE BUILDER METHODS
// =============================================================================

// WithText устанавливает новый текст для редактирования.
func (b *EditMessageBuilder) WithText(text string) *EditMessageBuilder {
	b.config = tgbotapi.NewEditMessageText(b.chatID, b.messageID, text)

	return b
}

// WithKeyboard устанавливает клавиатуру для редактирования.
func (b *EditMessageBuilder) WithKeyboard(keyboard *tgbotapi.InlineKeyboardMarkup) *EditMessageBuilder {
	b.config.ReplyMarkup = keyboard

	return b
}

// WithParseMode устанавливает режим парсинга для редактирования.
func (b *EditMessageBuilder) WithParseMode(mode string) *EditMessageBuilder {
	b.config.ParseMode = mode

	return b
}

// WithHTML устанавливает HTML режим парсинга для редактирования.
func (b *EditMessageBuilder) WithHTML() *EditMessageBuilder {
	b.config.ParseMode = localization.ParseModeHTML

	return b
}

// WithMarkdown устанавливает Markdown режим парсинга для редактирования.
func (b *EditMessageBuilder) WithMarkdown() *EditMessageBuilder {
	b.config.ParseMode = "Markdown"

	return b
}

// WithUserID устанавливает ID пользователя для логирования.
func (b *EditMessageBuilder) WithUserID(userID int64) *EditMessageBuilder {
	b.userID = userID

	return b
}

// WithOperation устанавливает название операции для логирования.
func (b *EditMessageBuilder) WithOperation(operation string) *EditMessageBuilder {
	b.operation = operation

	return b
}

// Send отправляет редактирование сообщения.
func (b *EditMessageBuilder) Send() error {
	return b.factory.sendWithLogging(b.config, b.chatID, b.userID, b.operation, "builder_edit")
}

// =============================================================================
// ВНУТРЕННЯЯ ЛОГИКА
// =============================================================================

// sendWithLogging отправляет сообщение с логированием и обработкой ошибок.
func (f *MessageFactory) sendWithLogging(
	msg tgbotapi.Chattable,
	chatID int64,
	userID int64,
	operation string,
	messageType string,
) error {
	// Логирование перед отправкой
	f.logOutgoingMessage(chatID, userID, operation, messageType)

	// Отправка
	_, err := f.bot.Send(msg)

	// Обработка ошибки
	if err != nil {
		return f.errorHandler.HandleTelegramError(err, chatID, userID, operation)
	}

	return nil
}

// logOutgoingMessage логирует исходящее сообщение.
func (f *MessageFactory) logOutgoingMessage(chatID int64, userID int64, operation string, messageType string) {
	requestID := "out_" + time.Now().Format("20060102150405") + "_" + operation

	f.logger.Telegram().InfoWithContext(
		"Outgoing message",
		requestID,
		userID,
		chatID,
		operation,
		map[string]interface{}{
			"message_type": messageType,
			"chat_id":      chatID,
			"user_id":      userID,
		},
	)
}

// =============================================================================
// SEMANTIC MESSAGE HELPERS - для типичных паттернов сообщений
// =============================================================================
// Эти методы создают часто используемые типы сообщений с правильным форматированием.
// Преимущества:
// 1. Стандартизация внешнего вида сообщений об ошибках, предупреждениях и успехе
// 2. Автоматическая локализация ключевых слов
// 3. Сокращение дублирования кода

// SendError отправляет сообщение об ошибке с красным значком и жирным текстом.
func (f *MessageFactory) SendError(chatID int64, errorText string) error {
	message := fmt.Sprintf("❌ *Ошибка*\n\n%s", errorText)
	return f.SendHTML(chatID, message)
}

// SendWarning отправляет предупреждение с желтым значком и курсивным текстом.
func (f *MessageFactory) SendWarning(chatID int64, warningText string) error {
	message := fmt.Sprintf("⚠️ _Предупреждение_\n\n%s", warningText)
	return f.SendHTML(chatID, message)
}

// SendSuccess отправляет сообщение об успехе с зеленым значком.
func (f *MessageFactory) SendSuccess(chatID int64, successText string) error {
	message := fmt.Sprintf("✅ %s", successText)
	return f.SendHTML(chatID, message)
}

// SendInfo отправляет информационное сообщение с синим значком.
func (f *MessageFactory) SendInfo(chatID int64, infoText string) error {
	message := fmt.Sprintf("ℹ️ %s", infoText)
	return f.SendText(chatID, message)
}

// SendConfirmation отправляет запрос подтверждения с кнопками Да/Нет.
func (f *MessageFactory) SendConfirmation(chatID int64, questionText string, yesCallback, noCallback string) error {
	message := fmt.Sprintf("🤔 %s", questionText)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да", yesCallback),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет", noCallback),
		),
	)
	return f.SendWithKeyboard(chatID, message, keyboard)
}

// SendLocalizedError отправляет локализованное сообщение об ошибке.
func (f *MessageFactory) SendLocalizedError(chatID int64, localizer *localization.Localizer, lang, errorKey string, args ...interface{}) error {
	errorText := localizer.Get(lang, errorKey)
	if len(args) > 0 {
		errorText = fmt.Sprintf(errorText, args...)
	}
	return f.SendError(chatID, errorText)
}

// SendLocalizedSuccess отправляет локализованное сообщение об успехе.
func (f *MessageFactory) SendLocalizedSuccess(chatID int64, localizer *localization.Localizer, lang, successKey string, args ...interface{}) error {
	successText := localizer.Get(lang, successKey)
	if len(args) > 0 {
		successText = fmt.Sprintf(successText, args...)
	}
	return f.SendSuccess(chatID, successText)
}

// SendLocalizedWarning отправляет локализованное предупреждение.
func (f *MessageFactory) SendLocalizedWarning(chatID int64, localizer *localization.Localizer, lang, warningKey string, args ...interface{}) error {
	warningText := localizer.Get(lang, warningKey)
	if len(args) > 0 {
		warningText = fmt.Sprintf(warningText, args...)
	}
	return f.SendWarning(chatID, warningText)
}
