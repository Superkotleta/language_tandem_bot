// Package cache provides caching functionality with support for in-memory and Redis backends.
package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
)

// CacheCleanupInterval - интервал очистки кэша (используется из centralized constants).
var CacheCleanupInterval = localization.CacheCleanupMinutes * time.Minute

// PercentageMultiplier constant is now defined in localization/constants.go

// Service основной сервис кэширования.
type Service struct {
	// Кэши для разных типов данных
	languages          map[string]*Entry // Ключ: язык интерфейса
	interests          map[string]*Entry // Ключ: язык интерфейса
	translations       map[string]*Entry // Ключ: язык интерфейса
	users              map[int64]*Entry  // Ключ: user ID
	stats              map[string]*Entry // Ключ: тип статистики
	interestCategories map[string]*Entry // Ключ: язык интерфейса
	userStats          map[int64]*Entry  // Ключ: user ID
	configCache        map[string]*Entry // Ключ: config key

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
		languages:          make(map[string]*Entry),
		interests:          make(map[string]*Entry),
		translations:       make(map[string]*Entry),
		users:              make(map[int64]*Entry),
		stats:              make(map[string]*Entry),
		interestCategories: make(map[string]*Entry),
		userStats:          make(map[int64]*Entry),
		configCache:        make(map[string]*Entry),
		config:             config,
		stopCleanup:        make(chan struct{}),
		mutex:              sync.RWMutex{},
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
		hitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses) * localization.PercentageMultiplier
	}

	return fmt.Sprintf("Cache Stats: Hits=%d, Misses=%d, HitRate=%.2f%%, Size=%d, Evictions=%d, Memory=%dKB",
		stats.Hits, stats.Misses, hitRate, stats.Size, stats.Evictions, stats.MemoryUsage/localization.BytesInKilobyte)
}

// updateSize обновляет размер кэша и метрики.
func (cacheService *Service) updateSize() {
	cacheService.cacheStats.Size = len(cacheService.languages) + len(cacheService.interests) +
		len(cacheService.translations) + len(cacheService.users) + len(cacheService.stats) +
		len(cacheService.interestCategories) + len(cacheService.userStats) + len(cacheService.configCache)

	// Обновляем hit ratio
	total := cacheService.cacheStats.Hits + cacheService.cacheStats.Misses
	if total > 0 {
		cacheService.cacheStats.HitRatio = float64(cacheService.cacheStats.Hits) / float64(total)
	}

	// Обновляем использование памяти (приблизительная оценка)
	cacheService.cacheStats.MemoryUsage = int64(cacheService.cacheStats.Size * localization.BytesInKilobyte) // 1KB на запись
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
		cacheService.cleanupUsers() + cacheService.cleanupStats() + cacheService.cleanupInterestCategories() +
		cacheService.cleanupUserStats() + cacheService.cleanupConfigCache()

	// Обновляем счетчик evictions
	cacheService.cacheStats.Evictions += int64(cleaned)

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

// cleanupInterestCategories очищает истекшие категории интересов.
func (cacheService *Service) cleanupInterestCategories() int {
	cleaned := 0

	for key, entry := range cacheService.interestCategories {
		if entry.IsExpired() {
			delete(cacheService.interestCategories, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupUserStats очищает истекшую статистику пользователей.
func (cacheService *Service) cleanupUserStats() int {
	cleaned := 0

	for key, entry := range cacheService.userStats {
		if entry.IsExpired() {
			delete(cacheService.userStats, key)

			cleaned++
		}
	}

	return cleaned
}

// cleanupConfigCache очищает истекшую конфигурацию.
func (cacheService *Service) cleanupConfigCache() int {
	cleaned := 0

	for key, entry := range cacheService.configCache {
		if entry.IsExpired() {
			delete(cacheService.configCache, key)

			cleaned++
		}
	}

	return cleaned
}

// Set сохраняет произвольные данные в кэш.
func (cacheService *Service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	// Используем общую карту для произвольных данных
	if cacheService.stats == nil {
		cacheService.stats = make(map[string]*Entry)
	}

	cacheService.stats[key] = &Entry{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
	}

	cacheService.updateSize()

	return nil
}

// Get получает произвольные данные из кэша.
func (cacheService *Service) Get(ctx context.Context, key string, dest interface{}) error {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	entry, exists := cacheService.stats[key]
	if !exists || entry == nil || entry.IsExpired() {
		return fmt.Errorf("key not found: %s", key)
	}

	// Простое присваивание для совместимости
	// В реальной реализации нужно было бы использовать reflection
	*dest.(*interface{}) = entry.Data

	return nil
}

// Delete удаляет ключ из кэша.
func (cacheService *Service) Delete(ctx context.Context, key string) error {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	delete(cacheService.stats, key)
	cacheService.updateSize()

	return nil
}

// ===== НОВЫЕ МЕТОДЫ КЕШИРОВАНИЯ =====

// GetInterestCategories получает категории интересов из кэша.
func (cacheService *Service) GetInterestCategories(ctx context.Context, lang string) ([]*models.InterestCategory, bool) {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	entry, exists := cacheService.interestCategories[lang]
	if !exists || entry.IsExpired() {
		cacheService.cacheStats.Misses++

		return nil, false
	}

	cacheService.cacheStats.Hits++

	return entry.Data.([]*models.InterestCategory), true
}

// SetInterestCategories сохраняет категории интересов в кэш.
func (cacheService *Service) SetInterestCategories(ctx context.Context, lang string, categories []*models.InterestCategory) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.interestCategories[lang] = &Entry{
		Data:      categories,
		ExpiresAt: time.Now().Add(cacheService.config.LanguagesTTL),
	}
	cacheService.updateSize()
}

// GetUserStats получает статистику пользователя из кэша.
func (cacheService *Service) GetUserStats(ctx context.Context, userID int64) (map[string]interface{}, bool) {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	entry, exists := cacheService.userStats[userID]
	if !exists || entry.IsExpired() {
		cacheService.cacheStats.Misses++

		return nil, false
	}

	cacheService.cacheStats.Hits++

	return entry.Data.(map[string]interface{}), true
}

// SetUserStats сохраняет статистику пользователя в кэш.
func (cacheService *Service) SetUserStats(ctx context.Context, userID int64, stats map[string]interface{}) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.userStats[userID] = &Entry{
		Data:      stats,
		ExpiresAt: time.Now().Add(cacheService.config.LanguagesTTL),
	}
	cacheService.updateSize()
}

// GetConfig получает конфигурацию из кэша.
func (cacheService *Service) GetConfig(ctx context.Context, configKey string) (interface{}, bool) {
	cacheService.mutex.RLock()
	defer cacheService.mutex.RUnlock()

	entry, exists := cacheService.configCache[configKey]
	if !exists || entry.IsExpired() {
		cacheService.cacheStats.Misses++

		return nil, false
	}

	cacheService.cacheStats.Hits++

	return entry.Data, true
}

// SetConfig сохраняет конфигурацию в кэш.
func (cacheService *Service) SetConfig(ctx context.Context, configKey string, value interface{}) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.configCache[configKey] = &Entry{
		Data:      value,
		ExpiresAt: time.Now().Add(cacheService.config.LanguagesTTL),
	}
	cacheService.updateSize()
}

// InvalidateInterestCategories инвалидирует кэш категорий интересов.
func (cacheService *Service) InvalidateInterestCategories(ctx context.Context) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.interestCategories = make(map[string]*Entry)
	cacheService.updateSize()
}

// InvalidateUserStats инвалидирует кэш статистики пользователя.
func (cacheService *Service) InvalidateUserStats(ctx context.Context, userID int64) {
	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	delete(cacheService.userStats, userID)
	cacheService.updateSize()
}

// WarmUp предзагружает критичные данные в кэш при старте приложения.
func (cacheService *Service) WarmUp(ctx context.Context, dataLoader DataLoader) error {
	log.Println("Starting cache warming...")

	start := time.Now()

	// Список языков для предзагрузки
	languages := []string{"ru", "en", "es", "zh"}

	// Предзагружаем языки для всех поддерживаемых языков интерфейса
	for _, lang := range languages {
		if err := cacheService.warmUpLanguages(ctx, dataLoader, lang); err != nil {
			log.Printf("Failed to warm up languages for %s: %v", lang, err)
			// Продолжаем с другими языками даже если один не удался
		}
	}

	// Предзагружаем интересы для всех языков
	for _, lang := range languages {
		if err := cacheService.warmUpInterests(ctx, dataLoader, lang); err != nil {
			log.Printf("Failed to warm up interests for %s: %v", lang, err)
		}
	}

	// Предзагружаем категории интересов
	for _, lang := range languages {
		if err := cacheService.warmUpInterestCategories(ctx, dataLoader, lang); err != nil {
			log.Printf("Failed to warm up interest categories for %s: %v", lang, err)
		}
	}

	// Предзагружаем переводы
	for _, lang := range languages {
		if err := cacheService.warmUpTranslations(ctx, dataLoader, lang); err != nil {
			log.Printf("Failed to warm up translations for %s: %v", lang, err)
		}
	}

	duration := time.Since(start)
	log.Printf("Cache warming completed in %v", duration)

	return nil
}

// DataLoader интерфейс для загрузки данных из внешних источников.
type DataLoader interface {
	LoadLanguages(ctx context.Context, lang string) ([]*models.Language, error)
	LoadInterests(ctx context.Context, lang string) (map[int]string, error)
	LoadInterestCategories(ctx context.Context, lang string) (map[int]string, error)
	LoadTranslations(ctx context.Context, lang string) (map[string]string, error)
}

// warmUpLanguages предзагружает языки для указанного языка интерфейса.
func (cacheService *Service) warmUpLanguages(ctx context.Context, dataLoader DataLoader, lang string) error {
	languages, err := dataLoader.LoadLanguages(ctx, lang)
	if err != nil {
		return fmt.Errorf("failed to load languages for %s: %w", lang, err)
	}

	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cachedData := &CachedLanguages{
		Languages: languages,
		Lang:      lang,
	}

	cacheService.languages[lang] = &Entry{
		Data:      cachedData,
		ExpiresAt: time.Now().Add(cacheService.config.LanguagesTTL),
	}

	cacheService.updateSize()

	return nil
}

// warmUpInterests предзагружает интересы для указанного языка интерфейса.
func (cacheService *Service) warmUpInterests(ctx context.Context, dataLoader DataLoader, lang string) error {
	interests, err := dataLoader.LoadInterests(ctx, lang)
	if err != nil {
		return fmt.Errorf("failed to load interests for %s: %w", lang, err)
	}

	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cachedData := &CachedInterests{
		Interests: interests,
		Lang:      lang,
	}

	cacheService.interests[lang] = &Entry{
		Data:      cachedData,
		ExpiresAt: time.Now().Add(cacheService.config.InterestsTTL),
	}

	cacheService.updateSize()

	return nil
}

// warmUpInterestCategories предзагружает категории интересов для указанного языка интерфейса.
func (cacheService *Service) warmUpInterestCategories(ctx context.Context, dataLoader DataLoader, lang string) error {
	categories, err := dataLoader.LoadInterestCategories(ctx, lang)
	if err != nil {
		return fmt.Errorf("failed to load interest categories for %s: %w", lang, err)
	}

	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cachedData := &CachedInterests{
		Interests: categories,
		Lang:      lang,
	}

	cacheService.interestCategories[lang] = &Entry{
		Data:      cachedData,
		ExpiresAt: time.Now().Add(cacheService.config.InterestCategoriesTTL),
	}

	cacheService.updateSize()

	return nil
}

// warmUpTranslations предзагружает переводы для указанного языка интерфейса.
func (cacheService *Service) warmUpTranslations(ctx context.Context, dataLoader DataLoader, lang string) error {
	translations, err := dataLoader.LoadTranslations(ctx, lang)
	if err != nil {
		return fmt.Errorf("failed to load translations for %s: %w", lang, err)
	}

	cacheService.mutex.Lock()
	defer cacheService.mutex.Unlock()

	cacheService.translations[lang] = &Entry{
		Data:      translations,
		ExpiresAt: time.Now().Add(cacheService.config.TranslationsTTL),
	}

	cacheService.updateSize()

	return nil
}
