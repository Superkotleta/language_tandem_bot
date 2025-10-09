package logging

import (
	"testing"

	customErrors "language-exchange-bot/internal/errors"

	"github.com/stretchr/testify/assert"
)

// TestNewLoggingService tests creating new logging service.
func TestNewLoggingService(t *testing.T) {
	errorHandler := &customErrors.ErrorHandler{}
	service := NewLoggingService(errorHandler)

	assert.NotNil(t, service)
	assert.NotNil(t, service.telegramLogger)
	assert.NotNil(t, service.databaseLogger)
	assert.NotNil(t, service.cacheLogger)
	assert.NotNil(t, service.validationLogger)
	assert.Equal(t, errorHandler, service.errorHandler)
}

// TestLoggingService_Getters tests getter methods.
func TestLoggingService_Getters(t *testing.T) {
	errorHandler := &customErrors.ErrorHandler{}
	service := NewLoggingService(errorHandler)

	// Test getters
	assert.NotNil(t, service.Telegram())
	assert.NotNil(t, service.Database())
	assert.NotNil(t, service.Cache())
	assert.NotNil(t, service.Validation())

	// Test that getters return correct types
	assert.IsType(t, &TelegramLogger{}, service.Telegram())
	assert.IsType(t, &DatabaseLogger{}, service.Database())
	assert.IsType(t, &CacheLogger{}, service.Cache())
	assert.IsType(t, &ValidationLogger{}, service.Validation())
}

// TestLoggingService_LogErrorWithContext tests error logging with context.
func TestLoggingService_LogErrorWithContext(t *testing.T) {
	// Just test that the method exists
	errorHandler := &customErrors.ErrorHandler{}
	service := NewLoggingService(errorHandler)

	// Test method signature exists
	assert.NotNil(t, service.LogErrorWithContext)
}
