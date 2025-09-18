package mocks

import (
	"strings"
)

// LocalizerMock мок для локализатора
type LocalizerMock struct {
	translations map[string]map[string]string
}

// NewLocalizerMock создает новый мок локализатора
func NewLocalizerMock() *LocalizerMock {
	mock := &LocalizerMock{
		translations: make(map[string]map[string]string),
	}

	// Предзаполняем базовыми переводами
	mock.seedTranslations()

	return mock
}

// seedTranslations добавляет базовые переводы
func (l *LocalizerMock) seedTranslations() {
	// Английские переводы
	l.translations["en"] = map[string]string{
		"welcome_message":         "Welcome, {name}!",
		"choose_native_language":  "🌍 Choose your native language:",
		"choose_target_language":  "📚 What language are you learning?",
		"profile_summary_title":   "👤 Your profile",
		"profile_field_native":    "Native language",
		"profile_field_target":    "Target language",
		"profile_field_interests": "Interests",
		"language_en":             "English",
		"language_ru":             "Russian",
		"language_es":             "Spanish",
		"language_zh":             "Chinese",
		"interest_movies":         "Movies",
		"interest_music":          "Music",
		"interest_sports":         "Sports",
		"interest_travel":         "Travel",
		"interest_technology":     "Technology",
		"interest_food":           "Food",
	}

	// Русские переводы
	l.translations["ru"] = map[string]string{
		"welcome_message":         "Добро пожаловать, {name}!",
		"choose_native_language":  "🌍 Выберите ваш родной язык:",
		"choose_target_language":  "📚 Какой язык вы изучаете?",
		"profile_summary_title":   "👤 Твой профиль",
		"profile_field_native":    "Родной язык",
		"profile_field_target":    "Изучаемый язык",
		"profile_field_interests": "Интересы",
		"language_en":             "Английский",
		"language_ru":             "Русский",
		"language_es":             "Испанский",
		"language_zh":             "Китайский",
		"interest_movies":         "Фильмы",
		"interest_music":          "Музыка",
		"interest_sports":         "Спорт",
		"interest_travel":         "Путешествия",
		"interest_technology":     "Технологии",
		"interest_food":           "Еда",
	}

	// Испанские переводы
	l.translations["es"] = map[string]string{
		"welcome_message":         "¡Bienvenido, {name}!",
		"choose_native_language":  "🌍 Elige tu idioma nativo:",
		"choose_target_language":  "📚 ¿Qué idioma estás aprendiendo?",
		"profile_summary_title":   "👤 Tu perfil",
		"profile_field_native":    "Idioma nativo",
		"profile_field_target":    "Idioma objetivo",
		"profile_field_interests": "Intereses",
		"language_en":             "Inglés",
		"language_ru":             "Ruso",
		"language_es":             "Español",
		"language_zh":             "Chino",
		"interest_movies":         "Películas",
		"interest_music":          "Música",
		"interest_sports":         "Deportes",
		"interest_travel":         "Viajes",
		"interest_technology":     "Tecnología",
		"interest_food":           "Comida",
	}

	// Китайские переводы
	l.translations["zh"] = map[string]string{
		"welcome_message":         "欢迎，{name}！",
		"choose_native_language":  "🌍 选择您的母语：",
		"choose_target_language":  "📚 您在学习什么语言？",
		"profile_summary_title":   "👤 您的个人资料",
		"profile_field_native":    "母语",
		"profile_field_target":    "目标语言",
		"profile_field_interests": "兴趣",
		"language_en":             "英语",
		"language_ru":             "俄语",
		"language_es":             "西班牙语",
		"language_zh":             "中文",
		"interest_movies":         "电影",
		"interest_music":          "音乐",
		"interest_sports":         "运动",
		"interest_travel":         "旅行",
		"interest_technology":     "技术",
		"interest_food":           "食物",
	}
}

// Get возвращает перевод для ключа
func (l *LocalizerMock) Get(langCode, key string) string {
	if translations, exists := l.translations[langCode]; exists {
		if translation, exists := translations[key]; exists {
			return translation
		}
	}

	// Fallback на английский
	if translations, exists := l.translations["en"]; exists {
		if translation, exists := translations[key]; exists {
			return translation
		}
	}

	// Если ничего не найдено, возвращаем ключ
	return key
}

// GetWithParams возвращает перевод с параметрами
func (l *LocalizerMock) GetWithParams(langCode, key string, params map[string]string) string {
	text := l.Get(langCode, key)

	// Простая замена параметров
	for param, value := range params {
		text = strings.ReplaceAll(text, "{"+param+"}", value)
	}

	return text
}

// GetLanguageName возвращает название языка
func (l *LocalizerMock) GetLanguageName(langCode, interfaceLangCode string) string {
	key := "language_" + langCode
	return l.Get(interfaceLangCode, key)
}

// GetInterests возвращает интересы для языка
func (l *LocalizerMock) GetInterests(langCode string) (map[int]string, error) {
	interests := map[int]string{
		1: l.Get(langCode, "interest_movies"),
		2: l.Get(langCode, "interest_music"),
		3: l.Get(langCode, "interest_sports"),
		4: l.Get(langCode, "interest_travel"),
		5: l.Get(langCode, "interest_technology"),
		6: l.Get(langCode, "interest_food"),
	}
	return interests, nil
}
