package core

import (
	"context"
	"fmt"
	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
	"log"
	"sync"
	"time"
)

// OptimizedBotService оптимизированный сервис с кэшированием и batch операциями
type OptimizedBotService struct {
	db        *database.OptimizedDB
	cache     cache.Cache
	localizer *localization.Localizer
	ctx       context.Context
	cancel    context.CancelFunc

	// Кэш для часто используемых данных
	languagesCache map[string][]*models.Language
	interestsCache map[string]map[int]string
	cacheMutex     sync.RWMutex

	// Метрики
	metrics *ServiceMetrics
}

// ServiceMetrics метрики производительности
type ServiceMetrics struct {
	CacheHits           int64
	CacheMisses         int64
	DBQueries           int64
	BatchOperations     int64
	AverageResponseTime time.Duration
	mutex               sync.RWMutex
}

// NewOptimizedBotService создает оптимизированный сервис
func NewOptimizedBotService(db *database.OptimizedDB, cache cache.Cache) *OptimizedBotService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &OptimizedBotService{
		db:             db,
		cache:          cache,
		localizer:      localization.NewLocalizer(nil), // Будет использовать кэш
		ctx:            ctx,
		cancel:         cancel,
		languagesCache: make(map[string][]*models.Language),
		interestsCache: make(map[string]map[int]string),
		metrics:        &ServiceMetrics{},
	}

	// Предзагружаем часто используемые данные
	go service.preloadData()

	return service
}

// preloadData предзагружает часто используемые данные
func (s *OptimizedBotService) preloadData() {
	// Предзагружаем языки
	if _, err := s.GetLanguages(); err != nil {
		log.Printf("Failed to preload languages: %v", err)
	}

	// Предзагружаем интересы для основных языков
	languages := []string{"en", "ru", "es", "zh"}
	for _, lang := range languages {
		if _, err := s.GetLocalizedInterests(lang); err != nil {
			log.Printf("Failed to preload interests for %s: %v", lang, err)
		}
	}
}

// HandleUserRegistration оптимизированная регистрация пользователя
func (s *OptimizedBotService) HandleUserRegistration(telegramID int64, username, firstName, telegramLangCode string) (*models.User, error) {
	start := time.Now()
	defer func() {
		s.updateMetrics(time.Since(start))
	}()

	// Проверяем кэш пользователя
	user, err := s.getUserFromCache(telegramID)
	if err == nil && user != nil {
		s.incrementCacheHits()
		return user, nil
	}
	s.incrementCacheMisses()

	// Создаем/находим пользователя в БД
	user, err = s.db.FindOrCreateUser(telegramID, username, firstName)
	if err != nil {
		return nil, err
	}

	// Определяем язык интерфейса
	detected := s.DetectLanguage(telegramLangCode)
	if user.Status == models.StatusNew || user.InterfaceLanguageCode == "" {
		if detected == "" {
			user.InterfaceLanguageCode = "ru"
		} else {
			user.InterfaceLanguageCode = detected
		}

		// Обновляем язык интерфейса
		updates := map[string]interface{}{
			"interface_language_code": user.InterfaceLanguageCode,
		}
		if err := s.db.UpdateUserProfileBatch(user.ID, updates); err != nil {
			log.Printf("Failed to update interface language: %v", err)
		}
	}

	// Сохраняем в кэш
	s.cacheUser(user)

	return user, nil
}

// UpdateUserProfileBatch обновляет профиль пользователя batch операцией
func (s *OptimizedBotService) UpdateUserProfileBatch(userID int, updates map[string]interface{}) error {
	start := time.Now()
	defer func() {
		s.updateMetrics(time.Since(start))
		s.incrementBatchOperations()
	}()

	err := s.db.UpdateUserProfileBatch(userID, updates)
	if err != nil {
		return err
	}

	// Инвалидируем кэш пользователя
	s.cache.InvalidateUserProfile(s.ctx, userID)

	return nil
}

// GetLanguages получает языки с кэшированием
func (s *OptimizedBotService) GetLanguages() ([]*models.Language, error) {
	start := time.Now()
	defer func() {
		s.updateMetrics(time.Since(start))
	}()

	// Проверяем кэш
	s.cacheMutex.RLock()
	if languages, exists := s.languagesCache["all"]; exists {
		s.cacheMutex.RUnlock()
		s.incrementCacheHits()
		return languages, nil
	}
	s.cacheMutex.RUnlock()

	// Проверяем Redis кэш
	languages, err := s.cache.GetLanguages(s.ctx)
	if err == nil {
		s.cacheMutex.Lock()
		s.languagesCache["all"] = languages
		s.cacheMutex.Unlock()
		s.incrementCacheHits()
		return languages, nil
	}
	s.incrementCacheMisses()

	// Загружаем из БД (заглушка - в реальном проекте здесь был бы запрос к БД)
	languages = []*models.Language{
		{ID: 1, Code: "en", NameNative: "English", NameEn: "English"},
		{ID: 2, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
		{ID: 3, Code: "es", NameNative: "Español", NameEn: "Spanish"},
		{ID: 4, Code: "zh", NameNative: "中文", NameEn: "Chinese"},
	}

	// Сохраняем в кэш
	s.cache.SetLanguages(s.ctx, languages)
	s.cacheMutex.Lock()
	s.languagesCache["all"] = languages
	s.cacheMutex.Unlock()

	return languages, nil
}

// GetLocalizedInterests получает локализованные интересы с кэшированием
func (s *OptimizedBotService) GetLocalizedInterests(langCode string) (map[int]string, error) {
	start := time.Now()
	defer func() {
		s.updateMetrics(time.Since(start))
	}()

	// Проверяем кэш
	s.cacheMutex.RLock()
	if interests, exists := s.interestsCache[langCode]; exists {
		s.cacheMutex.RUnlock()
		s.incrementCacheHits()
		return interests, nil
	}
	s.cacheMutex.RUnlock()

	// Проверяем Redis кэш
	interests, err := s.cache.GetInterests(s.ctx, langCode)
	if err == nil {
		s.cacheMutex.Lock()
		s.interestsCache[langCode] = interests
		s.cacheMutex.Unlock()
		s.incrementCacheHits()
		return interests, nil
	}
	s.incrementCacheMisses()

	// Загружаем из локализатора
	interests, err = s.localizer.GetInterests(langCode)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	s.cache.SetInterests(s.ctx, langCode, interests)
	s.cacheMutex.Lock()
	s.interestsCache[langCode] = interests
	s.cacheMutex.Unlock()

	return interests, nil
}

// SaveUserInterestsBatch сохраняет интересы пользователя batch операцией
func (s *OptimizedBotService) SaveUserInterestsBatch(userID int, interestIDs []int) error {
	start := time.Now()
	defer func() {
		s.updateMetrics(time.Since(start))
		s.incrementBatchOperations()
	}()

	err := s.db.SaveUserInterestsBatch(userID, interestIDs)
	if err != nil {
		return err
	}

	// Инвалидируем кэш пользователя
	s.cache.InvalidateUserProfile(s.ctx, userID)

	return nil
}

// GetUnprocessedFeedbackBatch получает необработанные отзывы с пагинацией
func (s *OptimizedBotService) GetUnprocessedFeedbackBatch(limit, offset int) ([]map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		s.updateMetrics(time.Since(start))
	}()

	return s.db.GetUnprocessedFeedbackBatch(limit, offset)
}

// MarkFeedbackProcessedBatch помечает несколько отзывов как обработанные
func (s *OptimizedBotService) MarkFeedbackProcessedBatch(feedbackIDs []int, adminResponse string) error {
	start := time.Now()
	defer func() {
		s.updateMetrics(time.Since(start))
		s.incrementBatchOperations()
	}()

	return s.db.MarkFeedbackProcessedBatch(feedbackIDs, adminResponse)
}

// getUserFromCache получает пользователя из кэша
func (s *OptimizedBotService) getUserFromCache(telegramID int64) (*models.User, error) {
	// Здесь должен быть запрос к БД по telegram_id для получения user_id
	// Для упрощения возвращаем nil
	return nil, fmt.Errorf("user not in cache")
}

// cacheUser сохраняет пользователя в кэш
func (s *OptimizedBotService) cacheUser(user *models.User) {
	if err := s.cache.SetUserProfile(s.ctx, user); err != nil {
		log.Printf("Failed to cache user: %v", err)
	}
}

// updateMetrics обновляет метрики
func (s *OptimizedBotService) updateMetrics(duration time.Duration) {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()

	s.metrics.AverageResponseTime = (s.metrics.AverageResponseTime + duration) / 2
}

// incrementCacheHits увеличивает счетчик попаданий в кэш
func (s *OptimizedBotService) incrementCacheHits() {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()
	s.metrics.CacheHits++
}

// incrementCacheMisses увеличивает счетчик промахов кэша
func (s *OptimizedBotService) incrementCacheMisses() {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()
	s.metrics.CacheMisses++
}

// incrementBatchOperations увеличивает счетчик batch операций
func (s *OptimizedBotService) incrementBatchOperations() {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()
	s.metrics.BatchOperations++
}

// GetMetrics возвращает метрики сервиса
func (s *OptimizedBotService) GetMetrics() *ServiceMetrics {
	s.metrics.mutex.RLock()
	defer s.metrics.mutex.RUnlock()

	return &ServiceMetrics{
		CacheHits:           s.metrics.CacheHits,
		CacheMisses:         s.metrics.CacheMisses,
		DBQueries:           s.metrics.DBQueries,
		BatchOperations:     s.metrics.BatchOperations,
		AverageResponseTime: s.metrics.AverageResponseTime,
	}
}

// GetDBStats возвращает статистику БД
func (s *OptimizedBotService) GetDBStats() map[string]interface{} {
	return s.db.GetStats()
}

// HealthCheck проверяет здоровье сервиса
func (s *OptimizedBotService) HealthCheck() error {
	// Проверяем БД
	if err := s.db.HealthCheck(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Проверяем кэш
	if err := s.cache.Set(s.ctx, "health_check", "ok", time.Second); err != nil {
		return fmt.Errorf("cache health check failed: %w", err)
	}

	return nil
}

// Close закрывает сервис
func (s *OptimizedBotService) Close() error {
	s.cancel()

	if err := s.cache.Close(); err != nil {
		log.Printf("Failed to close cache: %v", err)
	}

	if err := s.db.Close(); err != nil {
		log.Printf("Failed to close database: %v", err)
	}

	return nil
}

// DetectLanguage определяет язык пользователя (копия из оригинального сервиса)
func (s *OptimizedBotService) DetectLanguage(telegramLangCode string) string {
	switch telegramLangCode {
	case "ru", "ru-RU":
		return "ru"
	case "es", "es-ES", "es-MX":
		return "es"
	case "zh", "zh-CN", "zh-TW":
		return "zh"
	default:
		return "en"
	}
}

// GetWelcomeMessage получает приветственное сообщение (копия из оригинального сервиса)
func (s *OptimizedBotService) GetWelcomeMessage(user *models.User) string {
	return s.localizer.GetWithParams(user.InterfaceLanguageCode, "welcome_message", map[string]string{
		"name": user.FirstName,
	})
}

// GetLanguagePrompt получает промпт для выбора языка (копия из оригинального сервиса)
func (s *OptimizedBotService) GetLanguagePrompt(user *models.User, promptType string) string {
	key := "choose_native_language"
	if promptType == "target" {
		key = "choose_target_language"
	}
	return s.localizer.Get(user.InterfaceLanguageCode, key)
}

// GetLocalizedLanguageName получает локализованное название языка (копия из оригинального сервиса)
func (s *OptimizedBotService) GetLocalizedLanguageName(langCode, interfaceLangCode string) string {
	return s.localizer.GetLanguageName(langCode, interfaceLangCode)
}

