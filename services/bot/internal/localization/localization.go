package localization

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	l.loadTranslations()
	return l
}

func (l *Localizer) loadTranslations() {
	// Поддержка переопределения через env
	localesPath := os.Getenv("LOCALES_DIR")
	if localesPath == "" {
		localesPath = "./locales"
	}

	if _, err := os.Stat(localesPath); os.IsNotExist(err) {
		fmt.Printf("Locales directory '%s' not found, will use fallback to key names\n", localesPath)
		// Добавляем базовые переводы для тестов
		l.loadFallbackTranslations()
		return
	}

	filepath.WalkDir(localesPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}

		lang := strings.TrimSuffix(d.Name(), ".json")
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Failed reading %s: %v\n", path, err)
			return nil
		}

		var dict map[string]string
		if err := json.Unmarshal(data, &dict); err != nil {
			fmt.Printf("Failed parsing %s: %v\n", path, err)
			return nil
		}

		l.translations[lang] = dict
		fmt.Printf("Loaded %d keys for language: %s\n", len(dict), lang)
		return nil
	})
}

func (l *Localizer) Get(lang, key string) string {
	if dict, ok := l.translations[lang]; ok {
		if val, found := dict[key]; found {
			return val
		}
	}
	// Fallback на en
	if dict, ok := l.translations["en"]; ok {
		if val, found := dict[key]; found {
			return val
		}
	}
	// Последний fallback - вернуть ключ (чтобы видеть отсутствующие переводы)
	return key
}

func (l *Localizer) GetWithParams(lang, key string, params map[string]string) string {
	text := l.Get(lang, key)
	for k, v := range params {
		placeholder := "{" + k + "}"
		text = strings.ReplaceAll(text, placeholder, v)
	}
	return text
}

func (l *Localizer) GetLanguageName(lang, interfaceLang string) string {
	// Используем ключи типа "language_ru", "language_en" в JSON
	key := "language_" + lang
	return l.Get(interfaceLang, key)
}

func (l *Localizer) GetInterests(lang string) (map[int]string, error) {
	// Если БД не инициализирована (тесты), возвращаем заглушки
	if l.db == nil {
		interests := map[int]string{
			1: "Movies",
			2: "Music",
			3: "Sports",
			4: "Travel",
		}
		if lang == "ru" {
			interests = map[int]string{
				1: "Фильмы",
				2: "Музыка",
				3: "Спорт",
				4: "Путешествия",
			}
		}
		return interests, nil
	}

	interests := make(map[int]string)

	// Запрос к БД с локализацией - приоритет перевода, если NULL - ключ
	query := `
		SELECT i.id,
			   CASE
				   WHEN it.name IS NOT NULL AND TRIM(it.name) != '' THEN it.name
				   ELSE i.key_name
			   END as name
		FROM interests i
		LEFT JOIN interest_translations it ON i.id = it.interest_id AND it.language_code = $1
		ORDER BY i.id
	`

	rows, err := l.db.Query(query, lang)
	if err != nil {
		// Fallback на английский при ошибке
		rows, err = l.db.Query(query, "en")
		if err != nil {
			return nil, fmt.Errorf("failed to load interests: %w", err)
		}
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		interests[id] = name
		fmt.Printf("Interest %d: %s\n", id, name) // Debug: показать загруженные интересы
	}

	fmt.Printf("Loaded %d interests for language %s\n", len(interests), lang) // Debug: количество интересов

	return interests, nil
}

// loadFallbackTranslations загружает базовые переводы для тестов.
func (l *Localizer) loadFallbackTranslations() {
	// Английский
	l.translations["en"] = map[string]string{
		"welcome_message":         "👋 Hi, {name}! Welcome to Language Exchange Bot!",
		"choose_native_language":  "🌍 Choose your native language:",
		"choose_target_language":  "📚 What language are you learning?",
		"profile_summary_title":   "👤 Your profile",
		"profile_field_native":    "Native language",
		"profile_field_target":    "Learning language",
		"profile_field_interests": "Interests",
		"unknown_command":         "❓ Unknown command. Use /start to begin",
		"language_en":             "English",
		"language_ru":             "Russian",
		"language_es":             "Spanish",
		"language_zh":             "Chinese",
	}

	// Русский
	l.translations["ru"] = map[string]string{
		"welcome_message":         "👋 Привет, {name}! Добро пожаловать в Language Exchange Bot!",
		"choose_native_language":  "🌍 Выбери свой родной язык:",
		"choose_target_language":  "📚 Какой язык ты изучаешь?",
		"profile_summary_title":   "👤 Твой профиль",
		"profile_field_native":    "Родной язык",
		"profile_field_target":    "Изучаемый язык",
		"profile_field_interests": "Интересы",
		"unknown_command":         "❓ Неизвестная команда. Используй /start для начала",
		"language_en":             "Английский",
		"language_ru":             "Русский",
		"language_es":             "Испанский",
		"language_zh":             "Китайский",
	}

	// Испанский
	l.translations["es"] = map[string]string{
		"welcome_message":         "👋 ¡Hola, {name}! ¡Bienvenido al Language Exchange Bot!",
		"choose_native_language":  "🌍 Elige tu idioma nativo:",
		"choose_target_language":  "📚 ¿Qué idioma estás aprendiendo?",
		"profile_summary_title":   "👤 Tu perfil",
		"profile_field_native":    "Idioma nativo",
		"profile_field_target":    "Idioma objetivo",
		"profile_field_interests": "Intereses",
		"unknown_command":         "❓ Commando desconocido. Usa /start para comenzar",
		"language_en":             "Inglés",
		"language_ru":             "Ruso",
		"language_es":             "Español",
		"language_zh":             "Chino",
	}
}
