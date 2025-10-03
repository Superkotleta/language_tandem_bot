package cache //nolint:testpackage

import (
	"context"
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
	assert.Equal(t, 0, stats.Hits)
	assert.Equal(t, 0, stats.Misses)
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
	assert.Equal(t, 1, stats.Hits)
	assert.Equal(t, 1, stats.Misses)
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

	// Запускаем горутины для конкурентного доступа
	done := make(chan bool, 10)

	for i := range 10 {
		go func(id int) {
			defer func() { done <- true }()

			// Добавляем данные
			languages := []*models.Language{{ID: id, Code: "ru"}}
			service.SetLanguages(context.Background(), "ru", languages)

			// Читаем данные
			_, found := service.GetLanguages(context.Background(), "ru")
			require.True(t, found)
		}(i)
	}

	// Ждем завершения всех горутин
	for range 10 {
		<-done
	}

	// Проверяем, что сервис не упал
	stats := service.GetCacheStats(context.Background())
	assert.GreaterOrEqual(t, stats.Size, 0)
}
