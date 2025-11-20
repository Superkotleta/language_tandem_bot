package validation

import (
	"fmt"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
)

// UserValidator валидирует данные пользователя.
type UserValidator struct {
	validator *Validator
}

// NewUserValidator создает новый валидатор пользователей.
func NewUserValidator() *UserValidator {
	return &UserValidator{
		validator: NewValidator(),
	}
}

// ValidateUser валидирует данные пользователя.
func (uv *UserValidator) ValidateUser(user *models.User) *Result {
	result := NewResult()

	uv.validateTelegramID(user.TelegramID, result)
	uv.validateFirstName(user.FirstName, result)
	uv.validateUsername(user.Username, result)
	uv.validateInterfaceLanguage(user.InterfaceLanguageCode, result)
	uv.validateUserState(user.State, result)

	return result
}

// validateTelegramID валидирует Telegram ID.
func (uv *UserValidator) validateTelegramID(telegramID int64, result *Result) {
	if errors := uv.validator.ValidateTelegramID(telegramID); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("telegram_id", err)
		}
	}
}

// validateFirstName валидирует имя пользователя.
func (uv *UserValidator) validateFirstName(firstName string, result *Result) {
	if errors := uv.validator.ValidateString(firstName, []string{"required", "max:50"}); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("first_name", err)
		}
	}
}

// validateUsername валидирует username.
func (uv *UserValidator) validateUsername(username string, result *Result) {
	if username != "" {
		if errors := uv.validator.ValidateString(username, []string{"username", "max:50"}); len(errors) > 0 {
			for _, err := range errors {
				result.AddError("username", err)
			}
		}
	}
}

// validateInterfaceLanguage валидирует язык интерфейса.
func (uv *UserValidator) validateInterfaceLanguage(languageCode string, result *Result) {
	if errors := uv.validator.ValidateLanguageCode(languageCode); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("interface_language_code", err)
		}
	}
}

// validateUserState валидирует состояние пользователя.
func (uv *UserValidator) validateUserState(state string, result *Result) {
	if errors := uv.validator.ValidateUserState(state); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("state", err)
		}
	}
}

// ValidateUserRegistration валидирует данные при регистрации пользователя.
func (uv *UserValidator) ValidateUserRegistration(telegramID int, username, firstName, languageCode string) *Result {
	result := NewResult()

	// Валидация Telegram ID
	if errors := uv.validator.ValidateTelegramID(int64(telegramID)); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("telegram_id", err)
		}
	}

	// Валидация имени пользователя
	if errors := uv.validator.ValidateString(firstName, []string{"required", "max:50"}); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("first_name", err)
		}
	}

	// Валидация username (если есть)
	if username != "" {
		if errors := uv.validator.ValidateString(username, []string{"username", "max:50"}); len(errors) > 0 {
			for _, err := range errors {
				result.AddError("username", err)
			}
		}
	}

	// Валидация кода языка
	if errors := uv.validator.ValidateLanguageCode(languageCode); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("language_code", err)
		}
	}

	return result
}

// ValidateUserUpdate валидирует данные при обновлении пользователя.
func (uv *UserValidator) ValidateUserUpdate(user *models.User) *Result {
	result := NewResult()

	// Валидация имени пользователя
	if errors := uv.validator.ValidateString(user.FirstName, []string{"required", "max:50"}); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("first_name", err)
		}
	}

	// Валидация username (если есть)
	if user.Username != "" {
		if errors := uv.validator.ValidateString(user.Username, []string{"username", "max:50"}); len(errors) > 0 {
			for _, err := range errors {
				result.AddError("username", err)
			}
		}
	}

	// Валидация кода языка интерфейса
	if errors := uv.validator.ValidateLanguageCode(user.InterfaceLanguageCode); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("interface_language_code", err)
		}
	}

	// Валидация состояния пользователя
	if errors := uv.validator.ValidateUserState(user.State); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("state", err)
		}
	}

	return result
}

// ValidateUserInterests валидирует интересы пользователя.
func (uv *UserValidator) ValidateUserInterests(interestIDs []int) *Result {
	result := NewResult()

	if len(interestIDs) == 0 {
		result.AddError("interests", "Необходимо выбрать хотя бы один интерес")

		return result
	}

	if len(interestIDs) > localization.MaxInterestCount {
		result.AddError("interests", fmt.Sprintf("Максимум %d интересов", localization.MaxInterestCount))

		return result
	}

	// Валидация каждого ID интереса
	for index, interestID := range interestIDs {
		if errors := uv.validator.ValidateInterestID(interestID); len(errors) > 0 {
			for _, err := range errors {
				result.AddError("interests", err)
			}
		}

		// Проверка на дубликаты
		for j, otherID := range interestIDs {
			if index != j && interestID == otherID {
				result.AddError("interests", "Дублирующиеся интересы не допускаются")

				break
			}
		}
	}

	return result
}

// ValidateUserLanguages валидирует языки пользователя.
func (uv *UserValidator) ValidateUserLanguages(nativeLanguage, targetLanguage string) *Result {
	result := NewResult()

	// Валидация родного языка
	if errors := uv.validator.ValidateLanguageCode(nativeLanguage); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("native_language", err)
		}
	}

	// Валидация целевого языка
	if errors := uv.validator.ValidateLanguageCode(targetLanguage); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("target_language", err)
		}
	}

	// Проверка, что языки разные
	if nativeLanguage == targetLanguage {
		result.AddError("languages", "Родной и целевой языки должны отличаться")
	}

	return result
}

// ValidateUserLanguageLevel валидирует уровень языка пользователя.
func (uv *UserValidator) ValidateUserLanguageLevel(level int) *Result {
	result := NewResult()

	if errors := uv.validator.ValidateLanguageLevel(level); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("language_level", err)
		}
	}

	return result
}
