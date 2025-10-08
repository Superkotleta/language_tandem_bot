package localization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewLocalizer tests creating new localizer
func TestNewLocalizer(t *testing.T) {
	localizer := NewLocalizer(nil)

	assert.NotNil(t, localizer)
	assert.Nil(t, localizer.db)
	assert.NotNil(t, localizer.translations)
	assert.NotNil(t, localizer.logger)
	assert.NotNil(t, localizer.errorHandler)
}

// TestLocalizer_Get tests getting translations
func TestLocalizer_Get(t *testing.T) {
	localizer := NewLocalizer(nil)

	// Test with empty translations map
	result := localizer.Get("en", "test_key")
	assert.Equal(t, "test_key", result) // Should return key when translation not found

	// Test with fallback translations loaded
	localizer.loadFallbackTranslations()
	result = localizer.Get("en", "welcome")
	assert.NotEmpty(t, result)
}

// TestLocalizer_GetWithFallback tests getting translations with fallback
func TestLocalizer_GetWithFallback(t *testing.T) {
	localizer := NewLocalizer(nil)

	// Load fallback translations
	localizer.loadFallbackTranslations()

	// Test existing translation
	result := localizer.Get("en", "welcome")
	assert.NotEmpty(t, result)

	// Test non-existing translation
	result = localizer.Get("en", "nonexistent_key")
	assert.Equal(t, "nonexistent_key", result)

	// Test fallback to English for unknown language
	result = localizer.Get("unknown_lang", "welcome")
	assert.NotEmpty(t, result)
}

// TestLocalizer_GetInterests tests getting interests
func TestLocalizer_GetInterests(t *testing.T) {
	localizer := NewLocalizer(nil)

	// Load fallback translations
	localizer.loadFallbackTranslations()

	// Test getting interests
	interests, err := localizer.GetInterests("en")
	assert.NoError(t, err)
	assert.NotNil(t, interests)
}
