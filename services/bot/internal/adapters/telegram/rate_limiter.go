package telegram

import (
	"fmt"
	"sync"
	"time"

	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/localization"
)

// RateLimitConfig конфигурация rate limiter'а.
type RateLimitConfig struct {
	MaxRequests     int           // Максимальное количество запросов
	WindowDuration  time.Duration // Период времени для подсчета
	BlockDuration   time.Duration // Длительность блокировки при превышении
	CleanupInterval time.Duration // Интервал очистки устаревших записей
}

// DefaultRateLimitConfig возвращает конфигурацию по умолчанию для защиты от спама
// Настройки подобраны для типичного использования: пиковые нагрузки до 20 сообщений/минуту,
// но обычно гораздо медленнее. При превышении - короткая блокировка для коррекции поведения.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests:     20,                                                 // 20 запросов (мягкий лимит для пиковых нагрузок)
		WindowDuration:  localization.RateLimitWindowMinutes * time.Minute,  // в минуту
		BlockDuration:   localization.RateLimitBlockMinutes * time.Minute,   // блокировка на 2 минуты (короткая для мягкого режима)
		CleanupInterval: localization.RateLimitCleanupMinutes * time.Minute, // очистка каждые 10 минут
	}
}

// UserRateLimit информация о rate limit для пользователя.
type UserRateLimit struct {
	RequestCount int       // Количество запросов
	FirstRequest time.Time // Время первого запроса в окне
	BlockUntil   time.Time // Время окончания блокировки
}

// RateLimiter реализация rate limiting для защиты от спама.
type RateLimiter struct {
	config     RateLimitConfig
	userLimits map[int64]*UserRateLimit
	mutex      sync.RWMutex
	stopChan   chan struct{}
}

// NewRateLimiter создает новый rate limiter.
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:     config,
		userLimits: make(map[int64]*UserRateLimit),
		stopChan:   make(chan struct{}),
	}

	// Запускаем горутину для очистки устаревших записей
	go rl.cleanupWorker()

	return rl
}

// IsAllowed проверяет, разрешен ли запрос от пользователя.
func (rl *RateLimiter) IsAllowed(userID int64) (bool, time.Duration) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	userLimit, exists := rl.userLimits[userID]

	if !exists {
		// Первый запрос от пользователя
		rl.userLimits[userID] = &UserRateLimit{
			RequestCount: 1,
			FirstRequest: now,
		}

		return true, 0
	}

	// Проверяем, заблокирован ли пользователь
	if now.Before(userLimit.BlockUntil) {
		remaining := userLimit.BlockUntil.Sub(now)

		return false, remaining
	}

	// Проверяем, истекло ли окно времени
	windowExpired := now.Sub(userLimit.FirstRequest) >= rl.config.WindowDuration

	// Если окно истекло, полностью сбрасываем
	if windowExpired {
		userLimit.RequestCount = 1 // Сбрасываем на 1 (текущий запрос)
		userLimit.FirstRequest = now
		userLimit.BlockUntil = time.Time{}

		return true, 0
	}

	// Проверяем, была ли недавняя блокировка (истекла в этом окне)
	// Если да, даем "второй шанс" - сбрасываем счетчик
	if !userLimit.BlockUntil.IsZero() && now.After(userLimit.BlockUntil) {
		userLimit.RequestCount = 1         // Сбрасываем на 1 (второй шанс)
		userLimit.BlockUntil = time.Time{} // Сбрасываем блокировку

		return true, 0
	}

	// Окно активно, увеличиваем счетчик
	userLimit.RequestCount++

	// Проверяем, превышен ли лимит
	if userLimit.RequestCount > rl.config.MaxRequests {
		userLimit.BlockUntil = now.Add(rl.config.BlockDuration)
		remaining := rl.config.BlockDuration

		return false, remaining
	}

	return true, 0
}

// GetStats возвращает статистику rate limiter'а.
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	now := time.Now()
	activeUsers := 0
	blockedUsers := 0

	for _, limit := range rl.userLimits {
		if now.Sub(limit.FirstRequest) < rl.config.WindowDuration*2 { // Активные за последние 2 окна
			activeUsers++

			if now.Before(limit.BlockUntil) {
				blockedUsers++
			}
		}
	}

	return map[string]interface{}{
		"total_tracked_users": len(rl.userLimits),
		"active_users":        activeUsers,
		"blocked_users":       blockedUsers,
		"max_requests":        rl.config.MaxRequests,
		"window_duration":     rl.config.WindowDuration.String(),
		"block_duration":      rl.config.BlockDuration.String(),
	}
}

// cleanupWorker периодически очищает устаревшие записи.
func (rl *RateLimiter) cleanupWorker() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanup удаляет устаревшие записи пользователей.
func (rl *RateLimiter) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.config.WindowDuration * 2) // Удаляем записи старше 2 окон

	for userID, limit := range rl.userLimits {
		if limit.FirstRequest.Before(cutoff) && now.After(limit.BlockUntil) {
			delete(rl.userLimits, userID)
		}
	}
}

// Stop останавливает rate limiter.
func (rl *RateLimiter) Stop() {
	if rl.stopChan != nil {
		select {
		case <-rl.stopChan:
			// Канал уже закрыт
		default:
			close(rl.stopChan)
		}
	}
}

// CheckRateLimit проверяет rate limit и возвращает ошибку если превышен.
func (rl *RateLimiter) CheckRateLimit(userID int64) error {
	allowed, remaining := rl.IsAllowed(userID)

	if !allowed {
		return errors.NewCustomError(
			errors.ErrorTypeValidation,
			fmt.Sprintf("Rate limit exceeded. Try again in %v", remaining.Round(time.Second)),
			"Слишком много запросов. Попробуйте позже",
			"",
		)
	}

	return nil
}
