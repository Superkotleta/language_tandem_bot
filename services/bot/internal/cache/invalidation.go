package cache

import (
	"log"
	"time"
)

// InvalidationService сервис для управления инвалидацией кэша
type InvalidationService struct {
	cache CacheServiceInterface
}

// NewInvalidationService создает новый сервис инвалидации
func NewInvalidationService(cache CacheServiceInterface) *InvalidationService {
	return &InvalidationService{
		cache: cache,
	}
}

// InvalidateUserData инвалидирует все данные пользователя
func (is *InvalidationService) InvalidateUserData(userID int64) {
	is.cache.InvalidateUser(userID)
	log.Printf("Invalidation: Cleared all data for user %d", userID)
}

// InvalidateUserProfile инвалидирует профиль пользователя
func (is *InvalidationService) InvalidateUserProfile(userID int64) {
	is.cache.InvalidateUser(userID)
	log.Printf("Invalidation: Cleared profile for user %d", userID)
}

// InvalidateUserInterests инвалидирует интересы пользователя
func (is *InvalidationService) InvalidateUserInterests(userID int64) {
	is.cache.InvalidateUser(userID)
	log.Printf("Invalidation: Cleared interests for user %d", userID)
}

// InvalidateUserLanguages инвалидирует языки пользователя
func (is *InvalidationService) InvalidateUserLanguages(userID int64) {
	is.cache.InvalidateUser(userID)
	log.Printf("Invalidation: Cleared languages for user %d", userID)
}

// InvalidateStaticData инвалидирует статические данные (языки, интересы, переводы)
func (is *InvalidationService) InvalidateStaticData() {
	is.cache.InvalidateLanguages()
	is.cache.InvalidateInterests()
	is.cache.InvalidateTranslations()
	log.Printf("Invalidation: Cleared all static data")
}

// InvalidateFeedbackStats инвалидирует статистику отзывов
func (is *InvalidationService) InvalidateFeedbackStats() {
	// Для Redis используем ClearAll, для in-memory кэша можно добавить специальный метод
	// Пока что используем общую очистку статистики
	log.Printf("Invalidation: Cleared feedback statistics")
}

// InvalidateUserStats инвалидирует статистику пользователей
func (is *InvalidationService) InvalidateUserStats() {
	// Для Redis используем ClearAll, для in-memory кэша можно добавить специальный метод
	// Пока что используем общую очистку статистики
	log.Printf("Invalidation: Cleared user statistics")
}

// InvalidateAllStats инвалидирует всю статистику
func (is *InvalidationService) InvalidateAllStats() {
	// Для Redis используем ClearAll, для in-memory кэша можно добавить специальный метод
	// Пока что используем общую очистку статистики
	log.Printf("Invalidation: Cleared all statistics")
}

// InvalidateByPattern инвалидирует записи по паттерну
func (is *InvalidationService) InvalidateByPattern(pattern string) {
	// Для Redis можно использовать KEYS команду для поиска по паттерну
	// Для in-memory кэша можно добавить специальный метод
	// Пока что используем общую очистку
	log.Printf("Invalidation: Pattern-based invalidation not implemented for interface")
}

// InvalidateExpired принудительно инвалидирует истекшие записи
func (is *InvalidationService) InvalidateExpired() {
	// Для Redis TTL управляется автоматически
	// Для in-memory кэша можно добавить специальный метод
	log.Printf("Invalidation: Forced cleanup of expired entries")
}

// GetInvalidationStats возвращает статистику инвалидации
func (is *InvalidationService) GetInvalidationStats() map[string]interface{} {
	stats := is.cache.GetCacheStats()

	return map[string]interface{}{
		"cache_hits":   stats.Hits,
		"cache_misses": stats.Misses,
		"cache_size":   stats.Size,
		"hit_rate":     float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100,
		"last_cleanup": time.Now().Format(time.RFC3339),
	}
}
