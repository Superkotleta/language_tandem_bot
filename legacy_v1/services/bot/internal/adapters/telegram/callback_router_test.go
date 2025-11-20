package telegram

import (
	"testing"

	"language-exchange-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestCallbackRouter_RegisterSimple(t *testing.T) {
	router := NewCallbackRouter()

	called := false
	handler := func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		called = true

		return nil
	}

	router.RegisterSimple("test_callback", handler)

	user := &models.User{ID: 1}
	callback := &tgbotapi.CallbackQuery{Data: "test_callback"}

	err := router.Handle(callback, user)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestCallbackRouter_RegisterPrefix(t *testing.T) {
	router := NewCallbackRouter()

	var capturedParam string

	handler := func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		capturedParam = params["param"]

		return nil
	}

	router.RegisterPrefix("category_", handler)

	user := &models.User{ID: 1}
	callback := &tgbotapi.CallbackQuery{Data: "category_entertainment"}

	err := router.Handle(callback, user)
	assert.NoError(t, err)
	assert.Equal(t, "entertainment", capturedParam)
}

func TestCallbackRouter_NoHandler(t *testing.T) {
	router := NewCallbackRouter()

	user := &models.User{ID: 1}
	callback := &tgbotapi.CallbackQuery{Data: "unknown_callback"}

	err := router.Handle(callback, user)
	assert.NoError(t, err) // Нет подходящего обработчика - не ошибка
}

func TestCallbackRouter_InvalidRegex(t *testing.T) {
	router := NewCallbackRouter()

	handler := func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		return nil
	}

	err := router.Register("[invalid", handler)
	assert.Error(t, err)
}

func TestCallbackRouter_MultipleRoutes(t *testing.T) {
	router := NewCallbackRouter()

	callCount := 0
	handler1 := func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		callCount = 1

		return nil
	}

	handler2 := func(callback *tgbotapi.CallbackQuery, user *models.User, params map[string]string) error {
		callCount = 2

		return nil
	}

	router.RegisterSimple("callback1", handler1)
	router.RegisterSimple("callback2", handler2)

	user := &models.User{ID: 1}

	// Test first callback
	callback1 := &tgbotapi.CallbackQuery{Data: "callback1"}
	err := router.Handle(callback1, user)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Test second callback
	callback2 := &tgbotapi.CallbackQuery{Data: "callback2"}
	err = router.Handle(callback2, user)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}
