package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"language-exchange-bot/internal/models"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache интерфейс для кэширования.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	GetLanguages(ctx context.Context) ([]*models.Language, error)
	SetLanguages(ctx context.Context, languages []*models.Language) error
	GetInterests(ctx context.Context, langCode string) (map[int]string, error)
	SetInterests(ctx context.Context, langCode string, interests map[int]string) error
	GetUserProfile(ctx context.Context, userID int) (*models.User, error)
	SetUserProfile(ctx context.Context, user *models.User) error
	InvalidateUserProfile(ctx context.Context, userID int) error
	Close() error
}

// RedisCache реализация кэша на Redis.
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache создает новый Redis кэш.
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		// Оптимизация настроек
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx := context.Background()

	// Проверяем соединение
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Get получает значение из кэша.
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set сохраняет значение в кэш.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// Delete удаляет значение из кэша.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// GetLanguages получает языки из кэша.
func (c *RedisCache) GetLanguages(ctx context.Context) ([]*models.Language, error) {
	data, err := c.client.Get(ctx, "languages").Result()
	if err != nil {
		return nil, err
	}

	var languages []*models.Language
	err = json.Unmarshal([]byte(data), &languages)
	return languages, err
}

// SetLanguages сохраняет языки в кэш.
func (c *RedisCache) SetLanguages(ctx context.Context, languages []*models.Language) error {
	return c.Set(ctx, "languages", languages, 24*time.Hour) // Кэшируем на сутки
}

// GetInterests получает интересы из кэша.
func (c *RedisCache) GetInterests(ctx context.Context, langCode string) (map[int]string, error) {
	key := fmt.Sprintf("interests:%s", langCode)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var interests map[int]string
	err = json.Unmarshal([]byte(data), &interests)
	return interests, err
}

// SetInterests сохраняет интересы в кэш.
func (c *RedisCache) SetInterests(ctx context.Context, langCode string, interests map[int]string) error {
	key := fmt.Sprintf("interests:%s", langCode)
	return c.Set(ctx, key, interests, 12*time.Hour) // Кэшируем на 12 часов
}

// GetUserProfile получает профиль пользователя из кэша.
func (c *RedisCache) GetUserProfile(ctx context.Context, userID int) (*models.User, error) {
	key := fmt.Sprintf("user:%d", userID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var user models.User
	err = json.Unmarshal([]byte(data), &user)
	return &user, err
}

// SetUserProfile сохраняет профиль пользователя в кэш.
func (c *RedisCache) SetUserProfile(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("user:%d", user.ID)
	return c.Set(ctx, key, user, 30*time.Minute) // Кэшируем на 30 минут
}

// InvalidateUserProfile удаляет профиль пользователя из кэша.
func (c *RedisCache) InvalidateUserProfile(ctx context.Context, userID int) error {
	key := fmt.Sprintf("user:%d", userID)
	return c.Delete(ctx, key)
}

// Close закрывает соединение с Redis.
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// MemoryCache реализация кэша в памяти.
type MemoryCache struct {
	data   map[string]interface{}
	mutex  sync.RWMutex
	ttl    map[string]time.Time
	ctx    context.Context
	cancel context.CancelFunc
}

// NewMemoryCache создает новый кэш в памяти.
func NewMemoryCache() *MemoryCache {
	ctx, cancel := context.WithCancel(context.Background())

	cache := &MemoryCache{
		data:   make(map[string]interface{}),
		ttl:    make(map[string]time.Time),
		ctx:    ctx,
		cancel: cancel,
	}

	// Запускаем горутину для очистки истекших ключей
	go cache.cleanup()

	return cache
}

// cleanup очищает истекшие ключи.
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mutex.Lock()
			now := time.Now()
			for key, expiry := range c.ttl {
				if now.After(expiry) {
					delete(c.data, key)
					delete(c.ttl, key)
				}
			}
			c.mutex.Unlock()
		}
	}
}

// Get получает значение из кэша.
func (c *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, exists := c.data[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	// Проверяем TTL
	if expiry, hasTTL := c.ttl[key]; hasTTL && time.Now().After(expiry) {
		return "", fmt.Errorf("key expired")
	}

	data, err := json.Marshal(value)
	return string(data), err
}

// Set сохраняет значение в кэш.
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = value
	if expiration > 0 {
		c.ttl[key] = time.Now().Add(expiration)
	}
	return nil
}

// Delete удаляет значение из кэша.
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
	delete(c.ttl, key)
	return nil
}

// GetLanguages получает языки из кэша.
func (c *MemoryCache) GetLanguages(ctx context.Context) ([]*models.Language, error) {
	data, err := c.Get(ctx, "languages")
	if err != nil {
		return nil, err
	}

	var languages []*models.Language
	err = json.Unmarshal([]byte(data), &languages)
	return languages, err
}

// SetLanguages сохраняет языки в кэш.
func (c *MemoryCache) SetLanguages(ctx context.Context, languages []*models.Language) error {
	return c.Set(ctx, "languages", languages, 24*time.Hour)
}

// GetInterests получает интересы из кэша.
func (c *MemoryCache) GetInterests(ctx context.Context, langCode string) (map[int]string, error) {
	key := fmt.Sprintf("interests:%s", langCode)
	data, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var interests map[int]string
	err = json.Unmarshal([]byte(data), &interests)
	return interests, err
}

// SetInterests сохраняет интересы в кэш.
func (c *MemoryCache) SetInterests(ctx context.Context, langCode string, interests map[int]string) error {
	key := fmt.Sprintf("interests:%s", langCode)
	return c.Set(ctx, key, interests, 12*time.Hour)
}

// GetUserProfile получает профиль пользователя из кэша.
func (c *MemoryCache) GetUserProfile(ctx context.Context, userID int) (*models.User, error) {
	key := fmt.Sprintf("user:%d", userID)
	data, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = json.Unmarshal([]byte(data), &user)
	return &user, err
}

// SetUserProfile сохраняет профиль пользователя в кэш.
func (c *MemoryCache) SetUserProfile(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("user:%d", user.ID)
	return c.Set(ctx, key, user, 30*time.Minute)
}

// InvalidateUserProfile удаляет профиль пользователя из кэша.
func (c *MemoryCache) InvalidateUserProfile(ctx context.Context, userID int) error {
	key := fmt.Sprintf("user:%d", userID)
	return c.Delete(ctx, key)
}

// Close закрывает кэш.
func (c *MemoryCache) Close() error {
	c.cancel()
	return nil
}
