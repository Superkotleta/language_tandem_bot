// Package cache provides caching functionality with support for in-memory and Redis backends.
package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"language-exchange-bot/internal/models"
)

// Константы для вычислений.
const (
	// PercentageMultiplier - множитель для преобразования в проценты.
	PercentageMultiplier = 100.0
	// CacheCleanupInterval - интервал очистки кэша.
	CacheCleanupInterval = 5 * time.Minute
)

// Service основной сервис кэширования.
type Service struct {
	// Кэши для разных типов данных
	languages    map[string]*Entry // Ключ: язык интерфейса
	interests    map[string]*Entry // Ключ: язык интерфейса
	translations map[string]*Entry // Ключ: язык интерфейса
	users        map[int64]*Entry  // Ключ: user ID
	stats        map[string]*Entry // Ключ: тип статистики

	// Конфигурация
	config *Config

	// Потокобезопасность
	mutex sync.RWMutex

	// Статистика
	cacheStats Stats

	// Канал для остановки очистки
	stopCleanup chan struct{}
}

// NewService создает новый экземпляр Service.
func NewService(config *Config) *Service {
	if config == nil {
		config = DefaultConfig()
	}

	cacheService := &Service{
		languages:    make(map[string]*Entry),
		interests:    make(map[string]*Entry),
		translations: make(map[string]*Entry),
		users:        make(map[int64]*Entry),
		stats:        make(map[string]*Entry),
		config:       config,
		stopCleanup:  make(chan struct{}),
		mutex:        sync.RWMutex{},
		cacheStats: Stats{
			Hits:   0,
			Misses: 0,
			Size:   0,
		},
	}

	// Запускаем фоновую очистку истекших записей
	go cacheService.startCleanup()

	return cacheService
}

// GetLanguages получает языки из кэша или возвращает nil если нет в кэше.
func (cacheService *Service) GetLanguages(_ context.Context, lang string) ([]*models.Language, bool) {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	entry, exists := cacheService.languages[lang]
	if !exists || entry == nil || entry.IsExpired() {
		cacheService.cacheStats.Misses++

		return nil, false
	}

	cacheService.cacheStats.Hits++

	if data, ok := entry.Data.(*CachedLanguages); ok && data != nil {
		return data.Languages, true
	}

	cacheService.cacheStats.Misses++

	return nil, false
}

// SetLanguages сохраняет языки в кэш.
func (cacheService *Service) SetLanguages(_ context.Context, lang string, languages []*models.Language) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.languages[lang] = &Entry{
		Data: &CachedLanguages{
			Languages: languages,
			Lang:      lang,
		},
		ExpiresAt: time.Now().Add(cacheService.config.LanguagesTTL),
	}

	cacheService.updateSize()
}

// GetInterests получает интересы из кэша или возвращает nil если нет в кэше.
func (cacheService *Service) GetInterests(_ context.Context, lang string) (map[int]string, bool) {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	entry, exists := cacheService.interests[lang]
	if !exists || entry == nil || entry.IsExpired() {
		cacheService.cacheStats.Misses++

		return nil, false
	}

	cacheService.cacheStats.Hits++

	if data, ok := entry.Data.(*CachedInterests); ok && data != nil {
		return data.Interests, true
	}

	cacheService.cacheStats.Misses++

	return nil, false
}

// SetInterests сохраняет интересы в кэш.
func (cacheService *Service) SetInterests(_ context.Context, lang string, interests map[int]string) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.interests[lang] = &Entry{
		Data: &CachedInterests{
			Interests: interests,
			Lang:      lang,
		},
		ExpiresAt: time.Now().Add(cacheService.config.InterestsTTL),
	}

	cacheService.updateSize()
}

// GetUser получает пользователя из кэша или возвращает nil если нет в кэше.
func (cacheService *Service) GetUser(_ context.Context, userID int64) (*models.User, bool) {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	entry, exists := cacheService.users[userID]
	if !exists || entry == nil || entry.IsExpired() {
		cacheService.cacheStats.Misses++

		return nil, false
	}

	cacheService.cacheStats.Hits++

	if data, ok := entry.Data.(*CachedUser); ok && data != nil {
		return data.User, true
	}

	cacheService.cacheStats.Misses++

	return nil, false
}

// SetUser сохраняет пользователя в кэш.
func (cacheService *Service) SetUser(_ context.Context, user *models.User) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.users[int64(user.ID)] = &Entry{
		Data: &CachedUser{
			User: user,
			Lang: user.InterfaceLanguageCode,
		},
		ExpiresAt: time.Now().Add(cacheService.config.UsersTTL),
	}

	cacheService.updateSize()
}

// GetTranslations получает переводы из кэша или возвращает nil если нет в кэше.
func (cacheService *Service) GetTranslations(_ context.Context, lang string) (map[string]string, bool) {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	entry, exists := cacheService.translations[lang]
	if !exists || entry.IsExpired() {
		cacheService.cacheStats.Misses++

		return nil, false
	}

	cacheService.cacheStats.Hits++

	if data, ok := entry.Data.(map[string]string); ok {
		return data, true
	}

	cacheService.cacheStats.Misses++

	return nil, false
}

// SetTranslations сохраняет переводы в кэш.
func (cacheService *Service) SetTranslations(_ context.Context, lang string, translations map[string]string) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.translations[lang] = &Entry{
		Data:      translations,
		ExpiresAt: time.Now().Add(cacheService.config.TranslationsTTL),
	}

	cacheService.updateSize()
}

// GetStats получает статистику из кэша или возвращает nil если нет в кэше.
func (cacheService *Service) GetStats(_ context.Context, statsType string) (map[string]interface{}, bool) {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	entry, exists := cacheService.stats[statsType]
	if !exists || entry.IsExpired() {
		cacheService.cacheStats.Misses++

		return nil, false
	}

	cacheService.cacheStats.Hits++

	if data, ok := entry.Data.(*CachedStats); ok {
		return data.Data, true
	}

	cacheService.cacheStats.Misses++

	return nil, false
}

// SetStats сохраняет статистику в кэш.
func (cacheService *Service) SetStats(_ context.Context, statsType string, data map[string]interface{}) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.stats[statsType] = &Entry{
		Data: &CachedStats{
			Data: data,
			Type: statsType,
		},
		ExpiresAt: time.Now().Add(cacheService.config.StatsTTL),
	}

	cacheService.updateSize()
}

// InvalidateUser удаляет пользователя из кэша.
func (cacheService *Service) InvalidateUser(_ context.Context, userID int64) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	delete(cacheService.users, userID)
	cacheService.updateSize()

	log.Printf("Cache: Invalidated user %d", userID)
}

// InvalidateLanguages удаляет языки из кэша.
func (cacheService *Service) InvalidateLanguages(_ context.Context) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.languages = make(map[string]*Entry)
	cacheService.updateSize()

	log.Printf("Cache: Invalidated all languages")
}

// InvalidateInterests удаляет интересы из кэша.
func (cacheService *Service) InvalidateInterests(_ context.Context) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.interests = make(map[string]*Entry)
	cacheService.updateSize()

	log.Printf("Cache: Invalidated all interests")
}

// InvalidateTranslations удаляет переводы из кэша.
func (cacheService *Service) InvalidateTranslations(_ context.Context) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.translations = make(map[string]*Entry)
	cacheService.updateSize()

	log.Printf("Cache: Invalidated all translations")
}

// ClearAll очищает весь кэш.
func (cacheService *Service) ClearAll(_ context.Context) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.languages = make(map[string]*Entry)
	cacheService.interests = make(map[string]*Entry)
	cacheService.translations = make(map[string]*Entry)
	cacheService.users = make(map[int64]*Entry)
	cacheService.stats = make(map[string]*Entry)
	cacheService.updateSize()

	log.Printf("Cache: Cleared all data")
}

// GetCacheStats returns cache statisticacheService.
func (cacheService *Service) GetCacheStats(_ context.Context) Stats {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	return Stats{
		Hits:   cacheService.cacheStats.Hits,
		Misses: cacheService.cacheStats.Misses,
		Size:   cacheService.cacheStats.Size,
	}
}

// Stop останавливает кэш-сервис.
func (cacheService *Service) Stop() {
	close(cacheService.stopCleanup)
	log.Printf("Cache: Service stopped")
}

// String возвращает строковое представление статистики кэша.
func (cacheService *Service) String() string {
	stats := cacheService.GetCacheStats(context.Background())

	hitRate := float64(0)

	if stats.Hits+stats.Misses > 0 {
		hitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses) * PercentageMultiplier
	}

	return fmt.Sprintf("Cache Stats: Hits=%d, Misses=%d, HitRate=%.2f%%, Size=%d",
		stats.Hits, stats.Misses, hitRate, stats.Size)
}

// updateSize обновляет размер кэша.
func (cacheService *Service) updateSize() {
	cacheService.cacheStats.Size = len(cacheService.languages) + len(cacheService.interests) +
		len(cacheService.translations) + len(cacheService.users) + len(cacheService.stats)
}

// startCleanup запускает фоновую очистку истекших записей.
func (cacheService *Service) startCleanup() {
	ticker := time.NewTicker(CacheCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cacheService.cleanupExpired()
		case <-cacheService.stopCleanup:
			return
		}
	}
}

// cleanupExpired удаляет истекшие записи.
func (cacheService *Service) cleanupExpired() {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cleaned := cacheService.cleanupLanguages() + cacheService.cleanupInterests() + cacheService.cleanupTranslations() +
		cacheService.cleanupUsers() + cacheService.cleanupStats()

	cacheService.updateSize()

	if cleaned > 0 {
		log.Printf("Cache: Cleaned %d expired entries", cleaned)
	}
}

// cleanupLanguages очищает истекшие языки.
func (cacheService *Service) cleanupLanguages() int {
	cleaned := 0

	for key, entry := range cacheService.languages {
		if entry.IsExpired() {
			delete(cacheService.languages, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupInterests очищает истекшие интересы.
func (cacheService *Service) cleanupInterests() int {
	cleaned := 0

	for key, entry := range cacheService.interests {
		if entry.IsExpired() {
			delete(cacheService.interests, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupTranslations очищает истекшие переводы.
func (cacheService *Service) cleanupTranslations() int {
	cleaned := 0

	for key, entry := range cacheService.translations {
		if entry.IsExpired() {
			delete(cacheService.translations, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupUsers очищает истекших пользователей.
func (cacheService *Service) cleanupUsers() int {
	cleaned := 0

	for key, entry := range cacheService.users {
		if entry.IsExpired() {
			delete(cacheService.users, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupStats очищает истекшую статистику.
func (cacheService *Service) cleanupStats() int {
	cleaned := 0

	for key, entry := range cacheService.stats {
		if entry.IsExpired() {
			delete(cacheService.stats, key)

			cleaned++
		}
	}

	return cleaned
}
