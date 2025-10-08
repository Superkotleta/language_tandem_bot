package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	customerrors "language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/logging"
)

func TestNewMessageFactory(t *testing.T) {
	// Создаем реальные объекты для тестирования
	bot := &tgbotapi.BotAPI{}
	errorHandler := &customerrors.ErrorHandler{}
	loggingService := &logging.LoggingService{}

	factory := NewMessageFactory(bot, errorHandler, loggingService)

	assert.NotNil(t, factory)
	assert.Equal(t, bot, factory.bot)
	assert.Equal(t, errorHandler, factory.errorHandler)
	assert.Equal(t, loggingService, factory.logger)
}

func TestMessageFactory_NewMessage(t *testing.T) {
	bot := &tgbotapi.BotAPI{}
	errorHandler := &customerrors.ErrorHandler{}
	loggingService := &logging.LoggingService{}

	factory := NewMessageFactory(bot, errorHandler, loggingService)

	builder := factory.NewMessage(12345)

	assert.NotNil(t, builder)
	assert.Equal(t, factory, builder.factory)
	assert.Equal(t, int64(12345), builder.chatID)
	assert.Equal(t, "NewMessage", builder.operation)
}

func TestMessageFactory_NewEditMessage(t *testing.T) {
	bot := &tgbotapi.BotAPI{}
	errorHandler := &customerrors.ErrorHandler{}
	loggingService := &logging.LoggingService{}

	factory := NewMessageFactory(bot, errorHandler, loggingService)

	builder := factory.NewEditMessage(12345, 67890)

	assert.NotNil(t, builder)
	assert.Equal(t, factory, builder.factory)
	assert.Equal(t, int64(12345), builder.chatID)
	assert.Equal(t, 67890, builder.messageID)
	assert.Equal(t, "NewEditMessage", builder.operation)
}

func TestMessageBuilder_WithText(t *testing.T) {
	bot := &tgbotapi.BotAPI{}
	errorHandler := &customerrors.ErrorHandler{}
	loggingService := &logging.LoggingService{}

	factory := NewMessageFactory(bot, errorHandler, loggingService)
	builder := factory.NewMessage(12345)

	result := builder.WithText("Test message")

	assert.Equal(t, builder, result)
	assert.Equal(t, "Test message", builder.config.Text)
}

func TestMessageBuilder_WithHTML(t *testing.T) {
	bot := &tgbotapi.BotAPI{}
	errorHandler := &customerrors.ErrorHandler{}
	loggingService := &logging.LoggingService{}

	factory := NewMessageFactory(bot, errorHandler, loggingService)
	builder := factory.NewMessage(12345)

	result := builder.WithHTML()

	assert.Equal(t, builder, result)
	assert.Equal(t, "HTML", builder.config.ParseMode)
}

func TestEditMessageBuilder_WithText(t *testing.T) {
	bot := &tgbotapi.BotAPI{}
	errorHandler := &customerrors.ErrorHandler{}
	loggingService := &logging.LoggingService{}

	factory := NewMessageFactory(bot, errorHandler, loggingService)
	builder := factory.NewEditMessage(12345, 67890)

	result := builder.WithText("Updated message")

	assert.Equal(t, builder, result)
	assert.Equal(t, "Updated message", builder.config.Text)
}
