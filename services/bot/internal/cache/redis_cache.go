package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"language-exchange-bot/internal/models"

	"github.com/redis/go-redis/v9"
)

// RedisCacheService реализация кэша на основе Redis
type RedisCacheService struct {
	client *redis.Client
	config *CacheConfig
	ctx    context.Context
}

// NewRedisCacheService создает новый Redis кэш-сервис
func NewRedisCacheService(redisURL, password string, db int, config *CacheConfig) (*RedisCacheService, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Проверяем подключение
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis cache service initialized: %s (DB: %d)", redisURL, db)

	return &RedisCacheService{
		client: client,
		config: config,
		ctx:    ctx,
	}, nil
}

// GetLanguages получает языки из Redis кэша
func (r *RedisCacheService) GetLanguages(lang string) ([]*models.Language, bool) {
	key := fmt.Sprintf("languages:%s", lang)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
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

// SetLanguages сохраняет языки в Redis кэш
func (r *RedisCacheService) SetLanguages(lang string, languages []*models.Language) {
	key := fmt.Sprintf("languages:%s", lang)

	data, err := json.Marshal(languages)
	if err != nil {
		log.Printf("Redis error marshaling languages: %v", err)
		return
	}

	err = r.client.Set(r.ctx, key, data, r.config.LanguagesTTL).Err()
	if err != nil {
		log.Printf("Redis error setting languages: %v", err)
	}
}

// GetInterests получает интересы из Redis кэша
func (r *RedisCacheService) GetInterests(lang string) (map[int]string, bool) {
	key := fmt.Sprintf("interests:%s", lang)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
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

// SetInterests сохраняет интересы в Redis кэш
func (r *RedisCacheService) SetInterests(lang string, interests map[int]string) {
	key := fmt.Sprintf("interests:%s", lang)

	data, err := json.Marshal(interests)
	if err != nil {
		log.Printf("Redis error marshaling interests: %v", err)
		return
	}

	err = r.client.Set(r.ctx, key, data, r.config.InterestsTTL).Err()
	if err != nil {
		log.Printf("Redis error setting interests: %v", err)
	}
}

// GetUser получает пользователя из Redis кэша
func (r *RedisCacheService) GetUser(userID int64) (*models.User, bool) {
	key := fmt.Sprintf("user:%d", userID)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
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

// SetUser сохраняет пользователя в Redis кэш
func (r *RedisCacheService) SetUser(user *models.User) {
	key := fmt.Sprintf("user:%d", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		log.Printf("Redis error marshaling user: %v", err)
		return
	}

	err = r.client.Set(r.ctx, key, data, r.config.UsersTTL).Err()
	if err != nil {
		log.Printf("Redis error setting user: %v", err)
	}
}

// GetTranslations получает переводы из Redis кэша
func (r *RedisCacheService) GetTranslations(lang string) (map[string]string, bool) {
	key := fmt.Sprintf("translations:%s", lang)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
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

// SetTranslations сохраняет переводы в Redis кэш
func (r *RedisCacheService) SetTranslations(lang string, translations map[string]string) {
	key := fmt.Sprintf("translations:%s", lang)

	data, err := json.Marshal(translations)
	if err != nil {
		log.Printf("Redis error marshaling translations: %v", err)
		return
	}

	err = r.client.Set(r.ctx, key, data, r.config.TranslationsTTL).Err()
	if err != nil {
		log.Printf("Redis error setting translations: %v", err)
	}
}

// GetStats получает статистику из Redis кэша
func (r *RedisCacheService) GetStats(statsType string) (map[string]interface{}, bool) {
	key := fmt.Sprintf("stats:%s", statsType)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
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

// SetStats сохраняет статистику в Redis кэш
func (r *RedisCacheService) SetStats(statsType string, data map[string]interface{}) {
	key := fmt.Sprintf("stats:%s", statsType)

	statsData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Redis error marshaling stats: %v", err)
		return
	}

	err = r.client.Set(r.ctx, key, statsData, r.config.StatsTTL).Err()
	if err != nil {
		log.Printf("Redis error setting stats: %v", err)
	}
}

// InvalidateUser удаляет пользователя из Redis кэша
func (r *RedisCacheService) InvalidateUser(userID int64) {
	key := fmt.Sprintf("user:%d", userID)

	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		log.Printf("Redis error deleting user: %v", err)
	} else {
		log.Printf("Redis: Invalidated user %d", userID)
	}
}

// InvalidateLanguages удаляет языки из Redis кэша
func (r *RedisCacheService) InvalidateLanguages() {
	pattern := "languages:*"

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		log.Printf("Redis error getting language keys: %v", err)
		return
	}

	if len(keys) > 0 {
		err = r.client.Del(r.ctx, keys...).Err()
		if err != nil {
			log.Printf("Redis error deleting languages: %v", err)
		} else {
			log.Printf("Redis: Invalidated %d language entries", len(keys))
		}
	}
}

// InvalidateInterests удаляет интересы из Redis кэша
func (r *RedisCacheService) InvalidateInterests() {
	pattern := "interests:*"

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		log.Printf("Redis error getting interest keys: %v", err)
		return
	}

	if len(keys) > 0 {
		err = r.client.Del(r.ctx, keys...).Err()
		if err != nil {
			log.Printf("Redis error deleting interests: %v", err)
		} else {
			log.Printf("Redis: Invalidated %d interest entries", len(keys))
		}
	}
}

// InvalidateTranslations удаляет переводы из Redis кэша
func (r *RedisCacheService) InvalidateTranslations() {
	pattern := "translations:*"

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		log.Printf("Redis error getting translation keys: %v", err)
		return
	}

	if len(keys) > 0 {
		err = r.client.Del(r.ctx, keys...).Err()
		if err != nil {
			log.Printf("Redis error deleting translations: %v", err)
		} else {
			log.Printf("Redis: Invalidated %d translation entries", len(keys))
		}
	}
}

// ClearAll очищает весь Redis кэш
func (r *RedisCacheService) ClearAll() {
	err := r.client.FlushDB(r.ctx).Err()
	if err != nil {
		log.Printf("Redis error clearing all: %v", err)
	} else {
		log.Printf("Redis: Cleared all data")
	}
}

// GetCacheStats возвращает статистику Redis кэша
func (r *RedisCacheService) GetCacheStats() CacheStats {
	// Получаем количество ключей
	keys, err := r.client.DBSize(r.ctx).Result()
	if err != nil {
		log.Printf("Redis error getting key count: %v", err)
		return CacheStats{}
	}

	return CacheStats{
		Hits:   0, // Redis не предоставляет hit/miss статистику по умолчанию
		Misses: 0,
		Size:   int(keys),
	}
}

// Stop останавливает Redis кэш-сервис
func (r *RedisCacheService) Stop() {
	err := r.client.Close()
	if err != nil {
		log.Printf("Redis error closing connection: %v", err)
	} else {
		log.Printf("Redis: Service stopped")
	}
}

// String возвращает строковое представление статистики Redis кэша
func (r *RedisCacheService) String() string {
	stats := r.GetCacheStats()
	return fmt.Sprintf("Redis Cache Stats: Size=%d", stats.Size)
}
