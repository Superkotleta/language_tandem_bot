package cache //nolint:testpackage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"language-exchange-bot/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetLanguages_CacheHit(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	languages := []*models.Language{
		{ID: 1, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
		{ID: 2, Code: "en", NameNative: "English", NameEn: "English"},
	}

	// Сохраняем в кэш
	service.SetLanguages(context.Background(), "ru", languages)

	// Получаем из кэша
	result, found := service.GetLanguages(context.Background(), "ru")

	// Проверяем результаты
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, languages[0].Code, result[0].Code)
	assert.Equal(t, languages[1].Code, result[1].Code)
}

func TestService_GetLanguages_CacheMiss(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Получаем из кэша (должен быть промах)
	result, found := service.GetLanguages(context.Background(), "ru")

	// Проверяем результаты
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestService_GetLanguages_ExpiredEntry(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша с очень коротким TTL
	config := &Config{
		LanguagesTTL: 1 * time.Nanosecond,
	}
	service := NewService(config)

	// Подготавливаем тестовые данные
	languages := []*models.Language{
		{ID: 1, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
	}

	// Сохраняем в кэш
	service.SetLanguages(context.Background(), "ru", languages)

	// Ждем истечения TTL
	time.Sleep(1 * time.Millisecond)

	// Получаем из кэша (должен быть промах из-за истечения)
	result, found := service.GetLanguages(context.Background(), "ru")

	// Проверяем результаты
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestService_SetLanguages(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	languages := []*models.Language{
		{ID: 1, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
		{ID: 2, Code: "en", NameNative: "English", NameEn: "English"},
	}

	// Сохраняем в кэш
	service.SetLanguages(context.Background(), "ru", languages)

	// Проверяем, что данные сохранились
	stats := service.GetCacheStats(context.Background())
	assert.Equal(t, 1, stats.Size)
}

func TestService_GetInterests_CacheHit(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	interests := map[int]string{
		1: "Фильмы",
		2: "Музыка",
		3: "Спорт",
	}

	// Сохраняем в кэш
	service.SetInterests(context.Background(), "ru", interests)

	// Получаем из кэша
	result, found := service.GetInterests(context.Background(), "ru")

	// Проверяем результаты
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.Equal(t, "Фильмы", result[1])
	assert.Equal(t, "Музыка", result[2])
	assert.Equal(t, "Спорт", result[3])
}

func TestService_GetUser_CacheHit(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	user := &models.User{
		ID:                     1,
		TelegramID:             12345,
		Username:               "testuser",
		FirstName:              "Test",
		NativeLanguageCode:     "ru",
		TargetLanguageCode:     "en",
		TargetLanguageLevel:    "intermediate",
		InterfaceLanguageCode:  "ru",
		State:                  "active",
		Status:                 "active",
		ProfileCompletionLevel: 100,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	// Сохраняем в кэш
	service.SetUser(context.Background(), user)

	// Получаем из кэша
	result, found := service.GetUser(context.Background(), int64(user.ID))

	// Проверяем результаты
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.TelegramID, result.TelegramID)
	assert.Equal(t, user.Username, result.Username)
}

func TestService_InvalidateUser(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	user := &models.User{
		ID:         1,
		TelegramID: 12345,
		Username:   "testuser",
		FirstName:  "Test",
	}

	// Сохраняем в кэш
	service.SetUser(context.Background(), user)

	// Проверяем, что пользователь в кэше
	result, found := service.GetUser(context.Background(), int64(user.ID))
	assert.True(t, found)
	assert.NotNil(t, result)

	// Инвалидируем пользователя
	service.InvalidateUser(context.Background(), int64(user.ID))

	// Проверяем, что пользователь удален из кэша
	result, found = service.GetUser(context.Background(), int64(user.ID))
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestService_ClearAll(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Добавляем данные в кэш
	languages := []*models.Language{{ID: 1, Code: "ru"}}
	interests := map[int]string{1: "Фильмы"}
	user := &models.User{ID: 1, TelegramID: 12345}

	service.SetLanguages(context.Background(), "ru", languages)
	service.SetInterests(context.Background(), "ru", interests)
	service.SetUser(context.Background(), user)

	// Проверяем, что данные в кэше
	stats := service.GetCacheStats(context.Background())
	assert.Equal(t, 3, stats.Size)

	// Очищаем весь кэш
	service.ClearAll(context.Background())

	// Проверяем, что кэш пуст
	stats = service.GetCacheStats(context.Background())
	assert.Equal(t, 0, stats.Size)

	// Проверяем, что данные недоступны
	_, found := service.GetLanguages(context.Background(), "ru")
	assert.False(t, found)

	_, found = service.GetInterests(context.Background(), "ru")
	assert.False(t, found)

	_, found = service.GetUser(context.Background(), int64(user.ID))
	assert.False(t, found)
}

func TestService_GetCacheStats(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Начальная статистика
	stats := service.GetCacheStats(context.Background())
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, 0, stats.Size)

	// Добавляем данные и делаем запросы
	languages := []*models.Language{{ID: 1, Code: "ru"}}
	service.SetLanguages(context.Background(), "ru", languages)

	// Cache hit
	_, found := service.GetLanguages(context.Background(), "ru")
	assert.True(t, found)

	// Cache miss
	_, found = service.GetLanguages(context.Background(), "en")
	assert.False(t, found)

	// Проверяем статистику
	stats = service.GetCacheStats(context.Background())
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, 1, stats.Size)
}

func TestService_String(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Добавляем данные
	languages := []*models.Language{{ID: 1, Code: "ru"}}
	service.SetLanguages(context.Background(), "ru", languages)

	// Делаем запросы для генерации статистики
	service.GetLanguages(context.Background(), "ru") // Hit
	service.GetLanguages(context.Background(), "en") // Miss

	// Проверяем строковое представление
	result := service.String()
	assert.Contains(t, result, "Cache Stats:")
	assert.Contains(t, result, "Hits=1")
	assert.Contains(t, result, "Misses=1")
	assert.Contains(t, result, "Size=1")
	assert.Contains(t, result, "HitRate=50.00%")
}

func TestService_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Используем sync.WaitGroup для корректной синхронизации
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Используем уникальный ключ для каждого горутина
			langKey := fmt.Sprintf("ru_%d", id)

			// Добавляем данные
			languages := []*models.Language{{ID: id, Code: langKey}}
			service.SetLanguages(context.Background(), langKey, languages)

			// Читаем данные
			_, found := service.GetLanguages(context.Background(), langKey)
			require.True(t, found)
		}(i)
	}

	// Ждем завершения всех горутин
	wg.Wait()

	// Проверяем, что сервис не упал
	stats := service.GetCacheStats(context.Background())
	assert.GreaterOrEqual(t, stats.Size, 0)
}

func TestService_GetInterests_CacheMiss(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Получаем из кэша (должен быть промах)
	result, found := service.GetInterests(context.Background(), "ru")

	// Проверяем результаты
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestService_SetInterests(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	interests := map[int]string{
		1: "Music",
		2: "Sports",
		3: "Books",
	}

	// Сохраняем в кэш
	service.SetInterests(context.Background(), "ru", interests)

	// Получаем из кэша
	result, found := service.GetInterests(context.Background(), "ru")

	// Проверяем результаты
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.Equal(t, "Music", result[1])
	assert.Equal(t, "Sports", result[2])
	assert.Equal(t, "Books", result[3])
}

func TestService_GetTranslations_CacheHit(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	translations := map[string]string{
		"hello": "привет",
		"world": "мир",
	}

	// Сохраняем в кэш
	service.SetTranslations(context.Background(), "ru", translations)

	// Получаем из кэша
	result, found := service.GetTranslations(context.Background(), "ru")

	// Проверяем результаты
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "привет", result["hello"])
	assert.Equal(t, "мир", result["world"])
}

func TestService_GetTranslations_CacheMiss(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Получаем из кэша (должен быть промах)
	result, found := service.GetTranslations(context.Background(), "ru")

	// Проверяем результаты
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestService_GetStats_CacheHit(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	stats := map[string]interface{}{
		"total_users":  1000,
		"active_users": 500,
	}

	// Сохраняем в кэш
	service.SetStats(context.Background(), "user_stats", stats)

	// Получаем из кэша
	result, found := service.GetStats(context.Background(), "user_stats")

	// Проверяем результаты
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Equal(t, 1000, result["total_users"])
	assert.Equal(t, 500, result["active_users"])
}

func TestService_GetInterestCategories_CacheHit(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	categories := []*models.InterestCategory{
		{ID: 1, Name: "Sports", Description: "Sports activities"},
		{ID: 2, Name: "Music", Description: "Music related activities"},
	}

	// Сохраняем в кэш
	service.SetInterestCategories(context.Background(), "ru", categories)

	// Получаем из кэша
	result, found := service.GetInterestCategories(context.Background(), "ru")

	// Проверяем результаты
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "Sports", result[0].Name)
	assert.Equal(t, "Music", result[1].Name)
}

func TestService_GetUserStats_CacheHit(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	userStats := map[string]interface{}{
		"messages_sent": 100,
		"last_active":   time.Now(),
	}

	// Сохраняем в кэш
	service.SetUserStats(context.Background(), 12345, userStats)

	// Получаем из кэша
	result, found := service.GetUserStats(context.Background(), 12345)

	// Проверяем результаты
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Equal(t, 100, result["messages_sent"])
}

func TestService_GetConfig_CacheHit(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	configValue := map[string]interface{}{
		"max_connections": 100,
		"timeout":         30,
	}

	// Сохраняем в кэш
	service.SetConfig(context.Background(), "db_config", configValue)

	// Получаем из кэша
	result, found := service.GetConfig(context.Background(), "db_config")

	// Проверяем результаты
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Equal(t, configValue, result)
}

func TestService_InvalidateLanguages(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	languages := []*models.Language{
		{ID: 1, Code: "ru", NameNative: "Русский"},
	}

	// Сохраняем в кэш
	service.SetLanguages(context.Background(), "ru", languages)

	// Проверяем, что данные есть
	_, found := service.GetLanguages(context.Background(), "ru")
	assert.True(t, found)

	// Инвалидируем
	service.InvalidateLanguages(context.Background())

	// Проверяем, что данных нет
	_, found = service.GetLanguages(context.Background(), "ru")
	assert.False(t, found)
}

func TestService_InvalidateInterests(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	interests := map[int]string{
		1: "Music",
	}

	// Сохраняем в кэш
	service.SetInterests(context.Background(), "ru", interests)

	// Проверяем, что данные есть
	_, found := service.GetInterests(context.Background(), "ru")
	assert.True(t, found)

	// Инвалидируем
	service.InvalidateInterests(context.Background())

	// Проверяем, что данных нет
	_, found = service.GetInterests(context.Background(), "ru")
	assert.False(t, found)
}

func TestService_InvalidateTranslations(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	translations := map[string]string{
		"hello": "привет",
	}

	// Сохраняем в кэш
	service.SetTranslations(context.Background(), "ru", translations)

	// Проверяем, что данные есть
	_, found := service.GetTranslations(context.Background(), "ru")
	assert.True(t, found)

	// Инвалидируем
	service.InvalidateTranslations(context.Background())

	// Проверяем, что данных нет
	_, found = service.GetTranslations(context.Background(), "ru")
	assert.False(t, found)
}

func TestService_InvalidateInterestCategories(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	categories := []*models.InterestCategory{
		{ID: 1, Name: "Sports"},
	}

	// Сохраняем в кэш
	service.SetInterestCategories(context.Background(), "ru", categories)

	// Проверяем, что данные есть
	_, found := service.GetInterestCategories(context.Background(), "ru")
	assert.True(t, found)

	// Инвалидируем
	service.InvalidateInterestCategories(context.Background())

	// Проверяем, что данных нет
	_, found = service.GetInterestCategories(context.Background(), "ru")
	assert.False(t, found)
}

func TestService_InvalidateUserStats(t *testing.T) {
	t.Parallel()
	// Создаем сервис кэша
	config := DefaultConfig()
	service := NewService(config)

	// Подготавливаем тестовые данные
	userStats := map[string]interface{}{
		"messages_sent": 100,
	}

	// Сохраняем в кэш
	service.SetUserStats(context.Background(), 12345, userStats)

	// Проверяем, что данные есть
	_, found := service.GetUserStats(context.Background(), 12345)
	assert.True(t, found)

	// Инвалидируем
	service.InvalidateUserStats(context.Background(), 12345)

	// Проверяем, что данных нет
	_, found = service.GetUserStats(context.Background(), 12345)
	assert.False(t, found)
}
