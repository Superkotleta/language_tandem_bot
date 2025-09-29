package validation

import (
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"
	"testing"
)

// TestValidationService тестирует сервис валидации
func TestValidationService(t *testing.T) {
	// Создаем мок errorHandler
	adminNotifier := errors.NewAdminNotifier([]int64{123456789}, nil)
	errorHandler := errors.NewErrorHandler(adminNotifier)

	// Создаем сервис валидации
	validationService := NewValidationService(errorHandler)

	t.Run("ValidateUserWithErrorHandling", func(t *testing.T) {
		// Создаем валидного пользователя
		user := &models.User{
			ID:                    1,
			TelegramID:            123456789,
			FirstName:             "John",
			Username:              "john_doe",
			InterfaceLanguageCode: "en",
			State:                 "idle",
		}

		// Валидация должна пройти успешно
		err := validationService.ValidateUserWithErrorHandling(user, 123456789, 987654321, "TestUserValidation")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("ValidateUserWithErrorHandling_InvalidData", func(t *testing.T) {
		// Создаем невалидного пользователя
		user := &models.User{
			ID:                    1,
			TelegramID:            0,                  // Невалидный Telegram ID
			FirstName:             "",                 // Пустое имя
			Username:              "invalid@username", // Невалидный username
			InterfaceLanguageCode: "invalid",          // Невалидный код языка
			State:                 "invalid_state",    // Невалидное состояние
		}

		// Валидация должна вернуть ошибку
		err := validationService.ValidateUserWithErrorHandling(user, 123456789, 987654321, "TestUserValidation")
		if err == nil {
			t.Error("Expected validation error, got nil")
		}

		// Проверяем, что это ошибка валидации
		if !errors.IsValidationError(err) {
			t.Error("Expected validation error type")
		}
	})

	t.Run("ValidateUserRegistrationWithErrorHandling", func(t *testing.T) {
		// Валидные данные регистрации
		err := validationService.ValidateUserRegistrationWithErrorHandling(
			123456789, "john_doe", "John", "en", 123456789, 987654321, "TestUserRegistration")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("ValidateUserRegistrationWithErrorHandling_InvalidData", func(t *testing.T) {
		// Невалидные данные регистрации
		err := validationService.ValidateUserRegistrationWithErrorHandling(
			0, "invalid@username", "", "invalid", 123456789, 987654321, "TestUserRegistration")
		if err == nil {
			t.Error("Expected validation error, got nil")
		}

		// Проверяем, что это ошибка валидации
		if !errors.IsValidationError(err) {
			t.Error("Expected validation error type")
		}
	})

	t.Run("ValidateUserInterestsWithErrorHandling", func(t *testing.T) {
		// Валидные интересы
		interestIDs := []int{1, 2, 3}
		err := validationService.ValidateUserInterestsWithErrorHandling(interestIDs, 123456789, 987654321, "TestUserInterests")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("ValidateUserInterestsWithErrorHandling_InvalidData", func(t *testing.T) {
		// Невалидные интересы (пустой список)
		interestIDs := []int{}
		err := validationService.ValidateUserInterestsWithErrorHandling(interestIDs, 123456789, 987654321, "TestUserInterests")
		if err == nil {
			t.Error("Expected validation error, got nil")
		}

		// Проверяем, что это ошибка валидации
		if !errors.IsValidationError(err) {
			t.Error("Expected validation error type")
		}
	})

	t.Run("ValidateUserLanguagesWithErrorHandling", func(t *testing.T) {
		// Валидные языки
		err := validationService.ValidateUserLanguagesWithErrorHandling("en", "ru", 123456789, 987654321, "TestUserLanguages")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("ValidateUserLanguagesWithErrorHandling_SameLanguages", func(t *testing.T) {
		// Одинаковые языки (невалидно)
		err := validationService.ValidateUserLanguagesWithErrorHandling("en", "en", 123456789, 987654321, "TestUserLanguages")
		if err == nil {
			t.Error("Expected validation error, got nil")
		}

		// Проверяем, что это ошибка валидации
		if !errors.IsValidationError(err) {
			t.Error("Expected validation error type")
		}
	})

	t.Log("All validation service tests completed successfully")
}

// TestValidator тестирует базовый валидатор
func TestValidator(t *testing.T) {
	validator := NewValidator()

	t.Run("ValidateString", func(t *testing.T) {
		// Тест валидной строки
		errors := validator.ValidateString("valid_string", []string{"required", "max:50"})
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидной строки
		errors = validator.ValidateString("", []string{"required"})
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("ValidateInt", func(t *testing.T) {
		// Тест валидного числа
		errors := validator.ValidateInt(5, []string{"required", "min:1", "max:100"})
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидного числа
		errors = validator.ValidateInt(0, []string{"required"})
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("ValidateLanguageCode", func(t *testing.T) {
		// Тест валидного кода языка
		errors := validator.ValidateLanguageCode("en")
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидного кода языка
		errors = validator.ValidateLanguageCode("invalid")
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("ValidateTelegramID", func(t *testing.T) {
		// Тест валидного Telegram ID
		errors := validator.ValidateTelegramID(123456789)
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидного Telegram ID
		errors = validator.ValidateTelegramID(0)
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("ValidateChatID", func(t *testing.T) {
		// Тест валидного Chat ID
		errors := validator.ValidateChatID(123456789)
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидного Chat ID
		errors = validator.ValidateChatID(0)
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("ValidateUserState", func(t *testing.T) {
		// Тест валидного состояния
		errors := validator.ValidateUserState("idle")
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидного состояния
		errors = validator.ValidateUserState("invalid_state")
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("ValidateInterestID", func(t *testing.T) {
		// Тест валидного ID интереса
		errors := validator.ValidateInterestID(1)
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидного ID интереса
		errors = validator.ValidateInterestID(0)
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("ValidateLanguageLevel", func(t *testing.T) {
		// Тест валидного уровня языка
		errors := validator.ValidateLanguageLevel(3)
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидного уровня языка
		errors = validator.ValidateLanguageLevel(0)
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("ValidateFeedbackText", func(t *testing.T) {
		// Тест валидного текста отзыва
		errors := validator.ValidateFeedbackText("This is a valid feedback text with enough characters.")
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидного текста отзыва
		errors = validator.ValidateFeedbackText("Short")
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("ValidateCallbackData", func(t *testing.T) {
		// Тест валидных данных callback
		errors := validator.ValidateCallbackData("valid_callback_data")
		if len(errors) > 0 {
			t.Errorf("Expected no errors, got: %v", errors)
		}

		// Тест невалидных данных callback
		errors = validator.ValidateCallbackData("")
		if len(errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Log("All validator tests completed successfully")
}
