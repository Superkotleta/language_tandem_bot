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

	errorsPkg "language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/logging"
)

// Localizer предоставляет функциональность локализации.
type Localizer struct {
	db           *sql.DB
	translations map[string]map[string]string
	logger       *logging.ComponentLogger
	errorHandler *errorsPkg.ErrorHandler
}

// NewLocalizer создает новый экземпляр Localizer.
func NewLocalizer(db *sql.DB) *Localizer {
	localizer := &Localizer{
		db:           db,
		translations: make(map[string]map[string]string),
		logger:       logging.NewComponentLogger("localization"),
		errorHandler: errorsPkg.NewErrorHandler(nil),
	}
	localizer.loadTranslations()

	return localizer
}

func (l *Localizer) loadTranslations() {
	localesPath := l.getLocalesPath()

	if !l.localesDirectoryExists(localesPath) {
		l.loadFallbackTranslations()

		return
	}

	l.walkLocalesDirectory(localesPath)
}

// getLocalesPath возвращает путь к директории с переводами.
func (l *Localizer) getLocalesPath() string {
	localesPath := os.Getenv("LOCALES_DIR")
	if localesPath == "" {
		localesPath = "./locales"
	}

	return localesPath
}

// localesDirectoryExists проверяет существование директории с переводами.
func (l *Localizer) localesDirectoryExists(localesPath string) bool {
	if _, err := os.Stat(localesPath); os.IsNotExist(err) {
		l.logger.WarnWithContext(
			"Locales directory not found, using fallback translations",
			"", 0, 0, "LoadTranslations",
			map[string]interface{}{
				"locales_path": localesPath,
				"error":        err.Error(),
			},
		)

		return false
	}

	return true
}

// walkLocalesDirectory обходит директорию с переводами.
func (l *Localizer) walkLocalesDirectory(localesPath string) {
	err := filepath.WalkDir(localesPath, l.processLocaleFile)
	if err != nil {
		l.logger.ErrorWithContext(
			"Failed to walk locales directory, using fallback translations",
			"", 0, 0, "LoadTranslations",
			map[string]interface{}{
				"locales_path": localesPath,
				"error":        err.Error(),
			},
		)
		l.loadFallbackTranslations()
	}
}

// processLocaleFile обрабатывает один файл перевода.
func (l *Localizer) processLocaleFile(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
		return nil
	}

	lang := strings.TrimSuffix(d.Name(), ".json")
	cleanPath := filepath.Clean(path)

	if !l.isPathSafe(cleanPath) {
		l.logger.WarnWithContext(
			"Unsafe file path detected, skipping file",
			"", 0, 0, "ProcessLocaleFile",
			map[string]interface{}{
				"file_path": cleanPath,
				"language":  lang,
			},
		)

		return nil
	}

	return l.loadLocaleFile(cleanPath, lang)
}

// isPathSafe проверяет безопасность пути.
func (l *Localizer) isPathSafe(cleanPath string) bool {
	return !strings.Contains(cleanPath, "..") && !strings.Contains(cleanPath, "~")
}

// loadLocaleFile загружает файл перевода.
func (l *Localizer) loadLocaleFile(cleanPath, lang string) error {
	data, err := os.ReadFile(cleanPath) // #nosec G304 - путь проверен на безопасность
	if err != nil {
		l.logger.ErrorWithContext(
			"Failed to read locale file",
			"", 0, 0, "LoadLocaleFile",
			map[string]interface{}{
				"file_path": cleanPath,
				"language":  lang,
				"error":     err.Error(),
			},
		)

		return fmt.Errorf("failed to read file: %w", err)
	}

	var dict map[string]string
	if err := json.Unmarshal(data, &dict); err != nil {
		l.logger.ErrorWithContext(
			"Failed to parse locale file JSON",
			"", 0, 0, "LoadLocaleFile",
			map[string]interface{}{
				"file_path": cleanPath,
				"language":  lang,
				"error":     err.Error(),
			},
		)

		return fmt.Errorf("failed to unmarshal file: %w", err)
	}

	l.translations[lang] = dict
	l.logger.InfoWithContext(
		"Locale file loaded successfully",
		"", 0, 0, "LoadLocaleFile",
		map[string]interface{}{
			"language":   lang,
			"keys_count": len(dict),
			"file_path":  cleanPath,
		},
	)

	return nil
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
		return l.getFallbackInterests(lang), nil
	}

	return l.loadInterestsFromDB(lang)
}

// getFallbackInterests возвращает заглушки для тестов.
func (l *Localizer) getFallbackInterests(lang string) map[int]string {
	interests := map[int]string{
		1: "Movies",
		2: "Music",
		3: "Sports",
		4: "Travel",
	}

	// Для русского языка используем русские переводы
	if lang == "ru" {
		interests = map[int]string{
			1: "Фильмы",
			2: "Музыка",
			3: "Спорт",
			4: "Путешествия",
		}
	}

	// Для испанского языка
	if lang == "es" {
		interests = map[int]string{
			1: "Películas",
			2: "Música",
			3: "Deportes",
			4: "Viajes",
		}
	}

	// Для китайского языка
	if lang == "zh" {
		interests = map[int]string{
			1: "电影",
			2: "音乐",
			3: "运动",
			4: "旅行",
		}
	}

	return interests
}

// loadInterestsFromDB загружает интересы из базы данных.
func (l *Localizer) loadInterestsFromDB(lang string) (map[int]string, error) {
	interests := make(map[int]string)
	query := l.getInterestsQuery()

	rows, err := l.db.QueryContext(context.Background(), query, lang)
	if err != nil {
		// Fallback на английский при ошибке
		rows, err = l.db.QueryContext(context.Background(), query, "en")
		if err != nil {
			return nil, fmt.Errorf("failed to load interests: %w", err)
		}
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			l.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "LoadInterestsFromDB",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	l.scanInterestsRows(rows, interests)

	return interests, nil
}

// getInterestsQuery возвращает SQL запрос для получения интересов.
func (l *Localizer) getInterestsQuery() string {
	return `
		SELECT i.id,
			   CASE
				   WHEN it.name IS NOT NULL AND TRIM(it.name) != '' THEN it.name
				   ELSE i.key_name
			   END as name
		FROM interests i
		LEFT JOIN interest_translations it ON i.id = it.interest_id AND it.language_code = $1
		ORDER BY i.id
	`
}

// scanInterestsRows сканирует строки результата запроса интересов.
func (l *Localizer) scanInterestsRows(rows *sql.Rows, interests map[int]string) {
	for rows.Next() {
		var interestID int

		var name string

		err := rows.Scan(&interestID, &name)
		if err != nil {
			continue
		}

		interests[interestID] = name
	}

	l.logger.DebugWithContext(
		"Interests loaded from database",
		"", 0, 0, "ScanInterestsRows",
		map[string]interface{}{
			"interests_count": len(interests),
		},
	)
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
	// Инициализация базовых переводов для всех языков
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

	l.translations["en"] = map[string]string{
		"welcome_message":         "👋 Hi, {name}! Welcome to Language Exchange Bot!",
		"choose_native_language":  "🌍 Choose your native language:",
		"choose_target_language":  "📚 What language are you learning?",
		"profile_summary_title":   "👤 Your profile",
		"profile_field_native":    "Native language",
		"profile_field_target":    "Target language",
		"profile_field_interests": "Interests",
		"unknown_command":         "❓ Unknown command. Use /start to begin",
	}

	l.translations["es"] = map[string]string{
		"welcome_message":         "👋 ¡Hola, {name}! ¡Bienvenido a Language Exchange Bot!",
		"choose_native_language":  "🌍 Elige tu idioma nativo:",
		"choose_target_language":  "📚 ¿Qué idioma estás aprendiendo?",
		"profile_summary_title":   "👤 Tu perfil",
		"profile_field_native":    "Idioma nativo",
		"profile_field_target":    "Idioma de aprendizaje",
		"profile_field_interests": "Intereses",
		"unknown_command":         "❓ Commando desconocido. Usa /start para comenzar",
	}

	l.translations["zh"] = map[string]string{
		"welcome_message":         "👋 你好，{name}！欢迎使用语言交换机器人！",
		"choose_native_language":  "🌍 选择你的母语：",
		"choose_target_language":  "📚 你正在学习什么语言？",
		"profile_summary_title":   "👤 你的个人资料",
		"profile_field_native":    "母语",
		"profile_field_target":    "学习语言",
		"profile_field_interests": "兴趣",
		"unknown_command":         "❓ 未知命令。使用 /start 开始",
	}
}
