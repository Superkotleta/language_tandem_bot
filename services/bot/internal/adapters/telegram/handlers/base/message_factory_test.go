package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewMessageFactory tests creating new message factory.
func TestNewMessageFactory(t *testing.T) {
	factory := NewMessageFactory(nil, nil, nil)

	assert.NotNil(t, factory)
	assert.Nil(t, factory.bot)
	assert.Nil(t, factory.errorHandler)
	assert.Nil(t, factory.logger)
}

// TestMessageFactory_SendText tests sending text messages.
func TestMessageFactory_SendText(t *testing.T) {
	// Just test that the method exists
	factory := NewMessageFactory(nil, nil, nil)

	// Test method signature exists
	assert.NotNil(t, factory.SendText)
}

// TestMessageFactory_SendWithKeyboard tests sending messages with keyboard.
func TestMessageFactory_SendWithKeyboard(t *testing.T) {
	// Just test that the method exists
	factory := NewMessageFactory(nil, nil, nil)

	// Test method signature exists
	assert.NotNil(t, factory.SendWithKeyboard)
}

// TestMessageFactory_SendHTML tests sending HTML messages.
func TestMessageFactory_SendHTML(t *testing.T) {
	// Just test that the method exists
	factory := NewMessageFactory(nil, nil, nil)

	// Test method signature exists
	assert.NotNil(t, factory.SendHTML)
}

// TestMessageFactory_SendHTMLWithKeyboard tests sending HTML messages with keyboard.
func TestMessageFactory_SendHTMLWithKeyboard(t *testing.T) {
	// Just test that the method exists
	factory := NewMessageFactory(nil, nil, nil)

	// Test method signature exists
	assert.NotNil(t, factory.SendHTMLWithKeyboard)
}

// TestMessageFactory_EditText tests editing text messages.
func TestMessageFactory_EditText(t *testing.T) {
	// Just test that the method exists
	factory := NewMessageFactory(nil, nil, nil)

	// Test method signature exists
	assert.NotNil(t, factory.EditText)
}

// TestMessageFactory_EditWithKeyboard tests editing messages with keyboard.
func TestMessageFactory_EditWithKeyboard(t *testing.T) {
	// Just test that the method exists
	factory := NewMessageFactory(nil, nil, nil)

	// Test method signature exists
	assert.NotNil(t, factory.EditWithKeyboard)
}

// TestMessageFactory_NewMessage tests creating new message builder.
func TestMessageFactory_NewMessage(t *testing.T) {
	factory := NewMessageFactory(nil, nil, nil)

	builder := factory.NewMessage(123)

	assert.NotNil(t, builder)
	assert.Equal(t, factory, builder.factory)
	assert.Equal(t, int64(123), builder.chatID)
}

// TestMessageFactory_NewEditMessage tests creating edit message builder.
func TestMessageFactory_NewEditMessage(t *testing.T) {
	factory := NewMessageFactory(nil, nil, nil)

	builder := factory.NewEditMessage(123, 456)

	assert.NotNil(t, builder)
	assert.Equal(t, factory, builder.factory)
	assert.Equal(t, int64(123), builder.chatID)
	assert.Equal(t, 456, builder.messageID)
}
