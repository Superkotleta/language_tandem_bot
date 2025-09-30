package validation

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Константы для валидации
const (
	// minTelegramID - минимальный ID пользователя Telegram (обычно больше 100000000)
	minTelegramID = 100000000

	// maxUsernameLength - максимальная длина имени пользователя
	maxUsernameLength = 50

	// maxBioLength - максимальная длина биографии пользователя
	maxBioLength = 500

	// maxContactInfoLength - максимальная длина контактной информации
	maxContactInfoLength = 64

	// maxCommandLength - максимальная длина команды
	maxCommandLength = 32
)

// ValidationRule представляет правило валидации
type ValidationRule struct {
	Field    string
	Value    interface{}
	Rules    []string
	Messages map[string]string
}

// ValidationResult содержит результат валидации
type ValidationResult struct {
	IsValid bool
	Errors  map[string][]string
}

// NewValidationResult создает новый результат валидации
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		IsValid: true,
		Errors:  make(map[string][]string),
	}
}

// AddError добавляет ошибку валидации
func (vr *ValidationResult) AddError(field, message string) {
	vr.IsValid = false
	vr.Errors[field] = append(vr.Errors[field], message)
}

// GetErrors возвращает все ошибки валидации
func (vr *ValidationResult) GetErrors() map[string][]string {
	return vr.Errors
}

// HasErrors проверяет, есть ли ошибки
func (vr *ValidationResult) HasErrors() bool {
	return !vr.IsValid
}

// Validator содержит методы валидации
type Validator struct{}

// NewValidator создает новый валидатор
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateString валидирует строковое значение
func (v *Validator) ValidateString(value string, rules []string) []string {
	var errors []string

	for _, rule := range rules {
		switch rule {
		case "required":
			if strings.TrimSpace(value) == "" {
				errors = append(errors, "Поле обязательно для заполнения")
			}
		case "min:3":
			if utf8.RuneCountInString(value) < 3 {
				errors = append(errors, "Минимум 3 символа")
			}
		case "max:50":
			if utf8.RuneCountInString(value) > maxUsernameLength {
				errors = append(errors, "Максимум 50 символов")
			}
		case "max:100":
			if utf8.RuneCountInString(value) > 100 {
				errors = append(errors, "Максимум 100 символов")
			}
		case "max:500":
			if utf8.RuneCountInString(value) > maxBioLength {
				errors = append(errors, "Максимум 500 символов")
			}
		case "alphanumeric":
			if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value) {
				errors = append(errors, "Только буквы и цифры")
			}
		case "username":
			if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(value) {
				errors = append(errors, "Некорректный username")
			}
		case "email":
			if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(value) {
				errors = append(errors, "Некорректный email")
			}
		}
	}

	return errors
}

// ValidateInt валидирует целочисленное значение
func (v *Validator) ValidateInt(value int, rules []string) []string {
	var errors []string

	for _, rule := range rules {
		switch rule {
		case "required":
			if value == 0 {
				errors = append(errors, "Значение обязательно")
			}
		case "min:1":
			if value < 1 {
				errors = append(errors, "Минимум 1")
			}
		case "max:100":
			if value > 100 {
				errors = append(errors, "Максимум 100")
			}
		case "positive":
			if value <= 0 {
				errors = append(errors, "Должно быть положительным")
			}
		}
	}

	return errors
}

// ValidateLanguageCode валидирует код языка
func (v *Validator) ValidateLanguageCode(code string) []string {
	var errors []string

	if strings.TrimSpace(code) == "" {
		errors = append(errors, "Код языка обязателен")
		return errors
	}

	// Проверяем формат кода языка (2 символа)
	if len(code) != 2 {
		errors = append(errors, "Код языка должен содержать 2 символа")
	}

	// Проверяем, что код содержит только буквы
	if !regexp.MustCompile(`^[a-zA-Z]{2}$`).MatchString(code) {
		errors = append(errors, "Код языка должен содержать только буквы")
	}

	return errors
}

// ValidateTelegramID валидирует Telegram ID
func (v *Validator) ValidateTelegramID(id int64) []string {
	var errors []string

	if id <= 0 {
		errors = append(errors, "Telegram ID должен быть положительным")
	}

	// Telegram ID обычно больше minTelegramID
	if id < minTelegramID {
		errors = append(errors, "Некорректный Telegram ID")
	}

	return errors
}

// ValidateChatID валидирует Chat ID
func (v *Validator) ValidateChatID(id int64) []string {
	var errors []string

	if id == 0 {
		errors = append(errors, "Chat ID обязателен")
		return errors
	}

	// Chat ID может быть отрицательным для групп
	if id > 0 && id < 100000000 {
		errors = append(errors, "Некорректный Chat ID")
	}

	return errors
}

// ValidateUserState валидирует состояние пользователя
func (v *Validator) ValidateUserState(state string) []string {
	var errors []string

	validStates := []string{
		"idle", "setting_language", "setting_native_language",
		"setting_target_language", "setting_interests", "setting_profile",
		"waiting_for_feedback", "viewing_profile", "editing_profile",
	}

	isValid := false

	for _, validState := range validStates {
		if state == validState {
			isValid = true
			break
		}
	}

	if !isValid {
		errors = append(errors, "Некорректное состояние пользователя: "+state)
	}

	return errors
}

// ValidateInterestID валидирует ID интереса
func (v *Validator) ValidateInterestID(id int) []string {
	var errors []string

	if id <= 0 {
		errors = append(errors, "ID интереса должен быть положительным")
	}

	return errors
}

// ValidateLanguageLevel валидирует уровень языка
func (v *Validator) ValidateLanguageLevel(level int) []string {
	var errors []string

	if level < 1 || level > 5 {
		errors = append(errors, "Уровень языка должен быть от 1 до 5")
	}

	return errors
}

// ValidateFeedbackText валидирует текст отзыва
func (v *Validator) ValidateFeedbackText(text string) []string {
	var errors []string

	if strings.TrimSpace(text) == "" {
		errors = append(errors, "Текст отзыва обязателен")
		return errors
	}

	textLength := utf8.RuneCountInString(strings.TrimSpace(text))
	if textLength < 10 {
		errors = append(errors, "Минимум 10 символов")
	}

	if textLength > 1000 {
		errors = append(errors, "Максимум 1000 символов")
	}

	return errors
}

// ValidateCallbackData валидирует данные callback query
func (v *Validator) ValidateCallbackData(data string) []string {
	var errors []string

	if strings.TrimSpace(data) == "" {
		errors = append(errors, "Данные callback обязательны")
		return errors
	}

	if len(data) > maxContactInfoLength {
		errors = append(errors, "Максимум 64 символа")
	}

	// Проверяем, что данные содержат только разрешенные символы
	if !regexp.MustCompile(`^[a-zA-Z0-9_:.-]+$`).MatchString(data) {
		errors = append(errors, "Некорректные символы в данных")
	}

	return errors
}
