package validation

import (
	"strings"
)

// MessageValidator валидирует сообщения и callback'и.
type MessageValidator struct {
	validator *Validator
}

// NewMessageValidator создает новый валидатор сообщений.
func NewMessageValidator() *MessageValidator {
	return &MessageValidator{
		validator: NewValidator(),
	}
}

// ValidateMessage валидирует входящее сообщение.
func (mv *MessageValidator) ValidateMessage(chatID, userID int64, text string) *Result {
	result := NewResult()

	// Валидация Chat ID
	if errors := mv.validator.ValidateChatID(chatID); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("chat_id", err)
		}
	}

	// Валидация User ID
	if errors := mv.validator.ValidateTelegramID(int64(userID)); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("user_id", err)
		}
	}

	// Валидация текста сообщения
	if text != "" {
		if errors := mv.validator.ValidateString(text, []string{"max:4096"}); len(errors) > 0 {
			for _, err := range errors {
				result.AddError("text", err)
			}
		}
	}

	return result
}

// ValidateCallbackQuery валидирует callback query.
func (mv *MessageValidator) ValidateCallbackQuery(chatID, userID int64, data string) *Result {
	result := NewResult()

	// Валидация Chat ID
	if errors := mv.validator.ValidateChatID(chatID); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("chat_id", err)
		}
	}

	// Валидация User ID
	if errors := mv.validator.ValidateTelegramID(int64(userID)); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("user_id", err)
		}
	}

	// Валидация данных callback
	if errors := mv.validator.ValidateCallbackData(data); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("callback_data", err)
		}
	}

	return result
}

// ValidateFeedbackMessage валидирует сообщение отзыва.
func (mv *MessageValidator) ValidateFeedbackMessage(chatID, userID int64, text string) *Result {
	result := NewResult()

	// Валидация Chat ID
	if errors := mv.validator.ValidateChatID(chatID); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("chat_id", err)
		}
	}

	// Валидация User ID
	if errors := mv.validator.ValidateTelegramID(int64(userID)); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("user_id", err)
		}
	}

	// Валидация текста отзыва
	if errors := mv.validator.ValidateFeedbackText(text); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("feedback_text", err)
		}
	}

	return result
}

// ValidateCommand валидирует команду.
func (mv *MessageValidator) ValidateCommand(chatID, userID int64, command string) *Result {
	result := NewResult()

	// Валидация Chat ID
	if errors := mv.validator.ValidateChatID(chatID); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("chat_id", err)
		}
	}

	// Валидация User ID
	if errors := mv.validator.ValidateTelegramID(int64(userID)); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("user_id", err)
		}
	}

	// Валидация команды
	if strings.TrimSpace(command) == "" {
		result.AddError("command", "Команда обязательна")

		return result
	}

	// Проверяем, что команда начинается с /
	if !strings.HasPrefix(command, "/") {
		result.AddError("command", "Команда должна начинаться с /")
	}

	// Проверяем длину команды
	if len(command) > maxCommandLength {
		result.AddError("command", "Команда слишком длинная")
	}

	// Проверяем, что команда содержит только разрешенные символы
	if errors := mv.validator.ValidateString(command, []string{"alphanumeric"}); len(errors) > 0 {
		result.AddError("command", "Команда содержит недопустимые символы")
	}

	return result
}

// ValidateLanguageSelection валидирует выбор языка.
func (mv *MessageValidator) ValidateLanguageSelection(chatID, userID int64, languageCode string) *Result {
	result := NewResult()

	// Валидация Chat ID
	if errors := mv.validator.ValidateChatID(chatID); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("chat_id", err)
		}
	}

	// Валидация User ID
	if errors := mv.validator.ValidateTelegramID(int64(userID)); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("user_id", err)
		}
	}

	// Валидация кода языка
	if errors := mv.validator.ValidateLanguageCode(languageCode); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("language_code", err)
		}
	}

	return result
}

// ValidateInterestSelection валидирует выбор интереса.
func (mv *MessageValidator) ValidateInterestSelection(chatID, userID int64, interestID int) *Result {
	result := NewResult()

	// Валидация Chat ID
	if errors := mv.validator.ValidateChatID(chatID); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("chat_id", err)
		}
	}

	// Валидация User ID
	if errors := mv.validator.ValidateTelegramID(int64(userID)); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("user_id", err)
		}
	}

	// Валидация ID интереса
	if errors := mv.validator.ValidateInterestID(interestID); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("interest_id", err)
		}
	}

	return result
}

// ValidateLanguageLevelSelection валидирует выбор уровня языка.
func (mv *MessageValidator) ValidateLanguageLevelSelection(chatID, userID int64, level int) *Result {
	result := NewResult()

	// Валидация Chat ID
	if errors := mv.validator.ValidateChatID(chatID); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("chat_id", err)
		}
	}

	// Валидация User ID
	if errors := mv.validator.ValidateTelegramID(int64(userID)); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("user_id", err)
		}
	}

	// Валидация уровня языка
	if errors := mv.validator.ValidateLanguageLevel(level); len(errors) > 0 {
		for _, err := range errors {
			result.AddError("language_level", err)
		}
	}

	return result
}
