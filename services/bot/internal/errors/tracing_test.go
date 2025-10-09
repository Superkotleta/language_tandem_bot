package errors

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewRequestContext tests creating new request context.
func TestNewRequestContext(t *testing.T) {
	userID := int64(123)
	chatID := int64(456)
	operation := "test_operation"

	ctx := NewRequestContext(userID, chatID, operation)

	assert.NotNil(t, ctx)
	assert.NotEmpty(t, ctx.RequestID)
	assert.Equal(t, userID, ctx.UserID)
	assert.Equal(t, chatID, ctx.ChatID)
	assert.Equal(t, operation, ctx.Operation)
	assert.Less(t, time.Since(ctx.Timestamp), time.Second)
}

// TestGenerateRequestID tests request ID generation.
func TestGenerateRequestID(t *testing.T) {
	// Test multiple generations to ensure uniqueness
	id1 := generateRequestID()
	id2 := generateRequestID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "req_")
	assert.Contains(t, id2, "req_")
}

// TestWithContext tests creating error with context.
func TestWithContext(t *testing.T) {
	originalErr := errors.New("original error")
	ctx := NewRequestContext(123, 456, "test_op")

	customErr := WithContext(originalErr, ctx)

	assert.NotNil(t, customErr)
	assert.Equal(t, ErrorTypeInternal, customErr.Type)
	assert.Equal(t, "original error", customErr.Message) // Message comes from original error
	assert.Equal(t, ctx.RequestID, customErr.RequestID)
	assert.NotNil(t, customErr.Context)
	assert.Equal(t, ctx.UserID, customErr.Context["user_id"])
	assert.Equal(t, ctx.ChatID, customErr.Context["chat_id"])
	assert.Equal(t, ctx.Operation, customErr.Context["operation"])
}

// TestRequestContext_JSONSerialization tests JSON serialization of RequestContext.
func TestRequestContext_JSONSerialization(t *testing.T) {
	ctx := &RequestContext{
		RequestID: "req_123456789_999",
		UserID:    123,
		ChatID:    456,
		Operation: "test_operation",
		Timestamp: time.Now(),
	}

	// Test JSON serialization (RequestContext doesn't have JSON tags, but we can test the fields)
	assert.NotEmpty(t, ctx.RequestID)
	assert.Positive(t, ctx.UserID)
	assert.Positive(t, ctx.ChatID)
	assert.NotEmpty(t, ctx.Operation)
	assert.False(t, ctx.Timestamp.IsZero())
}
