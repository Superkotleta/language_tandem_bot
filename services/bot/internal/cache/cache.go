// Package cache provides caching functionality with support for in-memory and Redis backends.
package cache

import (
	"fmt"
	"log"
	"sync"
	"time"

	"language-exchange-bot/internal/models"
)

// Константы для вычислений.
const (
	// percentageMultiplier - множитель для преобразования в проценты.
	percentageMultiplier = 100
	// cleanupIntervalMinutes - интервал очистки кэша в минутах.
	cleanupIntervalMinutes = 5
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

	cs := &Service{
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
	go cs.startCleanup()

	return cs
}

// GetLanguages получает языки из кэша или возвращает nil если нет в кэше.
func (cs *Service) GetLanguages(lang string) ([]*models.Language, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	entry, exists := cs.languages[lang]
	if !exists || entry.IsExpired() {
		cs.cacheStats.Misses++

		return nil, false
	}

	cs.cacheStats.Hits++

	if data, ok := entry.Data.(*CachedLanguages); ok {
		return data.Languages, true
	}

	cs.cacheStats.Misses++

	return nil, false
}

// SetLanguages сохраняет языки в кэш.
func (cs *Service) SetLanguages(lang string, languages []*models.Language) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.languages[lang] = &Entry{
		Data: &CachedLanguages{
			Languages: languages,
			Lang:      lang,
		},
		ExpiresAt: time.Now().Add(cs.config.LanguagesTTL),
	}

	cs.updateSize()
}

// GetInterests получает интересы из кэша или возвращает nil если нет в кэше.
func (cs *Service) GetInterests(lang string) (map[int]string, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	entry, exists := cs.interests[lang]
	if !exists || entry.IsExpired() {
		cs.cacheStats.Misses++

		return nil, false
	}

	cs.cacheStats.Hits++

	if data, ok := entry.Data.(*CachedInterests); ok {
		return data.Interests, true
	}

	cs.cacheStats.Misses++

	return nil, false
}

// SetInterests сохраняет интересы в кэш.
func (cs *Service) SetInterests(lang string, interests map[int]string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.interests[lang] = &Entry{
		Data: &CachedInterests{
			Interests: interests,
			Lang:      lang,
		},
		ExpiresAt: time.Now().Add(cs.config.InterestsTTL),
	}

	cs.updateSize()
}

// GetUser получает пользователя из кэша или возвращает nil если нет в кэше.
func (cs *Service) GetUser(userID int64) (*models.User, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	entry, exists := cs.users[userID]
	if !exists || entry.IsExpired() {
		cs.cacheStats.Misses++

		return nil, false
	}

	cs.cacheStats.Hits++

	if data, ok := entry.Data.(*CachedUser); ok {
		return data.User, true
	}

	cs.cacheStats.Misses++

	return nil, false
}

// SetUser сохраняет пользователя в кэш.
func (cs *Service) SetUser(user *models.User) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.users[int64(user.ID)] = &Entry{
		Data: &CachedUser{
			User: user,
			Lang: user.InterfaceLanguageCode,
		},
		ExpiresAt: time.Now().Add(cs.config.UsersTTL),
	}

	cs.updateSize()
}

// GetTranslations получает переводы из кэша или возвращает nil если нет в кэше.
func (cs *Service) GetTranslations(lang string) (map[string]string, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	entry, exists := cs.translations[lang]
	if !exists || entry.IsExpired() {
		cs.cacheStats.Misses++

		return nil, false
	}

	cs.cacheStats.Hits++

	if data, ok := entry.Data.(map[string]string); ok {
		return data, true
	}

	cs.cacheStats.Misses++

	return nil, false
}

// SetTranslations сохраняет переводы в кэш.
func (cs *Service) SetTranslations(lang string, translations map[string]string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.translations[lang] = &Entry{
		Data:      translations,
		ExpiresAt: time.Now().Add(cs.config.TranslationsTTL),
	}

	cs.updateSize()
}

// GetStats получает статистику из кэша или возвращает nil если нет в кэше.
func (cs *Service) GetStats(statsType string) (map[string]interface{}, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	entry, exists := cs.stats[statsType]
	if !exists || entry.IsExpired() {
		cs.cacheStats.Misses++

		return nil, false
	}

	cs.cacheStats.Hits++

	if data, ok := entry.Data.(*CachedStats); ok {
		return data.Data, true
	}

	cs.cacheStats.Misses++

	return nil, false
}

// SetStats сохраняет статистику в кэш.
func (cs *Service) SetStats(statsType string, data map[string]interface{}) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.stats[statsType] = &Entry{
		Data: &CachedStats{
			Data: data,
			Type: statsType,
		},
		ExpiresAt: time.Now().Add(cs.config.StatsTTL),
	}

	cs.updateSize()
}

// InvalidateUser удаляет пользователя из кэша.
func (cs *Service) InvalidateUser(userID int64) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.users, userID)
	cs.updateSize()

	log.Printf("Cache: Invalidated user %d", userID)
}

// InvalidateLanguages удаляет языки из кэша.
func (cs *Service) InvalidateLanguages() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.languages = make(map[string]*Entry)
	cs.updateSize()

	log.Printf("Cache: Invalidated all languages")
}

// InvalidateInterests удаляет интересы из кэша.
func (cs *Service) InvalidateInterests() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.interests = make(map[string]*Entry)
	cs.updateSize()

	log.Printf("Cache: Invalidated all interests")
}

// InvalidateTranslations удаляет переводы из кэша.
func (cs *Service) InvalidateTranslations() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.translations = make(map[string]*Entry)
	cs.updateSize()

	log.Printf("Cache: Invalidated all translations")
}

// ClearAll очищает весь кэш.
func (cs *Service) ClearAll() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.languages = make(map[string]*Entry)
	cs.interests = make(map[string]*Entry)
	cs.translations = make(map[string]*Entry)
	cs.users = make(map[int64]*Entry)
	cs.stats = make(map[string]*Entry)
	cs.updateSize()

	log.Printf("Cache: Cleared all data")
}

// GetCacheStats returns cache statistics.
func (cs *Service) GetCacheStats() Stats {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return Stats{
		Hits:   cs.cacheStats.Hits,
		Misses: cs.cacheStats.Misses,
		Size:   cs.cacheStats.Size,
	}
}

// Stop останавливает кэш-сервис.
func (cs *Service) Stop() {
	close(cs.stopCleanup)
	log.Printf("Cache: Service stopped")
}

// String возвращает строковое представление статистики кэша.
func (cs *Service) String() string {
	stats := cs.GetCacheStats()

	hitRate := float64(0)

	if stats.Hits+stats.Misses > 0 {
		hitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses) * percentageMultiplier
	}

	return fmt.Sprintf("Cache Stats: Hits=%d, Misses=%d, HitRate=%.2f%%, Size=%d",
		stats.Hits, stats.Misses, hitRate, stats.Size)
}

// updateSize обновляет размер кэша.
func (cs *Service) updateSize() {
	cs.cacheStats.Size = len(cs.languages) + len(cs.interests) + len(cs.translations) + len(cs.users) + len(cs.stats)
}

// startCleanup запускает фоновую очистку истекших записей.
func (cs *Service) startCleanup() {
	ticker := time.NewTicker(cleanupIntervalMinutes * time.Minute) // Проверяем каждые 5 минут
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cs.cleanupExpired()
		case <-cs.stopCleanup:
			return
		}
	}
}

// cleanupExpired удаляет истекшие записи.
func (cs *Service) cleanupExpired() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cleaned := cs.cleanupLanguages() + cs.cleanupInterests() + cs.cleanupTranslations() +
		cs.cleanupUsers() + cs.cleanupStats()

	cs.updateSize()

	if cleaned > 0 {
		log.Printf("Cache: Cleaned %d expired entries", cleaned)
	}
}

// cleanupLanguages очищает истекшие языки.
func (cs *Service) cleanupLanguages() int {
	cleaned := 0

	for key, entry := range cs.languages {
		if entry.IsExpired() {
			delete(cs.languages, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupInterests очищает истекшие интересы.
func (cs *Service) cleanupInterests() int {
	cleaned := 0

	for key, entry := range cs.interests {
		if entry.IsExpired() {
			delete(cs.interests, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupTranslations очищает истекшие переводы.
func (cs *Service) cleanupTranslations() int {
	cleaned := 0

	for key, entry := range cs.translations {
		if entry.IsExpired() {
			delete(cs.translations, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupUsers очищает истекших пользователей.
func (cs *Service) cleanupUsers() int {
	cleaned := 0

	for key, entry := range cs.users {
		if entry.IsExpired() {
			delete(cs.users, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupStats очищает истекшую статистику.
func (cs *Service) cleanupStats() int {
	cleaned := 0

	for key, entry := range cs.stats {
		if entry.IsExpired() {
			delete(cs.stats, key)

			cleaned++
		}
	}

	return cleaned
}
