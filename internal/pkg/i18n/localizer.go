package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Localizer handles multi-language support
type Localizer struct {
	translations map[string]map[string]string
	fallbackLang string
}

// NewLocalizer creates a new localizer instance
func NewLocalizer(localesPath, fallbackLang string) (*Localizer, error) {
	l := &Localizer{
		translations: make(map[string]map[string]string),
		fallbackLang: fallbackLang,
	}

	if err := l.loadTranslations(localesPath); err != nil {
		return nil, fmt.Errorf("failed to load translations: %w", err)
	}

	return l, nil
}

// loadTranslations reads all JSON files from localesPath
func (l *Localizer) loadTranslations(localesPath string) error {
	files, err := os.ReadDir(localesPath)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		lang := file.Name()[:len(file.Name())-5] // Remove .json extension
		filePath := filepath.Join(localesPath, file.Name())

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("failed to parse %s: %w", filePath, err)
		}

		l.translations[lang] = translations
	}

	return nil
}

// Get returns a translation for the given key and language
func (l *Localizer) Get(lang, key string) string {
	if translations, ok := l.translations[lang]; ok {
		if value, ok := translations[key]; ok {
			return value
		}
	}

	// Fallback to default language
	if translations, ok := l.translations[l.fallbackLang]; ok {
		if value, ok := translations[key]; ok {
			return value
		}
	}

	return key // Return key if translation not found
}
