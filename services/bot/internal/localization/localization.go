package localization

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Localizer struct {
	db           *sql.DB
	translations map[string]map[string]string
}

func NewLocalizer(db *sql.DB) *Localizer {
	l := &Localizer{
		db:           db,
		translations: make(map[string]map[string]string),
	}

	// Загружаем переводы из JSON файлов
	l.loadTranslations()
	return l
}

func (l *Localizer) loadTranslations() {
	// Путь к файлам переводов
	localesPath := "./locales"

	// Проверяем, существует ли папка
	if _, err := os.Stat(localesPath); os.IsNotExist(err) {
		// Если папки нет, используем встроенные переводы
		l.loadDefaultTranslations()
		return
	}

	// Читаем все JSON файлы в папке locales
	filepath.WalkDir(localesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			// Извлекаем код языка из имени файла
			langCode := strings.TrimSuffix(d.Name(), ".json")

			// Читаем файл
			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", path, err)
				return nil
			}

			// Парсим JSON
			var translations map[string]string
			if err := json.Unmarshal(data, &translations); err != nil {
				fmt.Printf("Error parsing %s: %v\n", path, err)
				return nil
			}

			l.translations[langCode] = translations
			fmt.Printf("Loaded translations for %s (%d keys)\n", langCode, len(translations))
		}

		return nil
	})
}

func (l *Localizer) loadDefaultTranslations() {
	// Встроенные переводы на случай, если JSON файлы не найдены
	l.translations = map[string]map[string]string{
		"en": {
			"welcome_message":           "👋 Hi, {name}! Welcome to Language Exchange Bot!\n\n🌟 Here you'll find conversation partners for language learning!\n\n🚀 Let's set up your profile:",
			"choose_native_language":    "🌍 Choose your native language:",
			"choose_target_language":    "📚 What language are you learning?",
			"choose_interface_language": "🌐 Choose interface language:",
			"choose_interests":          "🎯 Choose your interests (you can select multiple):",
			"profile_completed":         "✅ Great! Your profile is ready!\n\n🔍 Starting to search for suitable conversation partners...",
			"language_updated":          "✅ Interface language updated!",
			"interest_added":            "✅ Interest added to your profile!",
			"use_menu_above":            "👆 Please use the menu above",
			"unknown_command":           "❓ Unknown command. Use /start to begin",
			// добавлено:
			"your_status":               "Your status",
			"status":                    "Status",
			"state":                     "State",
			"profile_completion":        "Profile completion",
			"interface_language":        "Interface language",
			"profile_reset":             "Profile reset",
			"native_language_confirmed": "Native language",
			"target_language_confirmed": "Learning language",

			"profile_summary_title":   "👤 Your profile",
			"profile_field_native":    "Native language",
			"profile_field_target":    "Learning language",
			"profile_field_interests": "Interests",
			"profile_actions":         "What would you like to do?",
			"profile_show":            "👤 Show profile",
			"profile_reconfigure":     "🔁 Reconfigure",
			"profile_reset_title":     "♻️ Reset profile",
			"profile_reset_warning":   "This will clear your languages and interests. Are you sure?",
			"profile_reset_yes":       "✅ Yes, reset",
			"profile_reset_no":        "❌ No, keep it",
			"profile_reset_done":      "✅ Profile has been reset. Let's start again!",
		},
		"ru": {
			"welcome_message":           "👋 Привет, {name}! Добро пожаловать в Language Exchange Bot!\n\n🌟 Здесь ты найдешь собеседников для изучения языков!\n\n🚀 Давай настроим твой профиль:",
			"choose_native_language":    "🌍 Выбери свой родной язык:",
			"choose_target_language":    "📚 Какой язык ты изучаешь?",
			"choose_interface_language": "🌐 Выбери язык интерфейса:",
			"choose_interests":          "🎯 Выбери свои интересы (можно несколько):",
			"profile_completed":         "✅ Отлично! Твой профиль готов!\n\n🔍 Начинаю поиск подходящих собеседников...",
			"language_updated":          "✅ Язык интерфейса обновлен!",
			"interest_added":            "✅ Интерес добавлен к твоему профилю!",
			"use_menu_above":            "👆 Пожалуйста, используй меню выше",
			"unknown_command":           "❓ Неизвестная команда. Используй /start для начала",
			// добавлено:
			"your_status":               "Твой статус",
			"status":                    "Статус",
			"state":                     "Состояние",
			"profile_completion":        "Заполненность профиля",
			"interface_language":        "Язык интерфейса",
			"profile_reset":             "Сброс профиля",
			"native_language_confirmed": "Родной язык",
			"target_language_confirmed": "Изучаемый язык",

			"profile_summary_title":   "👤 Твой профиль",
			"profile_field_native":    "Родной язык",
			"profile_field_target":    "Изучаемый язык",
			"profile_field_interests": "Интересы",
			"profile_actions":         "Что дальше сделать?",
			"profile_show":            "👤 Показать профиль",
			"profile_reconfigure":     "🔁 Перезаполнить",
			"profile_reset_title":     "♻️ Сброс профиля",
			"profile_reset_warning":   "Будут очищены языки и интересы. Точно продолжить?",
			"profile_reset_yes":       "✅ Да, сбросить",
			"profile_reset_no":        "❌ Нет, оставить",
			"profile_reset_done":      "✅ Профиль сброшен. Начнём заново!",
		},
	}
	fmt.Println("Loaded default translations")
}

func (l *Localizer) Get(langCode, key string) string {
	if langMap, ok := l.translations[langCode]; ok {
		if text, exists := langMap[key]; exists {
			return text
		}
	}

	// Возвращаем английский по умолчанию
	if langMap, ok := l.translations["en"]; ok {
		if text, exists := langMap[key]; exists {
			return text
		}
	}

	return key
}

func (l *Localizer) GetWithParams(langCode, key string, params map[string]string) string {
	text := l.Get(langCode, key)

	// Замена параметров в стиле {name}
	for k, v := range params {
		placeholder := "{" + k + "}"
		text = strings.ReplaceAll(text, placeholder, v)
	}

	return text
}

func (l *Localizer) GetLanguageName(langCode, interfaceLangCode string) string {
	names := map[string]map[string]string{
		"en": {
			"ru": "Russian", "en": "English", "es": "Spanish", "zh": "Chinese",
		},
		"ru": {
			"ru": "Русский", "en": "Английский", "es": "Испанский", "zh": "Китайский",
		},
		"es": {
			"ru": "Ruso", "en": "Inglés", "es": "Español", "zh": "Chino",
		},
		"zh": {
			"ru": "俄语", "en": "英语", "es": "西班牙语", "zh": "中文",
		},
	}

	if langMap, ok := names[interfaceLangCode]; ok {
		if name, exists := langMap[langCode]; exists {
			return name
		}
	}

	return langCode
}

func (l *Localizer) GetInterests(langCode string) (map[int]string, error) {
	interests := map[string]map[int]string{
		"en": {
			1: "🎬 Movies & TV Shows", 2: "📚 Books & Literature",
			3: "🏃 Sports & Fitness", 4: "🎵 Music", 5: "🍳 Cooking & Food",
			6: "✈️ Travel & Culture", 7: "💻 Technology", 8: "🎨 Art & Design",
			9: "🎮 Gaming", 10: "📖 Education",
		},
		"ru": {
			1: "🎬 Фильмы и сериалы", 2: "📚 Книги и литература",
			3: "🏃 Спорт и фитнес", 4: "🎵 Музыка", 5: "🍳 Кулинария и еда",
			6: "✈️ Путешествия и культура", 7: "💻 Технологии", 8: "🎨 Искусство и дизайн",
			9: "🎮 Игры", 10: "📖 Образование",
		},
	}

	if langInterests, ok := interests[langCode]; ok {
		return langInterests, nil
	}

	return interests["en"], nil
}
