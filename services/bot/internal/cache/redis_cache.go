package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"

	"github.com/redis/go-redis/v9"
)

// Redis connection constants (using centralized constants for consistency)
var (
	DefaultDialTimeout     = localization.RedisDialTimeoutSeconds * time.Second
	DefaultReadTimeout     = localization.RedisReadTimeoutSeconds * time.Second
	DefaultWriteTimeout    = localization.RedisWriteTimeoutSeconds * time.Second
	DefaultMinRetryBackoff = localization.RedisMinRetryBackoffMs * time.Millisecond
	DefaultMaxRetryBackoff = localization.RedisMaxRetryBackoffMs * time.Millisecond
)

// Redis configuration constants
const (
	DefaultRedisProtocol = 3
	DefaultMaxRetries    = 3
	DefaultPoolSize      = 10
)

// RedisCacheService реализация кэша на основе Redis.
type RedisCacheService struct {
	client *redis.Client
	config *Config
}

// NewRedisCacheService создает новый Redis кэш-сервис.
func NewRedisCacheService(redisURL, password string, database int, config *Config) (*RedisCacheService, error) {
	if config == nil {
		config = DefaultConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:                         redisURL,
		Password:                     password,
		DB:                           database,
		Network:                      "tcp",
		ClientName:                   "",
		Dialer:                       nil,
		OnConnect:                    nil,
		Protocol:                     DefaultRedisProtocol,
		Username:                     "",
		CredentialsProvider:          nil,
		CredentialsProviderContext:   nil,
		StreamingCredentialsProvider: nil,
		MaxRetries:                   DefaultMaxRetries,
		MinRetryBackoff:              DefaultMinRetryBackoff,
		MaxRetryBackoff:              DefaultMaxRetryBackoff,
		DialTimeout:                  DefaultDialTimeout,
		ReadTimeout:                  DefaultReadTimeout,
		WriteTimeout:                 DefaultWriteTimeout,
		ContextTimeoutEnabled:        false,
		ReadBufferSize:               0,
		WriteBufferSize:              0,
		PoolFIFO:                     false,
		PoolSize:                     DefaultPoolSize,
		PoolTimeout:                  0,
		MinIdleConns:                 0,
		MaxIdleConns:                 0,
		MaxActiveConns:               0,
		ConnMaxIdleTime:              0,
		ConnMaxLifetime:              0,
		TLSConfig:                    nil,
		Limiter:                      nil,
		DisableIndentity:             false,
		DisableIdentity:              false,
		IdentitySuffix:               "",
		UnstableResp3:                false,
		FailingTimeoutSeconds:        0,
	})

	ctx := context.Background()

	// Проверяем подключение
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis cache service initialized: %s (DB: %d)", redisURL, database)

	return &RedisCacheService{
		client: client,
		config: config,
	}, nil
}

// GetLanguages получает языки из Redis кэша.
func (r *RedisCacheService) GetLanguages(ctx context.Context, lang string) ([]*models.Language, bool) {
	key := "languages:" + lang

	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false
	}

	if err != nil {
		log.Printf("Redis error getting languages: %v", err)

		return nil, false
	}

	var languages []*models.Language
	if err := json.Unmarshal([]byte(val), &languages); err != nil {
		log.Printf("Redis error unmarshaling languages: %v", err)

		return nil, false
	}

	return languages, true
}

// SetLanguages сохраняет языки в Redis кэш.
func (r *RedisCacheService) SetLanguages(ctx context.Context, lang string, languages []*models.Language) {
	key := "languages:" + lang

	data, err := json.Marshal(languages)
	if err != nil {
		log.Printf("Redis error marshaling languages: %v", err)

		return
	}

	err = r.client.Set(ctx, key, data, r.config.LanguagesTTL).Err()
	if err != nil {
		log.Printf("Redis error setting languages: %v", err)
	}
}

// GetInterests получает интересы из Redis кэша.
func (r *RedisCacheService) GetInterests(ctx context.Context, lang string) (map[int]string, bool) {
	key := "interests:" + lang

	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false
	}

	if err != nil {
		log.Printf("Redis error getting interests: %v", err)

		return nil, false
	}

	var interests map[int]string
	if err := json.Unmarshal([]byte(val), &interests); err != nil {
		log.Printf("Redis error unmarshaling interests: %v", err)

		return nil, false
	}

	return interests, true
}

// SetInterests сохраняет интересы в Redis кэш.
func (r *RedisCacheService) SetInterests(ctx context.Context, lang string, interests map[int]string) {
	key := "interests:" + lang

	data, err := json.Marshal(interests)
	if err != nil {
		log.Printf("Redis error marshaling interests: %v", err)

		return
	}

	err = r.client.Set(ctx, key, data, r.config.InterestsTTL).Err()
	if err != nil {
		log.Printf("Redis error setting interests: %v", err)
	}
}

// GetUser получает пользователя из Redis кэша.
func (r *RedisCacheService) GetUser(ctx context.Context, userID int64) (*models.User, bool) {
	key := fmt.Sprintf("user:%d", userID)

	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false
	}

	if err != nil {
		log.Printf("Redis error getting user: %v", err)

		return nil, false
	}

	var user models.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		log.Printf("Redis error unmarshaling user: %v", err)

		return nil, false
	}

	return &user, true
}

// SetUser сохраняет пользователя в Redis кэш.
func (r *RedisCacheService) SetUser(ctx context.Context, user *models.User) {
	key := fmt.Sprintf("user:%d", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		log.Printf("Redis error marshaling user: %v", err)

		return
	}

	err = r.client.Set(ctx, key, data, r.config.UsersTTL).Err()
	if err != nil {
		log.Printf("Redis error setting user: %v", err)
	}
}

// GetTranslations получает переводы из Redis кэша.
func (r *RedisCacheService) GetTranslations(ctx context.Context, lang string) (map[string]string, bool) {
	key := "translations:" + lang

	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false
	}

	if err != nil {
		log.Printf("Redis error getting translations: %v", err)

		return nil, false
	}

	var translations map[string]string
	if err := json.Unmarshal([]byte(val), &translations); err != nil {
		log.Printf("Redis error unmarshaling translations: %v", err)

		return nil, false
	}

	return translations, true
}

// SetTranslations сохраняет переводы в Redis кэш.
func (r *RedisCacheService) SetTranslations(ctx context.Context, lang string, translations map[string]string) {
	key := "translations:" + lang

	data, err := json.Marshal(translations)
	if err != nil {
		log.Printf("Redis error marshaling translations: %v", err)

		return
	}

	err = r.client.Set(ctx, key, data, r.config.TranslationsTTL).Err()
	if err != nil {
		log.Printf("Redis error setting translations: %v", err)
	}
}

// GetStats получает статистику из Redis кэша.
func (r *RedisCacheService) GetStats(ctx context.Context, statsType string) (map[string]interface{}, bool) {
	key := "stats:" + statsType

	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false
	}

	if err != nil {
		log.Printf("Redis error getting stats: %v", err)

		return nil, false
	}

	var stats map[string]interface{}
	if err := json.Unmarshal([]byte(val), &stats); err != nil {
		log.Printf("Redis error unmarshaling stats: %v", err)

		return nil, false
	}

	return stats, true
}

// SetStats сохраняет статистику в Redis кэш.
func (r *RedisCacheService) SetStats(ctx context.Context, statsType string, data map[string]interface{}) {
	key := "stats:" + statsType

	statsData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Redis error marshaling stats: %v", err)

		return
	}

	err = r.client.Set(ctx, key, statsData, r.config.StatsTTL).Err()
	if err != nil {
		log.Printf("Redis error setting stats: %v", err)
	}
}

// InvalidateUser удаляет пользователя из Redis кэша.
func (r *RedisCacheService) InvalidateUser(ctx context.Context, userID int64) {
	key := fmt.Sprintf("user:%d", userID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Redis error deleting user: %v", err)
	} else {
		log.Printf("Redis: Invalidated user %d", userID)
	}
}

// InvalidateLanguages удаляет языки из Redis кэша.
func (r *RedisCacheService) InvalidateLanguages(ctx context.Context) {
	pattern := "languages:*"

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Redis error getting language keys: %v", err)

		return
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			log.Printf("Redis error deleting languages: %v", err)
		} else {
			log.Printf("Redis: Invalidated %d language entries", len(keys))
		}
	}
}

// InvalidateInterests удаляет интересы из Redis кэша.
func (r *RedisCacheService) InvalidateInterests(ctx context.Context) {
	pattern := "interests:*"

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Redis error getting interest keys: %v", err)

		return
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			log.Printf("Redis error deleting interests: %v", err)
		} else {
			log.Printf("Redis: Invalidated %d interest entries", len(keys))
		}
	}
}

// InvalidateTranslations удаляет переводы из Redis кэша.
func (r *RedisCacheService) InvalidateTranslations(ctx context.Context) {
	pattern := "translations:*"

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		log.Printf("Redis error getting translation keys: %v", err)

		return
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			log.Printf("Redis error deleting translations: %v", err)
		} else {
			log.Printf("Redis: Invalidated %d translation entries", len(keys))
		}
	}
}

// ClearAll очищает весь Redis кэш.
func (r *RedisCacheService) ClearAll(ctx context.Context) {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		log.Printf("Redis error clearing all: %v", err)
	} else {
		log.Printf("Redis: Cleared all data")
	}
}

// GetCacheStats возвращает статистику Redis кэша.
func (r *RedisCacheService) GetCacheStats(ctx context.Context) Stats {
	// Получаем количество ключей
	keys, err := r.client.DBSize(ctx).Result()
	if err != nil {
		log.Printf("Redis error getting key count: %v", err)

		return Stats{
			Hits:   0,
			Misses: 0,
			Size:   0,
		}
	}

	return Stats{
		Hits:   0, // Redis не предоставляет hit/miss статистику по умолчанию
		Misses: 0,
		Size:   int(keys),
	}
}

// Stop останавливает Redis кэш-сервис.
func (r *RedisCacheService) Stop() {
	err := r.client.Close()
	if err != nil {
		log.Printf("Redis error closing connection: %v", err)
	} else {
		log.Printf("Redis: Service stopped")
	}
}

// String возвращает строковое представление статистики Redis кэша.
func (r *RedisCacheService) String() string {
	stats := r.GetCacheStats(context.Background())

	return fmt.Sprintf("Redis Cache Stats: Size=%d", stats.Size)
}

// Set сохраняет произвольные данные в Redis кэш.
func (r *RedisCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}

	return nil
}

// Get получает произвольные данные из Redis кэша.
func (r *RedisCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("key not found: %s", key)
	}

	if err != nil {
		return fmt.Errorf("failed to get value from Redis: %w", err)
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete удаляет ключ из Redis кэша.
func (r *RedisCacheService) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}

	return nil
}

// ===== НОВЫЕ МЕТОДЫ КЕШИРОВАНИЯ =====

// GetInterestCategories получает категории интересов из Redis кэша.
func (r *RedisCacheService) GetInterestCategories(ctx context.Context, lang string) ([]*models.InterestCategory, bool) {
	key := fmt.Sprintf("interest_categories:%s", lang)

	var categories []*models.InterestCategory
	err := r.Get(ctx, key, &categories)
	if err != nil {
		return nil, false
	}

	return categories, true
}

// SetInterestCategories сохраняет категории интересов в Redis кэш.
func (r *RedisCacheService) SetInterestCategories(ctx context.Context, lang string, categories []*models.InterestCategory) {
	key := fmt.Sprintf("interest_categories:%s", lang)
	_ = r.Set(ctx, key, categories, r.config.LanguagesTTL)
}

// GetUserStats получает статистику пользователя из Redis кэша.
func (r *RedisCacheService) GetUserStats(ctx context.Context, userID int64) (map[string]interface{}, bool) {
	key := fmt.Sprintf("user_stats:%d", userID)

	var stats map[string]interface{}
	err := r.Get(ctx, key, &stats)
	if err != nil {
		return nil, false
	}

	return stats, true
}

// SetUserStats сохраняет статистику пользователя в Redis кэш.
func (r *RedisCacheService) SetUserStats(ctx context.Context, userID int64, stats map[string]interface{}) {
	key := fmt.Sprintf("user_stats:%d", userID)
	_ = r.Set(ctx, key, stats, r.config.UsersTTL)
}

// GetConfig получает конфигурацию из Redis кэша.
func (r *RedisCacheService) GetConfig(ctx context.Context, configKey string) (interface{}, bool) {
	key := fmt.Sprintf("config:%s", configKey)

	var value interface{}
	err := r.Get(ctx, key, &value)
	if err != nil {
		return nil, false
	}

	return value, true
}

// SetConfig сохраняет конфигурацию в Redis кэш.
func (r *RedisCacheService) SetConfig(ctx context.Context, configKey string, value interface{}) {
	key := fmt.Sprintf("config:%s", configKey)
	_ = r.Set(ctx, key, value, r.config.LanguagesTTL)
}

// InvalidateInterestCategories инвалидирует кэш категорий интересов.
func (r *RedisCacheService) InvalidateInterestCategories(ctx context.Context) {
	pattern := "interest_categories:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return
	}

	if len(keys) > 0 {
		r.client.Del(ctx, keys...)
	}
}

// InvalidateUserStats инвалидирует кэш статистики пользователя.
func (r *RedisCacheService) InvalidateUserStats(ctx context.Context, userID int64) {
	key := fmt.Sprintf("user_stats:%d", userID)
	r.client.Del(ctx, key)
}
