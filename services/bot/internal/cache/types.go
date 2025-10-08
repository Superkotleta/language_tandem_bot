package cache

import (
	"time"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
)

// Entry представляет запись в кэше с TTL.
type Entry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired проверяет, истек ли срок действия записи.
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Config конфигурация кэша.
type Config struct {
	LanguagesTTL          time.Duration // TTL для языков
	InterestsTTL          time.Duration // TTL для интересов
	InterestCategoriesTTL time.Duration // TTL для категорий интересов
	TranslationsTTL       time.Duration // TTL для переводов
	UsersTTL              time.Duration // TTL для пользователей
	UserStatsTTL          time.Duration // TTL для статистики пользователей
	StatsTTL              time.Duration // TTL для статистики
	ConfigTTL             time.Duration // TTL для конфигурации
}

// DefaultConfig возвращает конфигурацию по умолчанию.
func DefaultConfig() *Config {
	return &Config{
		LanguagesTTL:          time.Hour,                                         // 1 час - языки редко изменяются
		InterestsTTL:          time.Hour,                                         // 1 час - интересы редко изменяются
		InterestCategoriesTTL: time.Hour,                                         // 1 час - категории интересов статичны
		TranslationsTTL:       localization.TranslationsTTLMinutes * time.Minute, // 30 минут - переводы статичны
		UsersTTL:              localization.UsersTTLMinutes * time.Minute,        // 15 минут - пользователи могут изменяться
		UserStatsTTL:          time.Hour,                                         // 1 час - статистика пользователей
		StatsTTL:              localization.StatsTTLMinutes * time.Minute,        // 5 минут - статистика часто обновляется
		ConfigTTL:             time.Hour * 24,                                    // 24 часа - конфигурация редко изменяется
	}
}

// Stats статистика работы кэша.
type Stats struct {
	Hits        int64   // Количество попаданий в кэш
	Misses      int64   // Количество промахов кэша
	Size        int     // Текущий размер кэша
	HitRatio    float64 // Процент попаданий (0.0-1.0)
	Evictions   int64   // Количество вытеснений
	MemoryUsage int64   // Использование памяти в байтах
}

// CachedLanguages кэшированные языки.
type CachedLanguages struct {
	Languages []*models.Language
	Lang      string // Язык интерфейса для локализации
}

// CachedInterests кэшированные интересы.
type CachedInterests struct {
	Interests map[int]string
	Lang      string // Язык интерфейса для локализации
}

// CachedUser кэшированный пользователь.
type CachedUser struct {
	User *models.User
	Lang string // Язык интерфейса пользователя
}

// CachedStats кэшированная статистика.
type CachedStats struct {
	Data map[string]interface{}
	Type string // Тип статистики (feedbacks, users, etc.)
}
