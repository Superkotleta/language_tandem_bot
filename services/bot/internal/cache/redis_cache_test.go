package cache_test

import (
	"testing"

	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/models"
)

func TestRedisCacheServiceLanguages(t *testing.T) {
	t.Parallel()

	redisCache := createRedisCacheService(t)
	defer redisCache.Stop()

	languages := []*models.Language{
		{ID: 1, Code: "en", NameNative: "English", NameEn: "English"},
		{ID: 2, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
	}

	// Сохраняем в Redis
	redisCache.SetLanguages("en", languages)

	// Получаем из Redis
	cached, found := redisCache.GetLanguages("en")
	if !found {
		t.Error("Expected to find languages in Redis cache")
	}

	if len(cached) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(cached))
	}

	// Проверяем, что данные корректны
	if cached[0].Code != "en" {
		t.Errorf("Expected first language to be 'en', got %s", cached[0].Code)
	}
}

func TestRedisCacheServiceInterests(t *testing.T) {
	t.Parallel()

	redisCache := createRedisCacheService(t)
	defer redisCache.Stop()

	interests := map[int]string{
		1: "movies",
		2: "music",
	}

	// Сохраняем в Redis
	redisCache.SetInterests("en", interests)

	// Получаем из Redis
	cached, found := redisCache.GetInterests("en")
	if !found {
		t.Error("Expected to find interests in Redis cache")
	}

	if len(cached) != 2 {
		t.Errorf("Expected 2 interests, got %d", len(cached))
	}

	// Проверяем, что данные корректны
	if cached[1] != "movies" {
		t.Errorf("Expected interest 1 to be 'movies', got %s", cached[1])
	}
}

func TestRedisCacheServiceUsers(t *testing.T) {
	t.Parallel()

	redisCache := createRedisCacheService(t)
	defer redisCache.Stop()

	user := &models.User{
		ID:                    1,
		TelegramID:            12345,
		Username:              "testuser",
		FirstName:             "Test",
		InterfaceLanguageCode: "en",
		NativeLanguageCode:    "en",
		TargetLanguageCode:    "ru",
	}

	// Сохраняем в Redis
	redisCache.SetUser(user)

	// Получаем из Redis
	cached, found := redisCache.GetUser(1)
	if !found {
		t.Error("Expected to find user in Redis cache")
	}

	if cached != nil && cached.TelegramID != 12345 {
		t.Errorf("Expected user TelegramID to be 12345, got %d", cached.TelegramID)
	}
}

func TestRedisCacheServiceInvalidation(t *testing.T) {
	t.Parallel()

	redisCache := createRedisCacheService(t)
	defer redisCache.Stop()

	// Добавляем тестовые данные
	redisCache.SetLanguages("en", []*models.Language{{ID: 1, Code: "en"}})
	redisCache.SetInterests("en", map[int]string{1: "test"})

	// Проверяем, что данные есть
	_, found := redisCache.GetLanguages("en")
	if !found {
		t.Error("Expected languages to be in cache")
	}

	// Инвалидируем статические данные
	redisCache.InvalidateLanguages()
	redisCache.InvalidateInterests()

	// Проверяем, что данные удалены
	_, found = redisCache.GetLanguages("en")
	if found {
		t.Error("Expected languages to be invalidated")
	}
}

func TestRedisCacheServiceStats(t *testing.T) {
	t.Parallel()

	redisCache := createRedisCacheService(t)
	defer redisCache.Stop()

	stats := redisCache.GetCacheStats()
	if stats.Size < 0 {
		t.Error("Expected valid cache size")
	}
}

// createRedisCacheService создает Redis кэш сервис для тестов.
func createRedisCacheService(t *testing.T) *cache.RedisCacheService {
	t.Helper()

	redisCache, err := cache.NewRedisCacheService("localhost:6379", "", 0, cache.DefaultConfig())
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	return redisCache
}

func TestRedisCacheServiceConnection(t *testing.T) {
	t.Parallel()

	// Тест подключения к Redis
	_, err := cache.NewRedisCacheService("localhost:6379", "", 0, cache.DefaultConfig())
	if err != nil {
		t.Logf("Redis connection test failed (expected if Redis not running): %v", err)
	} else {
		t.Log("Redis connection successful")
	}
}

func TestRedisCacheServiceInvalidConnection(t *testing.T) {
	t.Parallel()

	// Тест с неверными параметрами подключения
	_, err := cache.NewRedisCacheService("localhost:9999", "", 0, cache.DefaultConfig())
	if err == nil {
		t.Error("Expected connection error for invalid Redis URL")
	}
}
