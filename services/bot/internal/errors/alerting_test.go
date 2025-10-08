package errors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewAdminNotifier tests creating new admin notifier
func TestNewAdminNotifier(t *testing.T) {
	adminChatIDs := []int64{123, 456}
	botAPI := "mock_bot_api"

	notifier := NewAdminNotifier(adminChatIDs, botAPI)

	assert.NotNil(t, notifier)
	assert.Equal(t, adminChatIDs, notifier.adminChatIDs)
	assert.Equal(t, botAPI, notifier.botAPI)
}

// TestAdminNotifierImpl_NotifyCriticalError tests critical error notification
func TestAdminNotifierImpl_NotifyCriticalError(t *testing.T) {
	adminChatIDs := []int64{123, 456}
	botAPI := "mock_bot_api"

	notifier := NewAdminNotifier(adminChatIDs, botAPI)

	// Create test error
	customErr := &CustomError{
		Type:      ErrorTypeInternal,
		Timestamp: time.Now(),
		RequestID: "req123",
		Message:   "Test critical error",
		Context: map[string]interface{}{
			"user_id":   123,
			"chat_id":   456,
			"operation": "test_operation",
			"extra":     "additional info",
		},
	}

	// Test method exists and doesn't panic
	assert.NotPanics(t, func() {
		notifier.NotifyCriticalError(customErr)
	})
}

// TestAdminNotifierImpl_formatContext tests context formatting
func TestAdminNotifierImpl_formatContext(t *testing.T) {
	adminChatIDs := []int64{123}
	botAPI := "mock_bot_api"

	notifier := NewAdminNotifier(adminChatIDs, botAPI)

	context := map[string]interface{}{
		"user_id":   123,
		"chat_id":   456,
		"operation": "test_op",
		"extra":     "value",
	}

	// Test method exists
	assert.NotPanics(t, func() {
		result := notifier.formatContext(context)
		assert.NotEmpty(t, result)
	})
}

// TestAdminNotifierImpl_sendToAdmins tests sending messages to admins
func TestAdminNotifierImpl_sendToAdmins(t *testing.T) {
	adminChatIDs := []int64{123, 456}
	botAPI := "mock_bot_api"

	notifier := NewAdminNotifier(adminChatIDs, botAPI)

	message := "Test alert message"

	// Test method exists and doesn't panic
	assert.NotPanics(t, func() {
		notifier.sendToAdmins(message)
	})
}
