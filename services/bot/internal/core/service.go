package core

import (
	"database/sql"
	"fmt"
	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
	"log"
	"strings"
	"time"
)

type BotService struct {
	DB                       database.Database
	Localizer                *localization.Localizer
	Cache                    cache.CacheServiceInterface
	InvalidationService      *cache.InvalidationService
	MetricsService           *cache.MetricsService
	BatchLoader              *database.BatchLoader
	FeedbackNotificationFunc func(data map[string]interface{}) error // функция для отправки уведомлений
}

func NewBotService(db *database.DB) *BotService {
	// Создаем кэш с конфигурацией по умолчанию
	cacheService := cache.NewCacheService(cache.DefaultCacheConfig())

	// Создаем сервисы для управления кэшем
	invalidationService := cache.NewInvalidationService(cacheService)
	metricsService := cache.NewMetricsService(cacheService)

	// Создаем BatchLoader для оптимизации N+1 запросов
	batchLoader := database.NewBatchLoader(db)

	return &BotService{
		DB:                  &databaseAdapter{db: db}, // Оборачиваем в адаптер
		Localizer:           localization.NewLocalizer(db.GetConnection()),
		Cache:               cacheService,
		InvalidationService: invalidationService,
		MetricsService:      metricsService,
		BatchLoader:         batchLoader,
	}
}

// NewBotServiceWithRedis создает BotService с Redis кэшем
func NewBotServiceWithRedis(db *database.DB, redisURL, redisPassword string, redisDB int) (*BotService, error) {
	// Создаем Redis кэш
	redisCache, err := cache.NewRedisCacheService(redisURL, redisPassword, redisDB, cache.DefaultCacheConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis cache: %w", err)
	}

	// Создаем сервисы для управления кэшем
	invalidationService := cache.NewInvalidationService(redisCache)
	metricsService := cache.NewMetricsService(redisCache)

	// Создаем BatchLoader для оптимизации N+1 запросов
	batchLoader := database.NewBatchLoader(db)

	return &BotService{
		DB:                  &databaseAdapter{db: db}, // Оборачиваем в адаптер
		Localizer:           localization.NewLocalizer(db.GetConnection()),
		Cache:               redisCache,
		InvalidationService: invalidationService,
		MetricsService:      metricsService,
		BatchLoader:         batchLoader,
	}, nil
}

// databaseAdapter адаптер для совместимости с интерфейсом Database
type databaseAdapter struct {
	db *database.DB
}

// Реализуем все методы интерфейса, делегируя к db или создавая заглушки

func (a *databaseAdapter) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	return a.db.FindOrCreateUser(telegramID, username, firstName)
}

func (a *databaseAdapter) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	return a.db.GetUserByTelegramID(telegramID)
}

func (a *databaseAdapter) UpdateUser(user *models.User) error {
	return a.db.UpdateUser(user)
}

func (a *databaseAdapter) UpdateUserInterfaceLanguage(userID int, language string) error {
	return a.db.UpdateUserInterfaceLanguage(userID, language)
}

func (a *databaseAdapter) UpdateUserState(userID int, state string) error {
	return a.db.UpdateUserState(userID, state)
}

func (a *databaseAdapter) UpdateUserStatus(userID int, status string) error {
	return a.db.UpdateUserStatus(userID, status)
}

func (a *databaseAdapter) UpdateUserNativeLanguage(userID int, langCode string) error {
	return a.db.UpdateUserNativeLanguage(userID, langCode)
}

func (a *databaseAdapter) UpdateUserTargetLanguage(userID int, langCode string) error {
	return a.db.UpdateUserTargetLanguage(userID, langCode)
}

func (a *databaseAdapter) UpdateUserTargetLanguageLevel(userID int, level string) error {
	return a.db.UpdateUserTargetLanguageLevel(userID, level)
}

func (a *databaseAdapter) ResetUserProfile(userID int) error {
	return a.db.ResetUserProfile(userID)
}

func (a *databaseAdapter) GetLanguages() ([]*models.Language, error) {
	return a.db.GetLanguages()
}

func (a *databaseAdapter) GetLanguageByCode(code string) (*models.Language, error) {
	return a.db.GetLanguageByCode(code)
}

func (a *databaseAdapter) GetInterests() ([]*models.Interest, error) {
	return a.db.GetInterests()
}

func (a *databaseAdapter) GetUserSelectedInterests(userID int) ([]int, error) {
	return a.db.GetUserSelectedInterests(userID)
}

func (a *databaseAdapter) SaveUserInterests(userID int64, interestIDs []int) error {
	return a.db.SaveUserInterests(userID, interestIDs)
}

func (a *databaseAdapter) SaveUserInterest(userID, interestID int, isPrimary bool) error {
	return a.db.SaveUserInterest(userID, interestID, isPrimary)
}

func (a *databaseAdapter) RemoveUserInterest(userID, interestID int) error {
	return a.db.RemoveUserInterest(userID, interestID)
}

func (a *databaseAdapter) ClearUserInterests(userID int) error {
	return a.db.ClearUserInterests(userID)
}

func (a *databaseAdapter) SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error {
	return a.db.SaveUserFeedback(userID, feedbackText, contactInfo)
}

func (a *databaseAdapter) GetUnprocessedFeedback() ([]map[string]interface{}, error) {
	return a.db.GetUnprocessedFeedback()
}

func (a *databaseAdapter) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	return a.db.MarkFeedbackProcessed(feedbackID, adminResponse)
}

func (a *databaseAdapter) GetConnection() *sql.DB {
	return a.db.GetConnection()
}

func (a *databaseAdapter) Close() error {
	return a.db.Close()
}

// NewBotServiceWithInterface создает BotService с интерфейсом Database (для тестов)
func NewBotServiceWithInterface(db database.Database, localizer *localization.Localizer) *BotService {
	return &BotService{
		DB:        db,
		Localizer: localizer,
	}
}

// SetFeedbackNotificationFunc устанавливает функцию для отправки уведомлений о новых отзывах
func (s *BotService) SetFeedbackNotificationFunc(fn func(map[string]interface{}) error) {
	s.FeedbackNotificationFunc = fn
}

func (s *BotService) DetectLanguage(telegramLangCode string) string {
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

func (s *BotService) HandleUserRegistration(telegramID int64, username, firstName, telegramLangCode string) (*models.User, error) {
	user, err := s.DB.FindOrCreateUser(telegramID, username, firstName)
	if err != nil {
		return nil, err
	}

	detected := s.DetectLanguage(telegramLangCode)
	// Определяем начальный язык интерфейса только для новых пользователей
	if user.Status == models.StatusNew || user.InterfaceLanguageCode == "" {
		// Для новых пользователей устанавливаем язык интерфейса по настройкам Telegram
		// Если язык не определен, используем русский как дефолт для проекта
		if detected == "" {
			user.InterfaceLanguageCode = "ru"
		} else {
			user.InterfaceLanguageCode = detected
		}
		_ = s.DB.UpdateUserInterfaceLanguage(user.ID, user.InterfaceLanguageCode)
	}
	return user, nil
}

func (s *BotService) GetWelcomeMessage(user *models.User) string {
	return s.Localizer.GetWithParams(user.InterfaceLanguageCode, "welcome_message", map[string]string{
		"name": user.FirstName,
	})
}

func (s *BotService) GetLanguagePrompt(user *models.User, promptType string) string {
	key := "choose_native_language"
	if promptType == "target" {
		key = "choose_target_language"
	}
	return s.Localizer.Get(user.InterfaceLanguageCode, key)
}

func (s *BotService) GetLocalizedLanguageName(langCode, interfaceLangCode string) string {
	return s.Localizer.GetLanguageName(langCode, interfaceLangCode)
}

func (s *BotService) GetLocalizedInterests(langCode string) (map[int]string, error) {
	return s.Localizer.GetInterests(langCode)
}

// IsProfileCompleted проверяет наличие языков и хотя бы одного интереса.
func (s *BotService) IsProfileCompleted(user *models.User) (bool, error) {
	if user.NativeLanguageCode == "" || user.TargetLanguageCode == "" {
		return false, nil
	}
	ids, err := s.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		return false, err
	}
	return len(ids) > 0, nil
}

// BuildProfileSummary возвращает локализованное резюме профиля.
func (s *BotService) BuildProfileSummary(user *models.User) (string, error) {
	lang := user.InterfaceLanguageCode
	nativeName := s.Localizer.GetLanguageName(user.NativeLanguageCode, lang)
	targetName := s.Localizer.GetLanguageName(user.TargetLanguageCode, lang)

	// Определяем флаги языков
	var nativeFlag, targetFlag string
	switch user.NativeLanguageCode {
	case "ru":
		nativeFlag = "🇷🇺"
	case "en":
		nativeFlag = "🇺🇸"
	case "es":
		nativeFlag = "🇪🇸"
	case "zh":
		nativeFlag = "🇨🇳"
	default:
		nativeFlag = "🌍"
	}

	switch user.TargetLanguageCode {
	case "ru":
		targetFlag = "🇷🇺"
	case "en":
		targetFlag = "🇺🇸"
	case "es":
		targetFlag = "🇪🇸"
	case "zh":
		targetFlag = "🇨🇳"
	default:
		targetFlag = "🌍"
	}

	ids, err := s.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		ids = []int{}
	}
	allInterests, _ := s.Localizer.GetInterests(lang)

	var picked []string
	for _, id := range ids {
		if name, ok := allInterests[id]; ok {
			picked = append(picked, name)
		}
	}
	interestsLine := fmt.Sprintf("🎯 %s: %d", s.Localizer.Get(lang, "profile_field_interests"), len(picked))
	if len(picked) > 0 {
		interestsLine = fmt.Sprintf("🎯 %s: %d\n• %s", s.Localizer.Get(lang, "profile_field_interests"), len(picked), strings.Join(picked, ", "))
	}

	title := s.Localizer.Get(lang, "profile_summary_title")
	native := fmt.Sprintf("%s %s: %s", nativeFlag, s.Localizer.Get(lang, "profile_field_native"), nativeName)
	target := fmt.Sprintf("%s %s: %s", targetFlag, s.Localizer.Get(lang, "profile_field_target"), targetName)

	return fmt.Sprintf("%s\n\n%s\n%s\n%s", title, native, target, interestsLine), nil
}

// Методы работы с обратной связью

// SendFeedbackNotification отправляет уведомление администраторам о новом отзыве
func (s *BotService) SendFeedbackNotification(feedbackData map[string]interface{}, admins []int64) error {
	if s.FeedbackNotificationFunc != nil {
		return s.FeedbackNotificationFunc(feedbackData)
	}

	// Fallback: логируем уведомление если функция не установлена
	adminMsg := fmt.Sprintf(`
📝 Новый отзыв от пользователя:

👤 Имя: %s
�� Telegram ID: %d

%s

📝 Отзыв:
%s
`,
		feedbackData["first_name"].(string),
		feedbackData["telegram_id"].(int64),
		func() string {
			if username, ok := feedbackData["username"].(*string); ok && username != nil {
				return fmt.Sprintf("👤 Username: @%s", *username)
			}
			return "👤 Username: отсутствует"
		}(),
		feedbackData["feedback_text"].(string),
	)

	// Добавляем контактную информацию, если есть
	if contactInfo, ok := feedbackData["contact_info"].(*string); ok && contactInfo != nil {
		adminMsg += fmt.Sprintf("\n📞 Контакты: %s", *contactInfo)
	}

	// Пока что просто логируем уведомление
	log.Printf("Отправка уведомления администраторам: %s", adminMsg)

	return nil
}

// ValidateFeedback проверяет корректность отзыва по длине
func (s *BotService) ValidateFeedback(feedbackText string) error {
	length := len([]rune(feedbackText)) // Учитываем Unicode
	if length < 10 {
		return fmt.Errorf("feedback too short: %d characters, minimum 10", length)
	}
	if length > 1000 {
		return fmt.Errorf("feedback too long: %d characters, maximum 1000", length)
	}
	return nil
}

// SaveUserFeedback сохраняет отзыв пользователя и отправляет уведомления
func (s *BotService) SaveUserFeedback(userID int, feedbackText string, contactInfo *string, admins []int64) error {
	// Валидируем отзыв
	if err := s.ValidateFeedback(feedbackText); err != nil {
		return err
	}

	// Сохраняем в базу данных
	if err := s.DB.SaveUserFeedback(userID, feedbackText, contactInfo); err != nil {
		return fmt.Errorf("ошибка сохранения отзыва в базу данных: %w", err)
	}

	// Получаем данные пользователя для уведомления администраторов
	userData, err := s.GetUserDataForFeedback(userID)
	if err != nil {
		log.Printf("Не удалось получить данные пользователя для уведомления: %v", err)
		return nil // Возвращаемся без ошибки
	}

	// Объединяем данные с отзывом
	fbData := userData
	fbData["feedback_text"] = feedbackText
	if contactInfo != nil {
		fbData["contact_info"] = contactInfo
	}

	// Отправляем уведомление администраторам
	if err := s.SendFeedbackNotification(fbData, admins); err != nil {
		log.Printf("Ошибка отправки уведомления администраторам: %v", err)
		// Пытаемся отправить уведомление через function callback если он установлен
		if s.FeedbackNotificationFunc != nil {
			if err := s.FeedbackNotificationFunc(fbData); err != nil {
				log.Printf("Fallback: также не удалось отправить через callback: %v", err)
			} else {
				log.Printf("Fallback: уведомление отправлено через callback")
			}
		}
	}

	return nil
}

// GetUserDataForFeedback получает данные пользователя для формирования уведомления о новом отзыве
func (s *BotService) GetUserDataForFeedback(userID int) (map[string]interface{}, error) {
	// Получаем пользователя по ID (нужно добавить метод в DB)
	var telegramID int64
	var username, firstName string
	err := s.DB.GetConnection().QueryRow(`
		SELECT telegram_id, username, first_name
		FROM users WHERE id = $1
	`, userID).Scan(&telegramID, &username, &firstName)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"telegram_id": telegramID,
		"first_name":  firstName,
	}

	if username != "" {
		result["username"] = &username
	}

	return result, nil
}

// GetAllUnprocessedFeedback получает все необработанные отзывы для администратора
func (s *BotService) GetAllUnprocessedFeedback() ([]map[string]interface{}, error) {
	return s.DB.GetUnprocessedFeedback()
}

// GetAllFeedback получает все отзывы для администратора
func (s *BotService) GetAllFeedback() ([]map[string]interface{}, error) {
	query := `
        SELECT uf.id, uf.feedback_text, uf.contact_info, uf.created_at,
               uf.is_processed, u.username, u.telegram_id, u.first_name,
               uf.admin_response
        FROM user_feedback uf
        JOIN users u ON uf.user_id = u.id
        ORDER BY uf.created_at DESC
    `

	rows, err := s.DB.GetConnection().Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []map[string]interface{}
	for rows.Next() {
		var (
			id           int
			feedbackText string
			contactInfo  sql.NullString
			createdAt    sql.NullTime
			isProcessed  bool
			username     sql.NullString
			telegramID   int64
			firstName    string
			adminResp    sql.NullString
		)

		err := rows.Scan(&id, &feedbackText, &contactInfo, &createdAt, &isProcessed,
			&username, &telegramID, &firstName, &adminResp)
		if err != nil {
			continue // Пропускаем ошибочные записи
		}

		feedback := map[string]interface{}{
			"id":            id,
			"feedback_text": feedbackText,
			"created_at":    createdAt.Time,
			"telegram_id":   telegramID,
			"first_name":    firstName,
			"is_processed":  isProcessed,
		}

		if username.Valid {
			feedback["username"] = username.String
		} else {
			feedback["username"] = nil
		}

		if contactInfo.Valid {
			feedback["contact_info"] = contactInfo.String
		} else {
			feedback["contact_info"] = nil
		}

		if adminResp.Valid {
			feedback["admin_response"] = adminResp.String
		} else {
			feedback["admin_response"] = nil
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// UpdateFeedbackStatus обновляет статус отзыва (обработан/не обработан)
func (s *BotService) UpdateFeedbackStatus(feedbackID int, isProcessed bool) error {
	query := `
		UPDATE user_feedback
		SET is_processed = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.DB.GetConnection().Exec(query, isProcessed, feedbackID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса отзыва: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить количество измененных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("отзыв с ID %d не найден", feedbackID)
	}

	return nil
}

// ArchiveFeedback архивирует отзыв
func (s *BotService) ArchiveFeedback(feedbackID int) error {
	query := `
		UPDATE user_feedback
		SET is_processed = true, updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.DB.GetConnection().Exec(query, feedbackID)
	if err != nil {
		return fmt.Errorf("ошибка архивирования отзыва: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("отзыв с ID %d не найден", feedbackID)
	}

	return nil
}

// DeleteFeedback удаляет отзыв из базы данных
func (s *BotService) DeleteFeedback(feedbackID int) error {
	query := `DELETE FROM user_feedback WHERE id = $1`

	result, err := s.DB.GetConnection().Exec(query, feedbackID)
	if err != nil {
		return fmt.Errorf("ошибка удаления отзыва: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить количество удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("отзыв с ID %d не найден", feedbackID)
	}

	return nil
}

// MarkFeedbackProcessed помечает отзыв как обработанный с ответом
func (s *BotService) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	return s.DB.MarkFeedbackProcessed(feedbackID, adminResponse)
}

// DeleteAllProcessedFeedbacks удаляет все обработанные отзывы
func (s *BotService) DeleteAllProcessedFeedbacks() (int, error) {
	query := `DELETE FROM user_feedback WHERE is_processed = true`
	result, err := s.DB.GetConnection().Exec(query)
	if err != nil {
		return 0, fmt.Errorf("ошибка удаления обработанных отзывов: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества удаленных строк: %w", err)
	}

	return int(rowsAffected), nil
}

// UnarchiveFeedback возвращает отзыв в активные (убирает флаг is_processed)
func (s *BotService) UnarchiveFeedback(feedbackID int) error {
	query := `
		UPDATE user_feedback
		SET is_processed = false, updated_at = NOW()
		WHERE id = $1
	`
	result, err := s.DB.GetConnection().Exec(query, feedbackID)
	if err != nil {
		return fmt.Errorf("ошибка возврата отзыва в активные: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("отзыв с ID %d не найден", feedbackID)
	}

	return nil
}

// ===== КЭШИРОВАННЫЕ МЕТОДЫ =====

// GetCachedLanguages получает языки из кэша или загружает из БД
func (s *BotService) GetCachedLanguages(lang string) ([]*models.Language, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if languages, found := s.Cache.GetLanguages(lang); found {
		return languages, nil
	}

	// Загружаем из БД
	languages, err := s.DB.GetLanguages()
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	// Сохраняем в кэш
	s.Cache.SetLanguages(lang, languages)

	return languages, nil
}

// GetCachedInterests получает интересы из кэша или загружает из БД
func (s *BotService) GetCachedInterests(lang string) (map[int]string, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if interests, found := s.Cache.GetInterests(lang); found {
		return interests, nil
	}

	// Загружаем из БД и локализуем
	interests, err := s.Localizer.GetInterests(lang)
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	// Сохраняем в кэш
	s.Cache.SetInterests(lang, interests)

	return interests, nil
}

// GetCachedUser получает пользователя из кэша или загружает из БД
func (s *BotService) GetCachedUser(telegramID int64) (*models.User, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if user, found := s.Cache.GetUser(telegramID); found {
		return user, nil
	}

	// Загружаем из БД
	user, err := s.DB.GetUserByTelegramID(telegramID)
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	// Сохраняем в кэш
	s.Cache.SetUser(user)

	return user, nil
}

// GetCachedTranslations получает переводы из кэша или загружает из файлов
func (s *BotService) GetCachedTranslations(lang string) (map[string]string, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if translations, found := s.Cache.GetTranslations(lang); found {
		return translations, nil
	}

	// Загружаем из файлов локализации
	// Здесь нужно будет добавить метод в Localizer для получения всех переводов
	// Пока что возвращаем пустую карту
	translations := make(map[string]string)

	// Сохраняем в кэш
	s.Cache.SetTranslations(lang, translations)

	return translations, nil
}

// UpdateCachedUser обновляет пользователя в БД и кэше
func (s *BotService) UpdateCachedUser(user *models.User) error {
	// Обновляем в БД
	if err := s.DB.UpdateUser(user); err != nil {
		return err
	}

	// Обновляем в кэше
	s.Cache.SetUser(user)

	return nil
}

// InvalidateUserCache инвалидирует кэш пользователя
func (s *BotService) InvalidateUserCache(userID int64) {
	s.InvalidationService.InvalidateUserData(userID)
}

// InvalidateStaticDataCache инвалидирует кэш статических данных
func (s *BotService) InvalidateStaticDataCache() {
	s.InvalidationService.InvalidateStaticData()
}

// GetCacheStats возвращает статистику кэша
func (s *BotService) GetCacheStats() map[string]interface{} {
	return s.MetricsService.GetMetrics()
}

// StopCache останавливает кэш-сервис
func (s *BotService) StopCache() {
	s.Cache.Stop()
}

// ===== BATCH LOADING МЕТОДЫ =====

// GetUserWithAllData получает пользователя со всеми связанными данными одним запросом
func (s *BotService) GetUserWithAllData(telegramID int64) (*database.UserWithAllData, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if userData, found := s.Cache.GetUser(telegramID); found {
		// Если пользователь есть в кэше, но нет полных данных, загружаем их
		if userData != nil {
			// Загружаем полные данные
			return s.BatchLoader.GetUserWithAllData(telegramID)
		}
	}

	// Загружаем из БД одним запросом
	userData, err := s.BatchLoader.GetUserWithAllData(telegramID)
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	// Сохраняем в кэш
	s.Cache.SetUser(userData.User)

	return userData, nil
}

// BatchLoadUsersWithInterests загружает нескольких пользователей с их интересами одним запросом
func (s *BotService) BatchLoadUsersWithInterests(telegramIDs []int64) (map[int64]*database.UserWithInterests, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	users, err := s.BatchLoader.BatchLoadUsersWithInterests(telegramIDs)
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	// Сохраняем пользователей в кэш
	for _, userData := range users {
		s.Cache.SetUser(userData.User)
	}

	return users, nil
}

// BatchLoadInterestsWithTranslations загружает интересы с переводами для нескольких языков
func (s *BotService) BatchLoadInterestsWithTranslations(languages []string) (map[string]map[int]string, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	interests, err := s.BatchLoader.BatchLoadInterestsWithTranslations(languages)
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	// Сохраняем в кэш
	for lang, langInterests := range interests {
		s.Cache.SetInterests(lang, langInterests)
	}

	return interests, nil
}

// BatchLoadLanguagesWithTranslations загружает языки с переводами для нескольких языков
func (s *BotService) BatchLoadLanguagesWithTranslations(languages []string) (map[string][]*models.Language, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	langs, err := s.BatchLoader.BatchLoadLanguagesWithTranslations(languages)
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	// Сохраняем в кэш
	for lang, langList := range langs {
		s.Cache.SetLanguages(lang, langList)
	}

	return langs, nil
}

// BatchLoadUserInterests загружает интересы для нескольких пользователей одним запросом
func (s *BotService) BatchLoadUserInterests(userIDs []int) (map[int][]int, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	interests, err := s.BatchLoader.BatchLoadUserInterests(userIDs)
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	return interests, nil
}

// BatchLoadUsers загружает пользователей по Telegram ID одним запросом
func (s *BotService) BatchLoadUsers(telegramIDs []int64) (map[int64]*models.User, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	users, err := s.BatchLoader.BatchLoadUsers(telegramIDs)
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	// Сохраняем в кэш
	for _, user := range users {
		s.Cache.SetUser(user)
	}

	return users, nil
}

// BatchLoadStats загружает статистику для нескольких типов одним запросом
func (s *BotService) BatchLoadStats(statTypes []string) (map[string]map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	stats, err := s.BatchLoader.BatchLoadStats(statTypes)
	if err != nil {
		s.MetricsService.RecordError()
		return nil, err
	}

	return stats, nil
}
