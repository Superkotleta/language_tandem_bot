package localization

import (
	"testing"

	"language-exchange-bot/internal/localization"

	"github.com/stretchr/testify/assert"
)

func TestLocalizer_Get(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		key      string
		expected string
	}{
		{
			name:     "English welcome message",
			lang:     "en",
			key:      "welcome_message",
			expected: "👋 Hi, {name}! Welcome to Language Exchange Bot!",
		},
		{
			name:     "Russian welcome message",
			lang:     "ru",
			key:      "welcome_message",
			expected: "👋 Привет, {name}! Добро пожаловать в Language Exchange Bot!",
		},
		{
			name:     "Spanish welcome message",
			lang:     "es",
			key:      "welcome_message",
			expected: "👋 ¡Hola, {name}! ¡Bienvenido al Language Exchange Bot!",
		},
		{
			name:     "Unknown language fallback to English",
			lang:     "unknown",
			key:      "welcome_message",
			expected: "👋 Hi, {name}! Welcome to Language Exchange Bot!",
		},
		{
			name:     "Unknown key returns key",
			lang:     "en",
			key:      "unknown_key",
			expected: "unknown_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			localizer := localization.NewLocalizer(nil)

			// Act
			result := localizer.Get(tt.lang, tt.key)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizer_GetWithParams(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		key      string
		params   map[string]string
		expected string
	}{
		{
			name:     "English with name parameter",
			lang:     "en",
			key:      "welcome_message",
			params:   map[string]string{"name": "John"},
			expected: "👋 Hi, John! Welcome to Language Exchange Bot!",
		},
		{
			name:     "Russian with name parameter",
			lang:     "ru",
			key:      "welcome_message",
			params:   map[string]string{"name": "Иван"},
			expected: "👋 Привет, Иван! Добро пожаловать в Language Exchange Bot!",
		},
		{
			name:     "Multiple parameters",
			lang:     "en",
			key:      "welcome_message",
			params:   map[string]string{"name": "Test", "extra": "ignored"},
			expected: "👋 Hi, Test! Welcome to Language Exchange Bot!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			localizer := localization.NewLocalizer(nil)

			// Act
			result := localizer.GetWithParams(tt.lang, tt.key, tt.params)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizer_GetLanguageName(t *testing.T) {
	tests := []struct {
		name          string
		langCode      string
		interfaceLang string
		expected      string
	}{
		{
			name:          "English language name in English",
			langCode:      "en",
			interfaceLang: "en",
			expected:      "English",
		},
		{
			name:          "Russian language name in English",
			langCode:      "ru",
			interfaceLang: "en",
			expected:      "Russian",
		},
		{
			name:          "English language name in Russian",
			langCode:      "en",
			interfaceLang: "ru",
			expected:      "Английский",
		},
		{
			name:          "Russian language name in Russian",
			langCode:      "ru",
			interfaceLang: "ru",
			expected:      "Русский",
		},
		{
			name:          "Spanish language name in Spanish",
			langCode:      "es",
			interfaceLang: "es",
			expected:      "Español",
		},
		{
			name:          "Unknown language fallback",
			langCode:      "unknown",
			interfaceLang: "en",
			expected:      "language_unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			localizer := localization.NewLocalizer(nil)

			// Act
			result := localizer.GetLanguageName(tt.langCode, tt.interfaceLang)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizer_GetInterests(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		expected map[int]string
	}{
		{
			name: "English interests",
			lang: "en",
			expected: map[int]string{
				1: "Movies",
				2: "Music",
				3: "Sports",
				4: "Travel",
			},
		},
		{
			name: "Russian interests",
			lang: "ru",
			expected: map[int]string{
				1: "Фильмы",
				2: "Музыка",
				3: "Спорт",
				4: "Путешествия",
			},
		},
		{
			name: "Spanish interests",
			lang: "es",
			expected: map[int]string{
				1: "Movies",
				2: "Music",
				3: "Sports",
				4: "Travel",
			},
		},
		{
			name: "Unknown language fallback to English",
			lang: "unknown",
			expected: map[int]string{
				1: "Movies",
				2: "Music",
				3: "Sports",
				4: "Travel",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			localizer := localization.NewLocalizer(nil)

			// Act
			result, err := localizer.GetInterests(tt.lang)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizer_NewLocalizer(t *testing.T) {
	// Test with nil database
	localizer := localization.NewLocalizer(nil)
	assert.NotNil(t, localizer)

	// Test that localizer works correctly
	result := localizer.Get("en", "welcome_message")
	assert.NotEmpty(t, result)
}

func TestLocalizer_EdgeCases(t *testing.T) {
	localizer := localization.NewLocalizer(nil)

	// Test empty language
	result := localizer.Get("", "welcome_message")
	assert.Equal(t, "👋 Hi, {name}! Welcome to Language Exchange Bot!", result)

	// Test empty key
	result = localizer.Get("en", "")
	assert.Equal(t, "", result)

	// Test nil parameters
	result = localizer.GetWithParams("en", "welcome_message", nil)
	assert.Equal(t, "👋 Hi, {name}! Welcome to Language Exchange Bot!", result)

	// Test empty parameters
	result = localizer.GetWithParams("en", "welcome_message", map[string]string{})
	assert.Equal(t, "👋 Hi, {name}! Welcome to Language Exchange Bot!", result)

	// Test missing parameter
	result = localizer.GetWithParams("en", "welcome_message", map[string]string{"other": "value"})
	assert.Equal(t, "👋 Hi, {name}! Welcome to Language Exchange Bot!", result)
}

func TestLocalizer_GetLanguageName_EdgeCases(t *testing.T) {
	localizer := localization.NewLocalizer(nil)

	// Test empty language code
	result := localizer.GetLanguageName("", "en")
	assert.Equal(t, "language_", result)

	// Test empty interface language
	result = localizer.GetLanguageName("en", "")
	assert.Equal(t, "English", result)

	// Test both empty
	result = localizer.GetLanguageName("", "")
	assert.Equal(t, "language_", result)
}

func TestLocalizer_GetInterests_EdgeCases(t *testing.T) {
	localizer := localization.NewLocalizer(nil)

	// Test empty language
	result, err := localizer.GetInterests("")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Movies", result[1])
}
