package handlers

import (
	"testing"
)

// TestNewBaseHandler_NilInputs тестирует создание BaseHandler с nil значениями
func TestNewBaseHandler_NilInputs(t *testing.T) {
	// Тестируем создание BaseHandler с nil значениями
	baseHandler := NewBaseHandler(nil, nil, nil, nil, nil)

	if baseHandler == nil {
		t.Fatal("NewBaseHandler returned nil even with nil inputs")
	}

	// Проверяем, что поля установлены в nil
	if baseHandler.bot != nil {
		t.Error("Bot field should be nil")
	}

	if baseHandler.service != nil {
		t.Error("Service field should be nil")
	}

	if baseHandler.keyboardBuilder != nil {
		t.Error("KeyboardBuilder field should be nil")
	}

	if baseHandler.errorHandler != nil {
		t.Error("ErrorHandler field should be nil")
	}

	if baseHandler.messageFactory != nil {
		t.Error("MessageFactory field should be nil")
	}
}

// TestBaseHandler_Getters тестирует getter методы BaseHandler
func TestBaseHandler_Getters(t *testing.T) {
	// Создаем BaseHandler с nil значениями для простоты тестирования
	baseHandler := NewBaseHandler(nil, nil, nil, nil, nil)

	// Тестируем getter методы - они должны возвращать nil для nil полей
	if baseHandler.GetBot() != nil {
		t.Error("GetBot() should return nil for nil bot field")
	}

	if baseHandler.GetService() != nil {
		t.Error("GetService() should return nil for nil service field")
	}

	if baseHandler.GetKeyboardBuilder() != nil {
		t.Error("GetKeyboardBuilder() should return nil for nil keyboardBuilder field")
	}

	if baseHandler.GetErrorHandler() != nil {
		t.Error("GetErrorHandler() should return nil for nil errorHandler field")
	}

	if baseHandler.GetMessageFactory() != nil {
		t.Error("GetMessageFactory() should return nil for nil messageFactory field")
	}
}

// TestNewBaseHandler_NotNil тестирует, что NewBaseHandler возвращает не-nil объект
func TestNewBaseHandler_NotNil(t *testing.T) {
	baseHandler := NewBaseHandler(nil, nil, nil, nil, nil)

	if baseHandler == nil {
		t.Fatal("NewBaseHandler should not return nil")
	}
}
