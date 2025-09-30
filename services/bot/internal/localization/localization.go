// Package localization provides internationalization and translation functionality.
package localization

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Localizer предоставляет функциональность локализации.
type Localizer struct {
	db           *sql.DB
	translations map[string]map[string]string
}

// NewLocalizer создает новый экземпляр Localizer.
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

	err := filepath.WalkDir(localesPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}

		lang := strings.TrimSuffix(d.Name(), ".json")

		cleanPath := filepath.Clean(path)

		if strings.Contains(cleanPath, "..") || strings.Contains(cleanPath, "~") {
			fmt.Printf("Небезопасный путь к файлу: %s\n", path)

			return nil
		}

		data, err := os.ReadFile(cleanPath)
		if err != nil {
			fmt.Printf("Failed reading %s: %v\n", cleanPath, err)

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
	if err != nil {
		fmt.Printf("Error walking locales directory: %v\n", err)
		l.loadFallbackTranslations()
	}
}

// Get возвращает локализованную строку по ключу.
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

// GetWithParams возвращает локализованную строку с подстановкой параметров.
func (l *Localizer) GetWithParams(lang, key string, params map[string]string) string {
	text := l.Get(lang, key)

	for k, v := range params {
		placeholder := "{" + k + "}"
		text = strings.ReplaceAll(text, placeholder, v)
	}

	return text
}

// GetLanguageName возвращает локализованное название языка.
func (l *Localizer) GetLanguageName(lang, interfaceLang string) string {
	// Используем ключи типа "language_ru", "language_en" в JSON
	key := "language_" + lang

	return l.Get(interfaceLang, key)
}

// GetInterests возвращает локализованные интересы для указанного языка.
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

	rows, err := l.db.QueryContext(context.Background(), query, lang)
	if err != nil {
		// Fallback на английский при ошибке
		rows, err = l.db.QueryContext(context.Background(), query, "en")
		if err != nil {
			return nil, fmt.Errorf("failed to load interests: %w", err)
		}
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			// Логируем ошибку закрытия, но не возвращаем её
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	for rows.Next() {
		var id int

		var name string

		err := rows.Scan(&id, &name)
		if err != nil {
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
	}
}
