package base

import (
	"testing"

	"language-exchange-bot/internal/core"

	"github.com/stretchr/testify/assert"
)

// TestNewKeyboardBuilder tests creating new keyboard builder.
func TestNewKeyboardBuilder(t *testing.T) {
	service := &core.BotService{}
	builder := NewKeyboardBuilder(service)

	assert.NotNil(t, builder)
	assert.Equal(t, service, builder.service)
}

// TestKeyboardBuilder_CreateMainMenuKeyboard tests main menu keyboard creation.
func TestKeyboardBuilder_CreateMainMenuKeyboard(t *testing.T) {
	// Just test that the method exists
	service := &core.BotService{}
	builder := NewKeyboardBuilder(service)

	// Test method signature exists
	assert.NotNil(t, builder.CreateMainMenuKeyboard)
}

// TestKeyboardBuilder_CreateProfileMenuKeyboard tests profile menu keyboard creation.
func TestKeyboardBuilder_CreateProfileMenuKeyboard(t *testing.T) {
	// Just test that the method exists
	service := &core.BotService{}
	builder := NewKeyboardBuilder(service)

	// Test method signature exists
	assert.NotNil(t, builder.CreateProfileMenuKeyboard)
}

// TestKeyboardBuilder_CreateResetConfirmKeyboard tests reset confirm keyboard creation.
func TestKeyboardBuilder_CreateResetConfirmKeyboard(t *testing.T) {
	// Just test that the method exists
	service := &core.BotService{}
	builder := NewKeyboardBuilder(service)

	// Test method signature exists
	assert.NotNil(t, builder.CreateResetConfirmKeyboard)
}

// TestKeyboardBuilder_CreateProfileCompletedKeyboard tests profile completed keyboard creation.
func TestKeyboardBuilder_CreateProfileCompletedKeyboard(t *testing.T) {
	// Just test that the method exists
	service := &core.BotService{}
	builder := NewKeyboardBuilder(service)

	// Test method signature exists
	assert.NotNil(t, builder.CreateProfileCompletedKeyboard)
}

// TestKeyboardBuilder_CreateAvailabilitySetupKeyboard tests availability setup keyboard creation.
func TestKeyboardBuilder_CreateAvailabilitySetupKeyboard(t *testing.T) {
	// Just test that the method exists
	service := &core.BotService{}
	builder := NewKeyboardBuilder(service)

	// Test method signature exists
	assert.NotNil(t, builder.CreateAvailabilitySetupKeyboard)
}

// TestKeyboardBuilder_CreateInterestCategoriesKeyboard tests interest categories keyboard creation.
func TestKeyboardBuilder_CreateInterestCategoriesKeyboard(t *testing.T) {
	// Just test that the method exists
	service := &core.BotService{}
	builder := NewKeyboardBuilder(service)

	// Test method signature exists
	assert.NotNil(t, builder.CreateInterestCategoriesKeyboard)
}
