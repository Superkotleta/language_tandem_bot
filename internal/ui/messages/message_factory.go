package messages

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageFactory handles message creation
type MessageFactory struct{}

// NewMessageFactory creates a new message factory
func NewMessageFactory() *MessageFactory {
	return &MessageFactory{}
}

// NewText creates a simple text message
func (mf *MessageFactory) NewText(chatID int64, text string) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, text)
}

// NewKeyboardMessage creates a message with a reply keyboard
func (mf *MessageFactory) NewKeyboardMessage(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	return msg
}

// NewInlineKeyboardMessage creates a message with an inline keyboard
func (mf *MessageFactory) NewInlineKeyboardMessage(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	return msg
}
