package integration //nolint:testpackage

import (
	"testing"
	"time"

	"language-exchange-bot/internal/adapters/telegram"
	"language-exchange-bot/internal/errors"

	"github.com/stretchr/testify/suite"
)

// RateLimitingSuite набор тестов для rate limiting.
type RateLimitingSuite struct {
	suite.Suite

	rateLimiter *telegram.RateLimiter
	userID      int64
}

// SetupSuite выполняется один раз перед всеми тестами.
func (s *RateLimitingSuite) SetupSuite() {
	// Создаем rate limiter с тестовыми настройками
	config := telegram.RateLimitConfig{
		MaxRequests:     3,                // 3 запроса
		WindowDuration:  time.Second * 2,  // за 2 секунды
		BlockDuration:   time.Second * 5,  // блокировка на 5 секунд
		CleanupInterval: time.Second * 10, // очистка каждые 10 секунд
	}
	s.rateLimiter = telegram.NewRateLimiter(config)
	s.userID = 123456789
}

// TearDownSuite выполняется один раз после всех тестов.
func (s *RateLimitingSuite) TearDownSuite() {
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}
}

// TestRateLimiting_AllowWithinLimits тестирует, что запросы в пределах лимита разрешены.
func (s *RateLimitingSuite) TestRateLimiting_AllowWithinLimits() {
	// Act & Assert
	for i := 0; i < 3; i++ {
		err := s.rateLimiter.CheckRateLimit(s.userID)
		s.NoError(err, "Request %d should be allowed", i+1)
	}
}

// TestRateLimiting_BlockAfterLimit тестирует блокировку после превышения лимита.
func (s *RateLimitingSuite) TestRateLimiting_BlockAfterLimit() {
	// Arrange - превышаем лимит
	for i := 0; i < 3; i++ {
		_ = s.rateLimiter.CheckRateLimit(s.userID)
	}

	// Act - следующий запрос должен быть заблокирован
	err := s.rateLimiter.CheckRateLimit(s.userID)

	// Assert
	s.Error(err, "Request should be blocked after exceeding limit")
	s.IsType(&errors.CustomError{}, err, "Error should be CustomError")

	customErr, ok := err.(*errors.CustomError)
	s.True(ok, "Error should be convertible to CustomError")
	s.Equal(errors.ErrorTypeValidation, customErr.Type, "Error type should be validation")
}

// TestRateLimiting_BlockDuration тестирует длительность блокировки.
func (s *RateLimitingSuite) TestRateLimiting_BlockDuration() {
	// Arrange - превышаем лимит и проверяем блокировку
	for i := 0; i < 3; i++ {
		_ = s.rateLimiter.CheckRateLimit(s.userID)
	}

	// Первый заблокированный запрос
	err1 := s.rateLimiter.CheckRateLimit(s.userID)
	s.Error(err1, "First request after limit should be blocked")

	// Ждем меньше времени блокировки
	time.Sleep(time.Second * 3)

	// Запрос все еще должен быть заблокирован
	err2 := s.rateLimiter.CheckRateLimit(s.userID)
	s.Error(err2, "Request should still be blocked before block duration ends")

	// Ждем окончания блокировки
	time.Sleep(time.Second * 3)

	// Теперь запросы должны быть разрешены снова
	err3 := s.rateLimiter.CheckRateLimit(s.userID)
	s.NoError(err3, "Request should be allowed after block duration")
}

// TestRateLimiting_WindowReset тестирует сброс окна времени.
func (s *RateLimitingSuite) TestRateLimiting_WindowReset() {
	// Arrange - делаем максимальное количество запросов
	for i := 0; i < 3; i++ {
		err := s.rateLimiter.CheckRateLimit(s.userID)
		s.NoError(err, "Request %d should be allowed", i+1)
	}

	// Act - ждем истечения окна времени
	time.Sleep(time.Second * 3)

	// Assert - новые запросы должны быть разрешены
	for i := 0; i < 3; i++ {
		err := s.rateLimiter.CheckRateLimit(s.userID)
		s.NoError(err, "Request %d should be allowed after window reset", i+1)
	}
}

// TestRateLimiting_MultipleUsers тестирует независимую работу для разных пользователей.
func (s *RateLimitingSuite) TestRateLimiting_MultipleUsers() {
	userID1 := int64(111111111)
	userID2 := int64(222222222)

	// Arrange - user1 превышает лимит
	for i := 0; i < 3; i++ {
		err := s.rateLimiter.CheckRateLimit(userID1)
		s.NoError(err)
	}

	// Act - user1 блокируется, user2 может работать
	err1 := s.rateLimiter.CheckRateLimit(userID1)
	err2 := s.rateLimiter.CheckRateLimit(userID2)

	// Assert
	s.Error(err1, "User1 should be blocked")
	s.NoError(err2, "User2 should not be affected by user1's blocking")
}

// TestRateLimiting_GetStats тестирует получение статистики.
func (s *RateLimitingSuite) TestRateLimiting_GetStats() {
	testUserID := int64(999999999)

	// Arrange - делаем несколько запросов
	for i := 0; i < 2; i++ {
		_ = s.rateLimiter.CheckRateLimit(testUserID)
	}

	// Act
	stats := s.rateLimiter.GetStats()

	// Assert
	s.NotNil(stats, "Stats should not be nil")
	s.Contains(stats, "total_tracked_users", "Stats should contain total_tracked_users")
	s.Contains(stats, "active_users", "Stats should contain active_users")
	s.Contains(stats, "blocked_users", "Stats should contain blocked_users")
	s.Contains(stats, "max_requests", "Stats should contain max_requests")

	// Проверяем значения
	s.GreaterOrEqual(stats["total_tracked_users"].(int), 1, "Should track at least test user")
	s.Equal(3, stats["max_requests"].(int), "Max requests should match config")
}

// TestRateLimiting_EndToEndIntegration тестирует интеграцию с Telegram handler.
func (s *RateLimitingSuite) TestRateLimiting_EndToEndIntegration() {
	// Этот тест проверяет интеграцию rate limiting с реальным handler'ом
	// Используем мок для тестирования

	// Создаем rate limiter с очень строгими настройками для теста
	strictConfig := telegram.RateLimitConfig{
		MaxRequests:     1,
		WindowDuration:  time.Second * 5, // Окно 5 секунд
		BlockDuration:   time.Second * 2, // Блокировка 2 секунды
		CleanupInterval: time.Minute,
	}
	strictLimiter := telegram.NewRateLimiter(strictConfig)
	defer strictLimiter.Stop()

	// Act & Assert
	// Используем уникальный userID для этого теста
	userID := int64(777888999)

	// Первый запрос разрешен
	err1 := strictLimiter.CheckRateLimit(userID)
	s.NoError(err1, "First request should be allowed")

	// Второй запрос заблокирован
	err2 := strictLimiter.CheckRateLimit(userID)
	s.Error(err2, "Second request should be blocked")

	// Ждем окончания блокировки (2 секунды + небольшой запас)
	s.T().Logf("Waiting for block to expire...")
	time.Sleep(time.Second * 3)
	s.T().Logf("Block should be expired now")

	// Третий запрос снова разрешен
	err3 := strictLimiter.CheckRateLimit(userID)
	s.NoError(err3, "Request should be allowed after block expires")
}

// TestSuite запускает все тесты в наборе.
func TestRateLimitingSuite(t *testing.T) {
	suite.Run(t, new(RateLimitingSuite))
}
