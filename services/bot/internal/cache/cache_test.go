package cache

import (
	"testing"
	"time"

	"language-exchange-bot/internal/models"
)

func TestCacheService(t *testing.T) {
	// Создаем кэш с коротким TTL для тестов
	config := &CacheConfig{
		LanguagesTTL:    100 * time.Millisecond,
		InterestsTTL:    100 * time.Millisecond,
		TranslationsTTL: 100 * time.Millisecond,
		UsersTTL:        100 * time.Millisecond,
		StatsTTL:        100 * time.Millisecond,
	}

	cache := NewCacheService(config)
	defer cache.Stop()

	// Тест языков
	t.Run("Languages", func(t *testing.T) {
		languages := []*models.Language{
			{ID: 1, Code: "en", NameNative: "English", NameEn: "English"},
			{ID: 2, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
		}

		// Сохраняем в кэш
		cache.SetLanguages("en", languages)

		// Получаем из кэша
		cached, found := cache.GetLanguages("en")
		if !found {
			t.Error("Expected to find languages in cache")
		}
		if len(cached) != 2 {
			t.Errorf("Expected 2 languages, got %d", len(cached))
		}

		// Проверяем, что данные корректны
		if cached[0].Code != "en" {
			t.Errorf("Expected first language to be 'en', got %s", cached[0].Code)
		}
	})

	// Тест интересов
	t.Run("Interests", func(t *testing.T) {
		interests := map[int]string{
			1: "movies",
			2: "music",
		}

		// Сохраняем в кэш
		cache.SetInterests("en", interests)

		// Получаем из кэша
		cached, found := cache.GetInterests("en")
		if !found {
			t.Error("Expected to find interests in cache")
		}
		if len(cached) != 2 {
			t.Errorf("Expected 2 interests, got %d", len(cached))
		}

		// Проверяем, что данные корректны
		if cached[1] != "movies" {
			t.Errorf("Expected interest 1 to be 'movies', got %s", cached[1])
		}
	})

	// Тест пользователей
	t.Run("Users", func(t *testing.T) {
		user := &models.User{
			ID:                    1,
			TelegramID:            12345,
			Username:              "testuser",
			FirstName:             "Test",
			InterfaceLanguageCode: "en",
			NativeLanguageCode:    "en",
			TargetLanguageCode:    "ru",
		}

		// Сохраняем в кэш
		cache.SetUser(user)

		// Получаем из кэша
		cached, found := cache.GetUser(1)
		if !found {
			t.Error("Expected to find user in cache")
		}
		if cached != nil && cached.TelegramID != 12345 {
			t.Errorf("Expected user TelegramID to be 12345, got %d", cached.TelegramID)
		}
	})

	// Тест истечения TTL
	t.Run("TTL Expiration", func(t *testing.T) {
		// Ждем истечения TTL
		time.Sleep(150 * time.Millisecond)

		// Запускаем очистку
		cache.cleanupExpired()

		// Проверяем, что данные истекли
		_, found := cache.GetLanguages("en")
		if found {
			t.Error("Expected languages to be expired")
		}

		_, found = cache.GetInterests("en")
		if found {
			t.Error("Expected interests to be expired")
		}

		_, found = cache.GetUser(1)
		if found {
			t.Error("Expected user to be expired")
		}
	})

	// Тест статистики
	t.Run("Stats", func(t *testing.T) {
		stats := cache.GetCacheStats()
		if stats.Hits == 0 && stats.Misses == 0 {
			t.Error("Expected some cache activity")
		}
	})
}

func TestInvalidationService(t *testing.T) {
	cache := NewCacheService(DefaultCacheConfig())
	defer cache.Stop()

	invalidation := NewInvalidationService(cache)

	// Добавляем тестовые данные
	cache.SetLanguages("en", []*models.Language{{ID: 1, Code: "en"}})
	cache.SetInterests("en", map[int]string{1: "test"})

	// Проверяем, что данные есть
	_, found := cache.GetLanguages("en")
	if !found {
		t.Error("Expected languages to be in cache")
	}

	// Инвалидируем статические данные
	invalidation.InvalidateStaticData()

	// Проверяем, что данные удалены
	_, found = cache.GetLanguages("en")
	if found {
		t.Error("Expected languages to be invalidated")
	}
}

func TestMetricsService(t *testing.T) {
	cache := NewCacheService(DefaultCacheConfig())
	defer cache.Stop()

	metrics := NewMetricsService(cache)

	// Записываем несколько метрик
	metrics.RecordRequest(10*time.Millisecond, true)
	metrics.RecordRequest(20*time.Millisecond, false)
	metrics.RecordError()

	// Проверяем метрики
	perfMetrics := metrics.GetPerformanceMetrics()
	if perfMetrics["total_requests"] != int64(2) {
		t.Errorf("Expected 2 total requests, got %v", perfMetrics["total_requests"])
	}

	if perfMetrics["error_count"] != int64(2) {
		t.Errorf("Expected 2 errors, got %v", perfMetrics["error_count"])
	}
}
