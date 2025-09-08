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
	interests := make(map[int]string)

	// Запрос к БД с локализацией
	query := `
		SELECT i.id, COALESCE(it.name, i.key_name) as name
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
	}

	return interests, nil
}
