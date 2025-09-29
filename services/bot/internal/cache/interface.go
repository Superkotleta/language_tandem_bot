package cache

import (
	"language-exchange-bot/internal/models"
)

// CacheServiceInterface интерфейс для кэш-сервиса
type CacheServiceInterface interface {
	// Languages
	GetLanguages(lang string) ([]*models.Language, bool)
	SetLanguages(lang string, languages []*models.Language)

	// Interests
	GetInterests(lang string) (map[int]string, bool)
	SetInterests(lang string, interests map[int]string)

	// Users
	GetUser(userID int64) (*models.User, bool)
	SetUser(user *models.User)

	// Translations
	GetTranslations(lang string) (map[string]string, bool)
	SetTranslations(lang string, translations map[string]string)

	// Stats
	GetStats(statsType string) (map[string]interface{}, bool)
	SetStats(statsType string, data map[string]interface{})

	// Invalidation
	InvalidateUser(userID int64)
	InvalidateLanguages()
	InvalidateInterests()
	InvalidateTranslations()
	ClearAll()

	// Stats and control
	GetCacheStats() CacheStats
	Stop()
	String() string
}
