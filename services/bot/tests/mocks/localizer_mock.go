package mocks

import (
	"strings"
)

// LocalizerMock простой мок локализатора для тестов
type LocalizerMock struct {
	translations map[string]map[string]string
}

// NewLocalizerMock создает новый мок локализатора
func NewLocalizerMock() *LocalizerMock {
	return &LocalizerMock{
		translations: map[string]map[string]string{
			"en": {
				"welcome_message":         "👋 Hi, {name}! Welcome to Language Exchange Bot!",
				"choose_native_language":  "🌍 Choose your native language:",
				"choose_target_language":  "📚 What language are you learning?",
				"profile_summary_title":   "👤 Your profile",
				"profile_field_native":    "Native language",
				"profile_field_target":    "Learning language",
				"profile_field_interests": "Interests",
				"unknown_command":         "❓ Unknown command. Use /start to begin",
			},
			"ru": {
				"welcome_message":         "👋 Привет, {name}! Добро пожаловать в Language Exchange Bot!",
				"choose_native_language":  "🌍 Выбери свой родной язык:",
				"choose_target_language":  "📚 Какой язык ты изучаешь?",
				"profile_summary_title":   "👤 Твой профиль",
				"profile_field_native":    "Родной язык",
				"profile_field_target":    "Изучаемый язык",
				"profile_field_interests": "Интересы",
				"unknown_command":         "❓ Неизвестная команда. Используй /start для начала",
			},
		},
	}
}

// Get возвращает перевод для ключа и языка
func (l *LocalizerMock) Get(langCode, key string) string {
	if lang, exists := l.translations[langCode]; exists {
		if value, exists := lang[key]; exists {
			return value
		}
	}

	// Fallback на английский
	if lang, exists := l.translations["en"]; exists {
		if value, exists := lang[key]; exists {
			return value
		}
	}

	// Если ничего не найдено, возвращаем ключ
	return key
}

// GetWithParams возвращает перевод с заменой параметров
func (l *LocalizerMock) GetWithParams(langCode, key string, params map[string]string) string {
	text := l.Get(langCode, key)

	// Заменяем параметры в тексте
	for paramKey, paramValue := range params {
		placeholder := "{" + paramKey + "}"
		text = strings.ReplaceAll(text, placeholder, paramValue)
	}

	return text
}

// GetLanguageName возвращает название языка
func (l *LocalizerMock) GetLanguageName(langCode, interfaceLangCode string) string {
	names := map[string]map[string]string{
		"en": {"en": "English", "ru": "Russian", "es": "Spanish", "zh": "Chinese"},
		"ru": {"en": "Английский", "ru": "Русский", "es": "Испанский", "zh": "Китайский"},
		"es": {"en": "Inglés", "ru": "Ruso", "es": "Español", "zh": "Chino"},
		"zh": {"en": "英语", "ru": "俄语", "es": "西班牙语", "zh": "中文"},
	}

	if lang, exists := names[interfaceLangCode]; exists {
		if name, exists := lang[langCode]; exists {
			return name
		}
	}

	// Fallback
	return langCode
}

// GetInterests возвращает список интересов (заглушка для тестов)
func (l *LocalizerMock) GetInterests(langCode string) (map[int]string, error) {
	interests := map[int]string{
		1: "Movies",
		2: "Music",
		3: "Sports",
		4: "Travel",
		5: "Technology",
		6: "Food",
	}

	if langCode == "ru" {
		interests = map[int]string{
			1: "Фильмы",
			2: "Музыка",
			3: "Спорт",
			4: "Путешествия",
			5: "Технологии",
			6: "Еда",
		}
	}

	return interests, nil
}
