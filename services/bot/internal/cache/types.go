package cache

import (
	"time"

	"language-exchange-bot/internal/models"
)

// Константы для TTL кэша.
const (
	// translationsTTLMinutes - время жизни переводов в кэше (30 минут).
	translationsTTLMinutes = 30

	// usersTTLMinutes - время жизни пользователей в кэше (15 минут).
	usersTTLMinutes = 15

	// statsTTLMinutes - время жизни статистики в кэше (5 минут).
	statsTTLMinutes = 5
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
	LanguagesTTL    time.Duration // TTL для языков
	InterestsTTL    time.Duration // TTL для интересов
	TranslationsTTL time.Duration // TTL для переводов
	UsersTTL        time.Duration // TTL для пользователей
	StatsTTL        time.Duration // TTL для статистики
}

// DefaultConfig возвращает конфигурацию по умолчанию.
func DefaultConfig() *Config {
	return &Config{
		LanguagesTTL:    time.Hour,                            // 1 час - языки редко изменяются
		InterestsTTL:    time.Hour,                            // 1 час - интересы редко изменяются
		TranslationsTTL: translationsTTLMinutes * time.Minute, // 30 минут - переводы статичны
		UsersTTL:        usersTTLMinutes * time.Minute,        // 15 минут - пользователи могут изменяться
		StatsTTL:        statsTTLMinutes * time.Minute,        // 5 минут - статистика часто обновляется
	}
}

// Stats статистика работы кэша.
type Stats struct {
	Hits   int64 // Количество попаданий в кэш
	Misses int64 // Количество промахов кэша
	Size   int   // Текущий размер кэша
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
