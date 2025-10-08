package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAlertLevel_String tests AlertLevel string representation
func TestAlertLevel_String(t *testing.T) {
	tests := []struct {
		level    AlertLevel
		expected string
	}{
		{AlertLevelInfo, "INFO"},
		{AlertLevelWarning, "WARNING"},
		{AlertLevelCritical, "CRITICAL"},
		{AlertLevelEmergency, "EMERGENCY"},
		{AlertLevel(-1), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

// TestNewCentralizedErrorHandler tests creating new centralized error handler
func TestNewCentralizedErrorHandler(t *testing.T) {
	handler := NewCentralizedErrorHandler()

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.alerts)
	assert.NotNil(t, handler.notifiers)
	assert.NotNil(t, handler.logger)
}

// TestCentralizedErrorHandler_HandleError tests error handling
func TestCentralizedErrorHandler_HandleError(t *testing.T) {
	handler := NewCentralizedErrorHandler()

	originalErr := errors.New("test error")

	// Test method exists and doesn't panic
	assert.NotPanics(t, func() {
		err := handler.HandleError(nil, originalErr, "req123", 123, 456, "test_op")
		// May return error, but shouldn't panic
		_ = err
	})
}

// TestCentralizedErrorHandler_GetActiveAlerts tests getting active alerts
func TestCentralizedErrorHandler_GetActiveAlerts(t *testing.T) {
	handler := NewCentralizedErrorHandler()

	alerts := handler.GetActiveAlerts()

	assert.NotNil(t, alerts)
	assert.IsType(t, map[string]*Alert{}, alerts)
}

// TestCentralizedErrorHandler_ResolveAlert tests alert resolution
func TestCentralizedErrorHandler_ResolveAlert(t *testing.T) {
	handler := NewCentralizedErrorHandler()

	alertID := "test_alert_123"

	// Test method exists and doesn't panic
	assert.NotPanics(t, func() {
		handler.ResolveAlert(alertID)
	})
}
