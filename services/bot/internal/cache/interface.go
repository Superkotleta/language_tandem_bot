package cache

import (
	"context"
	"language-exchange-bot/internal/models"
)

// ServiceInterface интерфейс для кэш-сервиса.
type ServiceInterface interface {
	// Languages
	GetLanguages(ctx context.Context, lang string) ([]*models.Language, bool)
	SetLanguages(ctx context.Context, lang string, languages []*models.Language)

	// Interests
	GetInterests(ctx context.Context, lang string) (map[int]string, bool)
	SetInterests(ctx context.Context, lang string, interests map[int]string)

	// Users
	GetUser(ctx context.Context, userID int64) (*models.User, bool)
	SetUser(ctx context.Context, user *models.User)

	// Translations
	GetTranslations(ctx context.Context, lang string) (map[string]string, bool)
	SetTranslations(ctx context.Context, lang string, translations map[string]string)

	// Stats
	GetStats(ctx context.Context, statsType string) (map[string]interface{}, bool)
	SetStats(ctx context.Context, statsType string, data map[string]interface{})

	// Invalidation
	InvalidateUser(ctx context.Context, userID int64)
	InvalidateLanguages(ctx context.Context)
	InvalidateInterests(ctx context.Context)
	InvalidateTranslations(ctx context.Context)
	ClearAll(ctx context.Context)

	// Stats and control
	GetCacheStats(ctx context.Context) Stats
	Stop()
	String() string
}
