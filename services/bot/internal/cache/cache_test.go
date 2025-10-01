package cache_test

import (
	"context"
	"testing"
	"time"

	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/models"
)

func TestCacheServiceLanguages(t *testing.T) {
	t.Parallel()

	config := createTestCacheConfig()

	cacheService := cache.NewService(config)
	defer cacheService.Stop()

	languages := []*models.Language{
		{ID: 1, Code: "en", NameNative: "English", NameEn: "English"},
		{ID: 2, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
	}

	// Сохраняем в кэш
	cacheService.SetLanguages(context.Background(), "en", languages)

	// Получаем из кэша
	cached, found := cacheService.GetLanguages(context.Background(), "en")
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
}

func TestCacheServiceInterests(t *testing.T) {
	t.Parallel()

	config := createTestCacheConfig()

	cacheService := cache.NewService(config)
	defer cacheService.Stop()

	interests := map[int]string{
		1: "movies",
		2: "music",
	}

	// Сохраняем в кэш
	cacheService.SetInterests(context.Background(), "en", interests)

	// Получаем из кэша
	cached, found := cacheService.GetInterests(context.Background(), "en")
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
}

func TestCacheServiceUsers(t *testing.T) {
	t.Parallel()

	config := createTestCacheConfig()

	cacheService := cache.NewService(config)
	defer cacheService.Stop()

	user := &models.User{
		ID:                     1,
		TelegramID:             12345,
		Username:               "testuser",
		FirstName:              "Test",
		NativeLanguageCode:     "en",
		TargetLanguageCode:     "ru",
		TargetLanguageLevel:    "beginner",
		InterfaceLanguageCode:  "en",
		State:                  "active",
		Status:                 "active",
		ProfileCompletionLevel: 100,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Interests:              []int{1, 2, 3},
		TimeAvailability: &models.TimeAvailability{
			DayType:      "any",
			SpecificDays: []string{},
			TimeSlot:     "any",
		},
		FriendshipPreferences: &models.FriendshipPreferences{
			ActivityType:       "casual_chat",
			CommunicationStyle: "text",
			CommunicationFreq:  "weekly",
		},
	}

	// Сохраняем в кэш
	cacheService.SetUser(context.Background(), user)

	// Получаем из кэша
	cached, found := cacheService.GetUser(context.Background(), 1)
	if !found {
		t.Error("Expected to find user in cache")
	}

	if cached != nil && cached.TelegramID != 12345 {
		t.Errorf("Expected user TelegramID to be 12345, got %d", cached.TelegramID)
	}
}

func TestCacheServiceTTLExpiration(t *testing.T) {
	t.Parallel()

	config := createTestCacheConfig()

	cacheService := cache.NewService(config)
	defer cacheService.Stop()

	// Ждем истечения TTL
	time.Sleep(150 * time.Millisecond)

	// Проверяем, что данные истекли
	_, found := cacheService.GetLanguages(context.Background(), "en")
	if found {
		t.Error("Expected languages to be expired")
	}

	_, found = cacheService.GetInterests(context.Background(), "en")
	if found {
		t.Error("Expected interests to be expired")
	}

	_, found = cacheService.GetUser(context.Background(), 1)
	if found {
		t.Error("Expected user to be expired")
	}
}

func TestCacheServiceStats(t *testing.T) {
	t.Parallel()

	config := createTestCacheConfig()

	cacheService := cache.NewService(config)
	defer cacheService.Stop()

	// Выполняем операции с кэшем для генерации статистики
	_, _ = cacheService.GetLanguages(context.Background(), "en")
	_, _ = cacheService.GetInterests(context.Background(), "en")
	_, _ = cacheService.GetTranslations(context.Background(), "en")

	stats := cacheService.GetCacheStats(context.Background())
	if stats.Hits == 0 && stats.Misses == 0 {
		t.Error("Expected some cache activity")
	}
}

// createTestCacheConfig создает конфигурацию кэша для тестов.
func createTestCacheConfig() *cache.Config {
	return &cache.Config{
		LanguagesTTL:    100 * time.Millisecond,
		InterestsTTL:    100 * time.Millisecond,
		TranslationsTTL: 100 * time.Millisecond,
		UsersTTL:        100 * time.Millisecond,
		StatsTTL:        100 * time.Millisecond,
	}
}

func TestInvalidationService(t *testing.T) {
	t.Parallel()

	cacheService := cache.NewService(cache.DefaultConfig())
	defer cacheService.Stop()

	invalidation := cache.NewInvalidationService(cacheService)

	// Добавляем тестовые данные
	cacheService.SetLanguages(context.Background(), "en", []*models.Language{{ID: 1, Code: "en"}})
	cacheService.SetInterests(context.Background(), "en", map[int]string{1: "test"})

	// Проверяем, что данные есть
	_, found := cacheService.GetLanguages(context.Background(), "en")
	if !found {
		t.Error("Expected languages to be in cache")
	}

	// Инвалидируем статические данные
	invalidation.InvalidateStaticData()

	// Проверяем, что данные удалены
	_, found = cacheService.GetLanguages(context.Background(), "en")
	if found {
		t.Error("Expected languages to be invalidated")
	}
}

func TestMetricsService(t *testing.T) {
	t.Parallel()

	cacheService := cache.NewService(cache.DefaultConfig())
	defer cacheService.Stop()

	metrics := cache.NewMetricsService(cacheService)

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
