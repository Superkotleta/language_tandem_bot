package validation_test

import (
	"testing"
	"time"

	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"
	"language-exchange-bot/internal/validation"
)

// createValidationService создает сервис валидации для тестов.
func createValidationService(t *testing.T) *validation.Service {
	t.Helper()

	adminNotifier := errors.NewAdminNotifier([]int64{123456789}, nil)
	errorHandler := errors.NewErrorHandler(adminNotifier)

	return validation.NewService(errorHandler)
}

// TestValidationServiceUserValidation тестирует валидацию пользователей.
func TestValidationServiceUserValidation(t *testing.T) {
	t.Parallel()

	validationService := createValidationService(t)

	t.Run("ValidateUserWithErrorHandling", func(t *testing.T) {
		t.Parallel()

		// Создаем валидного пользователя
		user := &models.User{
			ID:                     1,
			TelegramID:             123456789,
			FirstName:              "John",
			Username:               "john_doe",
			NativeLanguageCode:     "en",
			TargetLanguageCode:     "ru",
			TargetLanguageLevel:    "beginner",
			InterfaceLanguageCode:  "en",
			State:                  "idle",
			Status:                 "active",
			ProfileCompletionLevel: 50,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
			Interests:              []int{1, 2},
			TimeAvailability: &models.TimeAvailability{
				DayType:      "any",
				SpecificDays: []string{},
				TimeSlot:     "any",
			},
			FriendshipPreferences: &models.FriendshipPreferences{
				ActivityType:       "casual_chat",
				CommunicationStyle: "text",
				CommunicationFreq:  "weekly",
			},
		}

		// Валидация должна пройти успешно
		err := validationService.ValidateUserWithErrorHandling(user, 123456789, 987654321, "TestUserValidation")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("ValidateUserWithErrorHandling_InvalidData", func(t *testing.T) {
		t.Parallel()

		// Создаем невалидного пользователя
		user := &models.User{
			ID:                     1,
			TelegramID:             0,                  // Невалидный Telegram ID
			FirstName:              "",                 // Пустое имя
			Username:               "invalid@username", // Невалидный username
			NativeLanguageCode:     "en",
			TargetLanguageCode:     "ru",
			TargetLanguageLevel:    "beginner",
			InterfaceLanguageCode:  "invalid",       // Невалидный код языка
			State:                  "invalid_state", // Невалидное состояние
			Status:                 "active",
			ProfileCompletionLevel: 0,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
			Interests:              []int{},
			TimeAvailability: &models.TimeAvailability{
				DayType:      "any",
				SpecificDays: []string{},
				TimeSlot:     "any",
			},
			FriendshipPreferences: &models.FriendshipPreferences{
				ActivityType:       "casual_chat",
				CommunicationStyle: "text",
				CommunicationFreq:  "weekly",
			},
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
}

// TestValidationServiceRegistrationValid тестирует валидную регистрацию.
func TestValidationServiceRegistrationValid(t *testing.T) {
	t.Parallel()

	validationService := createValidationService(t)

	// Валидные данные регистрации
	err := validationService.ValidateUserRegistrationWithErrorHandling(
		123456789, "john_doe", "John", "en", 123456789, 987654321, "TestUserRegistration")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestValidationServiceRegistrationInvalid тестирует невалидную регистрацию.
func TestValidationServiceRegistrationInvalid(t *testing.T) {
	t.Parallel()

	validationService := createValidationService(t)

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
}

// TestValidationServiceInterestsValid тестирует валидные интересы.
func TestValidationServiceInterestsValid(t *testing.T) {
	t.Parallel()

	validationService := createValidationService(t)

	// Валидные интересы
	interestIDs := []int{1, 2, 3}

	err := validationService.ValidateUserInterestsWithErrorHandling(interestIDs, 123456789, 987654321, "TestUserInterests")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestValidationServiceInterestsInvalid тестирует невалидные интересы.
func TestValidationServiceInterestsInvalid(t *testing.T) {
	t.Parallel()

	validationService := createValidationService(t)

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
}

// TestValidationServiceLanguagesValid тестирует валидные языки.
func TestValidationServiceLanguagesValid(t *testing.T) {
	t.Parallel()

	validationService := createValidationService(t)

	// Валидные языки
	err := validationService.ValidateUserLanguagesWithErrorHandling("en", "ru", 123456789, 987654321, "TestUserLanguages")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestValidationServiceLanguagesSame тестирует одинаковые языки.
func TestValidationServiceLanguagesSame(t *testing.T) {
	t.Parallel()

	validationService := createValidationService(t)

	// Одинаковые языки (невалидно)
	err := validationService.ValidateUserLanguagesWithErrorHandling("en", "en", 123456789, 987654321, "TestUserLanguages")
	if err == nil {
		t.Error("Expected validation error, got nil")
	}

	// Проверяем, что это ошибка валидации
	if !errors.IsValidationError(err) {
		t.Error("Expected validation error type")
	}
}

// TestValidatorString тестирует валидацию строк.
func TestValidatorString(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}

// TestValidatorInt тестирует валидацию чисел.
func TestValidatorInt(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}

// TestValidatorLanguageCode тестирует валидацию кодов языков.
func TestValidatorLanguageCode(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}

// TestValidatorTelegramID тестирует валидацию Telegram ID.
func TestValidatorTelegramID(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}

// TestValidatorChatID тестирует валидацию Chat ID.
func TestValidatorChatID(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}

// TestValidatorUserState тестирует валидацию состояний пользователя.
func TestValidatorUserState(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}

// TestValidatorInterestID тестирует валидацию ID интересов.
func TestValidatorInterestID(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}

// TestValidatorLanguageLevel тестирует валидацию уровней языков.
func TestValidatorLanguageLevel(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}

// TestValidatorFeedbackText тестирует валидацию текста обратной связи.
func TestValidatorFeedbackText(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}

// TestValidatorCallbackData тестирует валидацию данных callback.
func TestValidatorCallbackData(t *testing.T) {
	t.Parallel()

	validator := validation.NewValidator()

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
}
