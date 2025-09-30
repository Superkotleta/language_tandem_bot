// Package adapters provides interfaces for external service adapters.
package adapters

import "context"

// BotAdapter определяет интерфейс для различных ботов (Telegram, Discord, etc).
type BotAdapter interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetPlatformName() string
}

// Message представляет универсальное сообщение.
type Message struct {
	ID        string
	UserID    string
	Username  string
	FirstName string
	Text      string
	ChatID    string
	IsCommand bool
	Command   string
	Language  string
}

// CallbackQuery представляет универсальный callback.
type CallbackQuery struct {
	ID      string
	UserID  string
	Data    string
	Message *Message
}

// MessageHandler обрабатывает сообщения для любой платформы.
type MessageHandler interface {
	HandleMessage(msg *Message) error
	HandleCallbackQuery(callback *CallbackQuery) error
}
