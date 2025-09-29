package validation

import (
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"
)

// ValidationService предоставляет сервис валидации с интеграцией ошибок
type ValidationService struct {
	userValidator    *UserValidator
	messageValidator *MessageValidator
	errorHandler     *errors.ErrorHandler
}

// NewValidationService создает новый сервис валидации
func NewValidationService(errorHandler *errors.ErrorHandler) *ValidationService {
	return &ValidationService{
		userValidator:    NewUserValidator(),
		messageValidator: NewMessageValidator(),
		errorHandler:     errorHandler,
	}
}

// ValidateUserWithErrorHandling валидирует пользователя с обработкой ошибок
func (vs *ValidationService) ValidateUserWithErrorHandling(user *models.User, userID, chatID int64, operation string) error {
	result := vs.userValidator.ValidateUser(user)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации пользователя",
			"Проверьте введенные данные",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateUserRegistrationWithErrorHandling валидирует регистрацию пользователя
func (vs *ValidationService) ValidateUserRegistrationWithErrorHandling(telegramID int, username, firstName, languageCode string, userID, chatID int64, operation string) error {
	result := vs.userValidator.ValidateUserRegistration(telegramID, username, firstName, languageCode)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации при регистрации",
			"Проверьте введенные данные",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateUserUpdateWithErrorHandling валидирует обновление пользователя
func (vs *ValidationService) ValidateUserUpdateWithErrorHandling(user *models.User, userID, chatID int64, operation string) error {
	result := vs.userValidator.ValidateUserUpdate(user)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации при обновлении",
			"Проверьте введенные данные",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateUserInterestsWithErrorHandling валидирует интересы пользователя
func (vs *ValidationService) ValidateUserInterestsWithErrorHandling(interestIDs []int, userID, chatID int64, operation string) error {
	result := vs.userValidator.ValidateUserInterests(interestIDs)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации интересов",
			"Проверьте выбранные интересы",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateUserLanguagesWithErrorHandling валидирует языки пользователя
func (vs *ValidationService) ValidateUserLanguagesWithErrorHandling(nativeLanguage, targetLanguage string, userID, chatID int64, operation string) error {
	result := vs.userValidator.ValidateUserLanguages(nativeLanguage, targetLanguage)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации языков",
			"Проверьте выбранные языки",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateMessageWithErrorHandling валидирует сообщение с обработкой ошибок
func (vs *ValidationService) ValidateMessageWithErrorHandling(chatID, userID int64, text, operation string) error {
	result := vs.messageValidator.ValidateMessage(chatID, userID, text)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации сообщения",
			"Проверьте введенные данные",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateCallbackQueryWithErrorHandling валидирует callback query с обработкой ошибок
func (vs *ValidationService) ValidateCallbackQueryWithErrorHandling(chatID, userID int64, data, operation string) error {
	result := vs.messageValidator.ValidateCallbackQuery(chatID, userID, data)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации callback query",
			"Проверьте данные запроса",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateFeedbackMessageWithErrorHandling валидирует сообщение отзыва с обработкой ошибок
func (vs *ValidationService) ValidateFeedbackMessageWithErrorHandling(chatID, userID int64, text, operation string) error {
	result := vs.messageValidator.ValidateFeedbackMessage(chatID, userID, text)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации отзыва",
			"Проверьте текст отзыва",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateCommandWithErrorHandling валидирует команду с обработкой ошибок
func (vs *ValidationService) ValidateCommandWithErrorHandling(chatID, userID int64, command, operation string) error {
	result := vs.messageValidator.ValidateCommand(chatID, userID, command)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации команды",
			"Проверьте команду",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateLanguageSelectionWithErrorHandling валидирует выбор языка с обработкой ошибок
func (vs *ValidationService) ValidateLanguageSelectionWithErrorHandling(chatID, userID int64, languageCode, operation string) error {
	result := vs.messageValidator.ValidateLanguageSelection(chatID, userID, languageCode)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации выбора языка",
			"Проверьте выбранный язык",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateInterestSelectionWithErrorHandling валидирует выбор интереса с обработкой ошибок
func (vs *ValidationService) ValidateInterestSelectionWithErrorHandling(chatID, userID int64, interestID int, operation string) error {
	result := vs.messageValidator.ValidateInterestSelection(chatID, userID, interestID)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации выбора интереса",
			"Проверьте выбранный интерес",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}

// ValidateLanguageLevelSelectionWithErrorHandling валидирует выбор уровня языка с обработкой ошибок
func (vs *ValidationService) ValidateLanguageLevelSelectionWithErrorHandling(chatID, userID int64, level int, operation string) error {
	result := vs.messageValidator.ValidateLanguageLevelSelection(chatID, userID, level)

	if result.HasErrors() {
		// Создаем ошибку валидации с контекстом
		ctx := errors.NewRequestContext(userID, chatID, operation)
		validationErr := errors.NewValidationError(
			"Ошибка валидации выбора уровня языка",
			"Проверьте выбранный уровень",
			ctx,
		)

		// Добавляем детали ошибок валидации в контекст
		for field, fieldErrors := range result.GetErrors() {
			validationErr.Context[field] = fieldErrors
		}

		return vs.errorHandler.Handle(validationErr, ctx)
	}

	return nil
}
