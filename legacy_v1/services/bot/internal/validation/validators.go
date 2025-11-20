package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"language-exchange-bot/internal/localization"
)

// Validation constants are now centralized in localization/constants.go

// Rule представляет правило валидации.
type Rule struct {
	Field    string
	Value    interface{}
	Rules    []string
	Messages map[string]string
}

// Result содержит результат валидации.
type Result struct {
	IsValid bool
	Errors  map[string][]string
}

// NewResult создает новый результат валидации.
func NewResult() *Result {
	return &Result{
		IsValid: true,
		Errors:  make(map[string][]string),
	}
}

// AddError добавляет ошибку валидации.
func (vr *Result) AddError(field, message string) {
	vr.IsValid = false
	vr.Errors[field] = append(vr.Errors[field], message)
}

// GetErrors возвращает все ошибки валидации.
func (vr *Result) GetErrors() map[string][]string {
	return vr.Errors
}

// HasErrors проверяет, есть ли ошибки.
func (vr *Result) HasErrors() bool {
	return !vr.IsValid
}

// Validator содержит методы валидации.
type Validator struct{}

// NewValidator создает новый валидатор.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateString валидирует строковое значение.
func (v *Validator) ValidateString(value string, rules []string) []string {
	var errors []string

	for _, rule := range rules {
		if err := v.validateStringRule(value, rule); err != "" {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateStringRule валидирует одно правило для строки.
func (v *Validator) validateStringRule(value, rule string) string {
	switch rule {
	case "required":
		return v.validateRequired(value)
	case "min:3":
		return v.validateMinLength(value, localization.MinStringLength)
	case "max:50":
		return v.validateMaxLength(value, localization.MaxUsernameLength, "50 символов")
	case "max:100":
		return v.validateMaxLength(value, localization.MaxStringLength, "100 символов")
	case "max:500":
		return v.validateMaxLength(value, localization.MaxBioLength, "500 символов")
	case "alphanumeric":
		return v.validatePattern(value, `^[a-zA-Z0-9]+$`, "Только буквы и цифры")
	case "username":
		return v.validatePattern(value, `^[a-zA-Z0-9_]+$`, "Некорректный username")
	case "email":
		return v.validatePattern(value, `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "Некорректный email")
	}

	return ""
}

// validateRequired проверяет обязательность поля.
func (v *Validator) validateRequired(value string) string {
	if strings.TrimSpace(value) == "" {
		return "Поле обязательно для заполнения"
	}

	return ""
}

// validateMinLength проверяет минимальную длину.
func (v *Validator) validateMinLength(value string, minLen int) string {
	if utf8.RuneCountInString(value) < minLen {
		return fmt.Sprintf("Минимум %d символов", minLen)
	}

	return ""
}

// validateMaxLength проверяет максимальную длину.
func (v *Validator) validateMaxLength(value string, maxLen int, message string) string {
	if utf8.RuneCountInString(value) > maxLen {
		return "Максимум " + message
	}

	return ""
}

// validatePattern проверяет соответствие паттерну.
func (v *Validator) validatePattern(value, pattern, message string) string {
	if !regexp.MustCompile(pattern).MatchString(value) {
		return message
	}

	return ""
}

// ValidateInt валидирует целочисленное значение.
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
			if value > localization.MaxStringLength {
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

// ValidateLanguageCode валидирует код языка.
func (v *Validator) ValidateLanguageCode(code string) []string {
	var errors []string

	if strings.TrimSpace(code) == "" {
		errors = append(errors, "Код языка обязателен")

		return errors
	}

	// Проверяем формат кода языка (2 символа)
	if len(code) != localization.LanguageCodeLength {
		errors = append(errors, "Код языка должен содержать 2 символа")
	}

	// Проверяем, что код содержит только буквы
	if !regexp.MustCompile(`^[a-zA-Z]{2}$`).MatchString(code) {
		errors = append(errors, "Код языка должен содержать только буквы")
	}

	return errors
}

// ValidateTelegramID валидирует Telegram ID.
func (v *Validator) ValidateTelegramID(telegramID int64) []string {
	var errors []string

	if telegramID <= 0 {
		errors = append(errors, "Telegram ID должен быть положительным")
	}

	// Telegram ID обычно больше MinTelegramID
	if telegramID < localization.MinTelegramID {
		errors = append(errors, "Некорректный Telegram ID")
	}

	return errors
}

// ValidateChatID валидирует Chat ID.
func (v *Validator) ValidateChatID(chatID int64) []string {
	var errors []string

	if chatID == 0 {
		errors = append(errors, "Chat ID обязателен")

		return errors
	}

	// Chat ID может быть отрицательным для групп
	if chatID > 0 && chatID < 100000000 {
		errors = append(errors, "Некорректный Chat ID")
	}

	return errors
}

// ValidateUserState валидирует состояние пользователя.
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

// ValidateInterestID валидирует ID интереса.
func (v *Validator) ValidateInterestID(id int) []string {
	var errors []string

	if id <= 0 {
		errors = append(errors, "ID интереса должен быть положительным")
	}

	return errors
}

// ValidateLanguageLevel валидирует уровень языка.
func (v *Validator) ValidateLanguageLevel(level int) []string {
	var errors []string

	if level < 1 || level > 5 {
		errors = append(errors, "Уровень языка должен быть от 1 до 5")
	}

	return errors
}

// ValidateFeedbackText валидирует текст отзыва.
func (v *Validator) ValidateFeedbackText(text string) []string {
	var errors []string

	if strings.TrimSpace(text) == "" {
		errors = append(errors, "Текст отзыва обязателен")

		return errors
	}

	textLength := utf8.RuneCountInString(strings.TrimSpace(text))
	if textLength < localization.MinTextLength {
		errors = append(errors, "Минимум 10 символов")
	}

	if textLength > localization.MaxTextLength {
		errors = append(errors, "Максимум 1000 символов")
	}

	return errors
}

// ValidateCallbackData валидирует данные callback query.
func (v *Validator) ValidateCallbackData(data string) []string {
	var errors []string

	if strings.TrimSpace(data) == "" {
		errors = append(errors, "Данные callback обязательны")

		return errors
	}

	if len(data) > localization.MaxContactInfoLength {
		errors = append(errors, "Максимум 64 символа")
	}

	// Проверяем, что данные содержат только разрешенные символы
	if !regexp.MustCompile(`^[a-zA-Z0-9_:.-]+$`).MatchString(data) {
		errors = append(errors, "Некорректные символы в данных")
	}

	return errors
}

// =============================================================================
// COMMON VALIDATION HELPERS - хелперы для типичных паттернов валидации
// =============================================================================
// Эти методы предоставляют быстрый доступ к часто используемым проверкам.
// Преимущества:
// 1. Сокращение дублирования кода
// 2. Стандартизация сообщений об ошибках
// 3. Быстрое создание валидации для новых полей

// ValidateNotEmpty проверяет, что значение не пустое (после trim).
func (v *Validator) ValidateNotEmpty(value string, fieldName string) string {
	if strings.TrimSpace(value) == "" {
		return fmt.Sprintf("Поле '%s' обязательно для заполнения", fieldName)
	}
	return ""
}

// ValidateLengthRange проверяет, что длина строки находится в заданном диапазоне.
func (v *Validator) ValidateLengthRange(value string, minLen, maxLen int, fieldName string) string {
	length := utf8.RuneCountInString(value)
	if length < minLen {
		return fmt.Sprintf("Поле '%s' должно содержать минимум %d символов", fieldName, minLen)
	}
	if length > maxLen {
		return fmt.Sprintf("Поле '%s' должно содержать максимум %d символов", fieldName, maxLen)
	}
	return ""
}

// ValidateMinLength проверяет минимальную длину с кастомным сообщением.
func (v *Validator) ValidateMinLength(value string, minLen int, fieldName string) string {
	if utf8.RuneCountInString(value) < minLen {
		return fmt.Sprintf("Поле '%s' должно содержать минимум %d символов", fieldName, minLen)
	}
	return ""
}

// ValidateMaxLength проверяет максимальную длину с кастомным сообщением.
func (v *Validator) ValidateMaxLength(value string, maxLen int, fieldName string) string {
	if utf8.RuneCountInString(value) > maxLen {
		return fmt.Sprintf("Поле '%s' должно содержать максимум %d символов", fieldName, maxLen)
	}
	return ""
}

// ValidatePositiveNumber проверяет, что число положительное.
func (v *Validator) ValidatePositiveNumber(value int, fieldName string) string {
	if value <= 0 {
		return fmt.Sprintf("Поле '%s' должно быть положительным числом", fieldName)
	}
	return ""
}

// ValidateRange проверяет, что число находится в заданном диапазоне.
func (v *Validator) ValidateRange(value, min, max int, fieldName string) string {
	if value < min {
		return fmt.Sprintf("Поле '%s' должно быть не меньше %d", fieldName, min)
	}
	if value > max {
		return fmt.Sprintf("Поле '%s' должно быть не больше %d", fieldName, max)
	}
	return ""
}

// ValidateOneOf проверяет, что значение входит в список допустимых.
func (v *Validator) ValidateOneOf(value string, allowed []string, fieldName string) string {
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return ""
		}
	}
	return fmt.Sprintf("Поле '%s' должно быть одним из: %s", fieldName, strings.Join(allowed, ", "))
}

// ValidateEmailFormat проверяет формат email адреса.
func (v *Validator) ValidateEmailFormat(email string, fieldName string) string {
	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email) {
		return fmt.Sprintf("Поле '%s' содержит некорректный email адрес", fieldName)
	}
	return ""
}

// ValidateURLFormat проверяет формат URL.
func (v *Validator) ValidateURLFormat(url string, fieldName string) string {
	if !regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`).MatchString(url) {
		return fmt.Sprintf("Поле '%s' содержит некорректный URL", fieldName)
	}
	return ""
}

// ValidateAlphanumeric проверяет, что строка содержит только буквы и цифры.
func (v *Validator) ValidateAlphanumeric(value string, fieldName string) string {
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value) {
		return fmt.Sprintf("Поле '%s' должно содержать только буквы и цифры", fieldName)
	}
	return ""
}

// ValidateUsernameFormat проверяет формат username.
func (v *Validator) ValidateUsernameFormat(username string, fieldName string) string {
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username) {
		return fmt.Sprintf("Поле '%s' содержит некорректные символы (только буквы, цифры и _)", fieldName)
	}
	return ""
}

// ValidatePhoneFormat проверяет формат телефонного номера (простая проверка).
func (v *Validator) ValidatePhoneFormat(phone string, fieldName string) string {
	// Удаляем все нецифровые символы для проверки
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return fmt.Sprintf("Поле '%s' содержит некорректный номер телефона", fieldName)
	}
	return ""
}

// ValidateNotNil проверяет, что указатель не nil (для interface{}).
func (v *Validator) ValidateNotNil(value interface{}, fieldName string) string {
	if value == nil {
		return fmt.Sprintf("Поле '%s' обязательно для заполнения", fieldName)
	}
	return ""
}

// ValidateSliceNotEmpty проверяет, что слайс не пустой.
func (v *Validator) ValidateSliceNotEmpty(slice interface{}, fieldName string) string {
	switch s := slice.(type) {
	case []string:
		if len(s) == 0 {
			return fmt.Sprintf("Поле '%s' должно содержать хотя бы один элемент", fieldName)
		}
	case []int:
		if len(s) == 0 {
			return fmt.Sprintf("Поле '%s' должно содержать хотя бы один элемент", fieldName)
		}
	case []interface{}:
		if len(s) == 0 {
			return fmt.Sprintf("Поле '%s' должно содержать хотя бы один элемент", fieldName)
		}
	default:
		return fmt.Sprintf("Неподдерживаемый тип для проверки '%s'", fieldName)
	}
	return ""
}
