// Package core provides the main business logic for the language exchange bot.
package core

import (
	"context"
	"database/sql"
	"fmt"
	"language-exchange-bot/internal/cache"
	"language-exchange-bot/internal/database"
	errorsPkg "language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/logging"
	"language-exchange-bot/internal/models"
	"language-exchange-bot/internal/validation"
	"log"
	"strings"
	"time"
)

// Константы для валидации.
const (
	// minFeedbackLength - минимальная длина отзыва.
	minFeedbackLength = 10

	// maxFeedbackLength - максимальная длина отзыва.
	maxFeedbackLength = 1000
)

// BotService provides the main business logic for the language exchange bot.
type BotService struct {
	DB                       database.Database
	Localizer                *localization.Localizer
	Cache                    cache.ServiceInterface
	InvalidationService      *cache.InvalidationService
	MetricsService           *cache.MetricsService
	BatchLoader              *database.BatchLoader
	Service                  *validation.Service
	LoggingService           *logging.LoggingService
	FeedbackNotificationFunc func(data map[string]interface{}) error // функция для отправки уведомлений
}

// NewBotService creates a new BotService instance.
func NewBotService(db *database.DB, errorHandler interface{}) *BotService {
	// Создаем кэш с конфигурацией по умолчанию
	cacheService := cache.NewService(cache.DefaultConfig())

	// Создаем сервисы для управления кэшем
	invalidationService := cache.NewInvalidationService(cacheService)
	metricsService := cache.NewMetricsService(cacheService)

	// Создаем BatchLoader для оптимизации N+1 запросов
	batchLoader := database.NewBatchLoader(db)

	// Создаем Service (пока без errorHandler для совместимости)
	var validationService *validation.Service

	var loggingService *logging.LoggingService

	if errorHandler != nil {
		if handler, ok := errorHandler.(*errorsPkg.ErrorHandler); ok {
			validationService = validation.NewService(handler)
			loggingService = logging.NewLoggingService(handler)
		}
	}

	return &BotService{
		DB:                       &databaseAdapter{db: db}, // Оборачиваем в адаптер
		Localizer:                localization.NewLocalizer(db.GetConnection()),
		Cache:                    cacheService,
		InvalidationService:      invalidationService,
		MetricsService:           metricsService,
		BatchLoader:              batchLoader,
		Service:                  validationService,
		LoggingService:           loggingService,
		FeedbackNotificationFunc: nil,
	}
}

// NewBotServiceWithRedis создает BotService с Redis кэшем.
func NewBotServiceWithRedis(
	db *database.DB,
	redisURL, redisPassword string,
	redisDB int,
	errorHandler interface{},
) (*BotService, error) {
	// Создаем Redis кэш
	redisCache, err := cache.NewRedisCacheService(redisURL, redisPassword, redisDB, cache.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis cache: %w", err)
	}

	// Создаем сервисы для управления кэшем
	invalidationService := cache.NewInvalidationService(redisCache)
	metricsService := cache.NewMetricsService(redisCache)

	// Создаем BatchLoader для оптимизации N+1 запросов
	batchLoader := database.NewBatchLoader(db)

	// Создаем Service и LoggingService
	var validationService *validation.Service

	var loggingService *logging.LoggingService

	if errorHandler != nil {
		if handler, ok := errorHandler.(*errorsPkg.ErrorHandler); ok {
			validationService = validation.NewService(handler)
			loggingService = logging.NewLoggingService(handler)
		}
	}

	return &BotService{
		DB:                       &databaseAdapter{db: db}, // Оборачиваем в адаптер
		Localizer:                localization.NewLocalizer(db.GetConnection()),
		Cache:                    redisCache,
		InvalidationService:      invalidationService,
		MetricsService:           metricsService,
		BatchLoader:              batchLoader,
		Service:                  validationService,
		LoggingService:           loggingService,
		FeedbackNotificationFunc: nil,
	}, nil
}

// databaseAdapter адаптер для совместимости с интерфейсом Database.
type databaseAdapter struct {
	db *database.DB
}

// Реализуем все методы интерфейса, делегируя к db или создавая заглушки

func (a *databaseAdapter) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	user, err := a.db.FindOrCreateUser(telegramID, username, firstName)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	return user, nil
}

func (a *databaseAdapter) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	user, err := a.db.GetUserByTelegramID(telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}

	return user, nil
}

func (a *databaseAdapter) UpdateUser(user *models.User) error {
	err := a.db.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (a *databaseAdapter) UpdateUserInterfaceLanguage(userID int, language string) error {
	err := a.db.UpdateUserInterfaceLanguage(userID, language)
	if err != nil {
		return fmt.Errorf("failed to update user interface language: %w", err)
	}

	return nil
}

func (a *databaseAdapter) UpdateUserState(userID int, state string) error {
	err := a.db.UpdateUserState(userID, state)
	if err != nil {
		return fmt.Errorf("failed to update user state: %w", err)
	}

	return nil
}

func (a *databaseAdapter) UpdateUserStatus(userID int, status string) error {
	err := a.db.UpdateUserStatus(userID, status)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

func (a *databaseAdapter) UpdateUserNativeLanguage(userID int, langCode string) error {
	err := a.db.UpdateUserNativeLanguage(userID, langCode)
	if err != nil {
		return fmt.Errorf("failed to update user native language: %w", err)
	}

	return nil
}

func (a *databaseAdapter) UpdateUserTargetLanguage(userID int, langCode string) error {
	err := a.db.UpdateUserTargetLanguage(userID, langCode)
	if err != nil {
		return fmt.Errorf("failed to update user target language: %w", err)
	}

	return nil
}

func (a *databaseAdapter) UpdateUserTargetLanguageLevel(userID int, level string) error {
	err := a.db.UpdateUserTargetLanguageLevel(userID, level)
	if err != nil {
		return fmt.Errorf("failed to update user target language level: %w", err)
	}

	return nil
}

func (a *databaseAdapter) ResetUserProfile(userID int) error {
	err := a.db.ResetUserProfile(userID)
	if err != nil {
		return fmt.Errorf("failed to reset user profile: %w", err)
	}

	return nil
}

func (a *databaseAdapter) GetLanguages() ([]*models.Language, error) {
	languages, err := a.db.GetLanguages()
	if err != nil {
		return nil, fmt.Errorf("failed to get languages: %w", err)
	}

	return languages, nil
}

func (a *databaseAdapter) GetLanguageByCode(code string) (*models.Language, error) {
	language, err := a.db.GetLanguageByCode(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get language by code: %w", err)
	}

	return language, nil
}

func (a *databaseAdapter) GetInterests() ([]*models.Interest, error) {
	interests, err := a.db.GetInterests()
	if err != nil {
		return nil, fmt.Errorf("failed to get interests: %w", err)
	}

	return interests, nil
}

func (a *databaseAdapter) GetUserSelectedInterests(userID int) ([]int, error) {
	interests, err := a.db.GetUserSelectedInterests(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user selected interests: %w", err)
	}

	return interests, nil
}

func (a *databaseAdapter) SaveUserInterests(userID int64, interestIDs []int) error {
	err := a.db.SaveUserInterests(userID, interestIDs)
	if err != nil {
		return fmt.Errorf("failed to save user interests: %w", err)
	}

	return nil
}

func (a *databaseAdapter) SaveUserInterest(userID, interestID int, isPrimary bool) error {
	err := a.db.SaveUserInterest(userID, interestID, isPrimary)
	if err != nil {
		return fmt.Errorf("failed to save user interest: %w", err)
	}

	return nil
}

func (a *databaseAdapter) RemoveUserInterest(userID, interestID int) error {
	err := a.db.RemoveUserInterest(userID, interestID)
	if err != nil {
		return fmt.Errorf("failed to remove user interest: %w", err)
	}

	return nil
}

func (a *databaseAdapter) ClearUserInterests(userID int) error {
	err := a.db.ClearUserInterests(userID)
	if err != nil {
		return fmt.Errorf("failed to clear user interests: %w", err)
	}

	return nil
}

func (a *databaseAdapter) SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error {
	err := a.db.SaveUserFeedback(userID, feedbackText, contactInfo)
	if err != nil {
		return fmt.Errorf("failed to save user feedback: %w", err)
	}

	return nil
}

func (a *databaseAdapter) GetUnprocessedFeedback() ([]map[string]interface{}, error) {
	feedback, err := a.db.GetUnprocessedFeedback()
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed feedback: %w", err)
	}

	return feedback, nil
}

func (a *databaseAdapter) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	err := a.db.MarkFeedbackProcessed(feedbackID, adminResponse)
	if err != nil {
		return fmt.Errorf("failed to mark feedback processed: %w", err)
	}

	return nil
}

func (a *databaseAdapter) GetConnection() *sql.DB {
	return a.db.GetConnection()
}

func (a *databaseAdapter) Close() error {
	err := a.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

// NewBotServiceWithInterface создает BotService с интерфейсом Database (для тестов).
func NewBotServiceWithInterface(db database.Database, localizer *localization.Localizer) *BotService {
	return &BotService{
		DB:                       db,
		Localizer:                localizer,
		Cache:                    nil,
		InvalidationService:      nil,
		MetricsService:           nil,
		BatchLoader:              nil,
		Service:                  nil,
		LoggingService:           nil,
		FeedbackNotificationFunc: nil,
	}
}

// SetFeedbackNotificationFunc устанавливает функцию для отправки уведомлений о новых отзывах.
func (s *BotService) SetFeedbackNotificationFunc(fn func(map[string]interface{}) error) {
	s.FeedbackNotificationFunc = fn
}

// DetectLanguage определяет язык интерфейса по коду языка Telegram.
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

// HandleUserRegistration обрабатывает регистрацию нового пользователя.
func (s *BotService) HandleUserRegistration(
	telegramID int64,
	username, firstName, telegramLangCode string,
) (*models.User, error) {
	user, err := s.DB.FindOrCreateUser(telegramID, username, firstName)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
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

// GetWelcomeMessage возвращает приветственное сообщение для пользователя.
func (s *BotService) GetWelcomeMessage(user *models.User) string {
	return s.Localizer.GetWithParams(user.InterfaceLanguageCode, "welcome_message", map[string]string{
		"name": user.FirstName,
	})
}

// GetLanguagePrompt возвращает подсказку для выбора языка.
func (s *BotService) GetLanguagePrompt(user *models.User, promptType string) string {
	key := "choose_native_language"

	if promptType == "target" {
		key = "choose_target_language"
	}

	return s.Localizer.Get(user.InterfaceLanguageCode, key)
}

// GetLocalizedLanguageName возвращает локализованное название языка.
func (s *BotService) GetLocalizedLanguageName(langCode, interfaceLangCode string) string {
	return s.Localizer.GetLanguageName(langCode, interfaceLangCode)
}

// GetLocalizedInterests возвращает локализованные интересы для указанного языка.
func (s *BotService) GetLocalizedInterests(langCode string) (map[int]string, error) {
	interests, err := s.Localizer.GetInterests(langCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get interests: %w", err)
	}

	return interests, nil
}

// IsProfileCompleted проверяет наличие языков и хотя бы одного интереса.
func (s *BotService) IsProfileCompleted(user *models.User) (bool, error) {
	if user.NativeLanguageCode == "" || user.TargetLanguageCode == "" {
		return false, nil
	}

	ids, err := s.DB.GetUserSelectedInterests(user.ID)
	if err != nil {
		return false, fmt.Errorf("operation failed: %w", err)
	}

	return len(ids) > 0, nil
}

// BuildProfileSummary возвращает локализованное резюме профиля.
func (s *BotService) BuildProfileSummary(user *models.User) (string, error) {
	lang := user.InterfaceLanguageCode

	// Получаем основную информацию
	basicInfo := s.buildBasicProfileInfo(user, lang)
	languageInfo := s.buildLanguageProfileInfo(user, lang)
	interestsInfo := s.buildInterestsProfileInfo(user, lang)
	additionalInfo := s.buildAdditionalProfileInfo(user, lang)

	// Объединяем все части
	lines := []string{basicInfo}
	lines = append(lines, "", languageInfo, interestsInfo)
	lines = append(lines, "")
	lines = append(lines, additionalInfo...)

	return strings.Join(lines, "\n"), nil
}

// buildBasicProfileInfo строит основную информацию профиля.
func (s *BotService) buildBasicProfileInfo(user *models.User, lang string) string {
	displayName := s.getDisplayName(user)
	nameLine := fmt.Sprintf("👤 %s: %s", s.Localizer.Get(lang, "profile_field_name"), displayName)

	usernameLine := ""
	if user.Username != "" {
		usernameLine = fmt.Sprintf("🔗 %s: @%s", s.Localizer.Get(lang, "profile_field_username"), user.Username)
	}

	if usernameLine != "" {
		return nameLine + "\n" + usernameLine
	}

	return nameLine
}

// buildLanguageProfileInfo строит информацию о языках.
func (s *BotService) buildLanguageProfileInfo(user *models.User, lang string) string {
	nativeName := s.Localizer.GetLanguageName(user.NativeLanguageCode, lang)
	targetName := s.Localizer.GetLanguageName(user.TargetLanguageCode, lang)

	nativeFlag := s.getLanguageFlag(user.NativeLanguageCode)
	targetFlag := s.getLanguageFlag(user.TargetLanguageCode)

	native := fmt.Sprintf("%s %s: %s", nativeFlag, s.Localizer.Get(lang, "profile_field_native"), nativeName)

	levelText := s.formatLanguageLevel(user.TargetLanguageLevel)
	target := fmt.Sprintf("%s %s: %s (%s)",
		targetFlag,
		s.Localizer.Get(lang, "profile_field_target"),
		targetName,
		levelText,
	)

	return native + "\n" + target
}

// buildInterestsProfileInfo строит информацию об интересах.
func (s *BotService) buildInterestsProfileInfo(user *models.User, lang string) string {
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
		interestsLine = fmt.Sprintf("🎯 %s: %d\n• %s",
			s.Localizer.Get(lang, "profile_field_interests"),
			len(picked),
			strings.Join(picked, ", "),
		)
	}

	return interestsLine
}

// buildAdditionalProfileInfo строит дополнительную информацию профиля.
func (s *BotService) buildAdditionalProfileInfo(user *models.User, lang string) []string {
	var lines []string

	// Временная доступность
	availabilityText := s.formatTimeAvailability(user.TimeAvailability, lang)
	lines = append(lines, fmt.Sprintf("⏰ %s: %s", s.Localizer.Get(lang, "profile_field_availability"), availabilityText))

	// Предпочтения общения
	communicationText := s.formatCommunicationPreferences(user.FriendshipPreferences, lang)
	lines = append(lines, fmt.Sprintf("💬 %s: %s", s.Localizer.Get(lang, "profile_field_communication"), communicationText))

	// Статус и время в системе
	statusText := s.formatUserStatus(user, lang)
	memberSinceText := s.formatMemberSince(user.CreatedAt, lang)
	lines = append(lines, "", statusText, memberSinceText)

	return lines
}

// formatTimeAvailability форматирует временную доступность.
func (s *BotService) formatTimeAvailability(ta *models.TimeAvailability, lang string) string {
	if ta == nil {
		return "Не указано"
	}

	dayText := s.formatDayType(ta, lang)
	timeText := s.formatTimeSlot(ta.TimeSlot, lang)

	return fmt.Sprintf("%s, %s", dayText, timeText)
}

// formatDayType форматирует тип дня.
func (s *BotService) formatDayType(ta *models.TimeAvailability, lang string) string {
	switch ta.DayType {
	case "weekdays":
		return s.Localizer.Get(lang, "time_weekdays")
	case "weekends":
		return s.Localizer.Get(lang, "time_weekends")
	case "any":
		return s.Localizer.Get(lang, "time_any")
	case "specific":
		return s.formatSpecificDays(ta.SpecificDays, lang)
	default:
		return s.Localizer.Get(lang, "time_any")
	}
}

// formatSpecificDays форматирует конкретные дни.
func (s *BotService) formatSpecificDays(specificDays []string, lang string) string {
	if len(specificDays) > 0 {
		return strings.Join(specificDays, ", ")
	}

	return s.Localizer.Get(lang, "time_any")
}

// formatTimeSlot форматирует временной слот.
func (s *BotService) formatTimeSlot(timeSlot, lang string) string {
	switch timeSlot {
	case "morning":
		return s.Localizer.Get(lang, "time_morning")
	case "day":
		return s.Localizer.Get(lang, "time_day")
	case "evening":
		return s.Localizer.Get(lang, "time_evening")
	case "late":
		return s.Localizer.Get(lang, "time_late")
	default:
		return s.Localizer.Get(lang, "time_any")
	}
}

// formatCommunicationPreferences форматирует предпочтения общения.
func (s *BotService) formatCommunicationPreferences(fp *models.FriendshipPreferences, lang string) string {
	if fp == nil {
		return "Не указано"
	}

	styleText := s.formatCommunicationStyle(fp.CommunicationStyle, lang)
	freqText := s.formatCommunicationFreq(fp.CommunicationFreq, lang)

	return fmt.Sprintf("%s, %s", styleText, freqText)
}

// formatCommunicationStyle форматирует стиль общения.
func (s *BotService) formatCommunicationStyle(style, lang string) string {
	switch style {
	case "text":
		return s.Localizer.Get(lang, "comm_text")
	case "voice_msg":
		return s.Localizer.Get(lang, "comm_voice")
	case "audio_call":
		return s.Localizer.Get(lang, "comm_audio")
	case "video_call":
		return s.Localizer.Get(lang, "comm_video")
	case "meet_person":
		return s.Localizer.Get(lang, "comm_meet")
	default:
		return style
	}
}

// formatCommunicationFreq форматирует частоту общения.
func (s *BotService) formatCommunicationFreq(freq, lang string) string {
	switch freq {
	case "spontaneous":
		return s.Localizer.Get(lang, "freq_spontaneous")
	case "weekly":
		return s.Localizer.Get(lang, "freq_weekly")
	case "daily":
		return s.Localizer.Get(lang, "freq_daily")
	case "intensive":
		return s.Localizer.Get(lang, "freq_intensive")
	default:
		return freq
	}
}

// formatUserStatus форматирует статус пользователя.
func (s *BotService) formatUserStatus(user *models.User, lang string) string {
	var statusText string

	var statusEmoji string

	switch user.Status {
	case "new":
		statusText = s.Localizer.Get(lang, "status_new")
		statusEmoji = "🆕"
	case "filling_profile":
		statusText = s.Localizer.Get(lang, "status_filling")
		statusEmoji = "📝"
	case "active":
		statusText = s.Localizer.Get(lang, "status_active")
		statusEmoji = "🟢"
	case "paused":
		statusText = s.Localizer.Get(lang, "status_paused")
		statusEmoji = "⏸️"
	default:
		statusText = user.Status
		statusEmoji = "❓"
	}

	return fmt.Sprintf("%s %s: %s", statusEmoji, s.Localizer.Get(lang, "profile_field_status"), statusText)
}

// formatMemberSince форматирует дату регистрации.
func (s *BotService) formatMemberSince(createdAt time.Time, lang string) string {
	dateStr := createdAt.Format("02.01.2006")

	return fmt.Sprintf("📅 %s: %s", s.Localizer.Get(lang, "profile_field_member_since"), dateStr)
}

// getDisplayName возвращает отображаемое имя пользователя.
func (s *BotService) getDisplayName(user *models.User) string {
	if user.Username == "madam_di_5" {
		return "Лисёнок 🦊"
	}

	return user.FirstName
}

// formatLanguageLevel форматирует уровень языка в читаемый вид.
func (s *BotService) formatLanguageLevel(level string) string {
	switch level {
	case "beginner":
		return "A1-A2"
	case "elementary":
		return "A2-B1"
	case "intermediate":
		return "B1-B2"
	case "upper_intermediate":
		return "B2-C1"
	case "advanced":
		return "C1-C2"
	default:
		return level
	}
}

// Методы работы с обратной связью

// SendFeedbackNotification отправляет уведомление администраторам о новом отзыве.
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
				return "👤 Username: @" + *username
			}

			return "👤 Username: отсутствует"
		}(),
		feedbackData["feedback_text"].(string),
	)

	// Добавляем контактную информацию, если есть
	if contactInfo, ok := feedbackData["contact_info"].(*string); ok && contactInfo != nil {
		adminMsg += "\n📞 Контакты: " + *contactInfo
	}

	// Пока что просто логируем уведомление
	log.Printf("Отправка уведомления администраторам: %s, to %v", adminMsg, admins)

	return nil
}

// ValidateFeedback проверяет корректность отзыва по длине.
func (s *BotService) ValidateFeedback(feedbackText string) error {
	length := len([]rune(feedbackText)) // Учитываем Unicode

	if length < minFeedbackLength {
		return errorsPkg.ErrFeedbackTooShort
	}

	if length > maxFeedbackLength {
		return errorsPkg.ErrFeedbackTooLong
	}

	return nil
}

// SaveUserFeedback сохраняет отзыв пользователя и отправляет уведомления.
func (s *BotService) SaveUserFeedback(userID int, feedbackText string, contactInfo *string, admins []int64) error {
	// Валидируем отзыв
	if err := s.ValidateFeedback(feedbackText); err != nil {
		return fmt.Errorf("operation failed: %w", err)
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
			err := s.FeedbackNotificationFunc(fbData)
			if err != nil {
				log.Printf("Fallback: также не удалось отправить через callback: %v", err)
			} else {
				log.Printf("Fallback: уведомление отправлено через callback")
			}
		}
	}

	return nil
}

// GetUserDataForFeedback получает данные пользователя для формирования уведомления о новом отзыве.
func (s *BotService) GetUserDataForFeedback(userID int) (map[string]interface{}, error) {
	// Получаем пользователя по ID (нужно добавить метод в DB)
	var telegramID int64

	var username, firstName string

	err := s.DB.GetConnection().QueryRowContext(context.Background(), `
		SELECT telegram_id, username, first_name
		FROM users WHERE id = $1
	`, userID).Scan(&telegramID, &username, &firstName)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
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

// GetAllUnprocessedFeedback получает все необработанные отзывы для администратора.
func (s *BotService) GetAllUnprocessedFeedback() ([]map[string]interface{}, error) {
	feedback, err := s.DB.GetUnprocessedFeedback()
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed feedback: %w", err)
	}

	return feedback, nil
}

// GetAllFeedback получает все отзывы для администратора.
func (s *BotService) GetAllFeedback() ([]map[string]interface{}, error) {
	query := getFeedbackQuery()

	rows, err := s.DB.GetConnection().QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// В defer мы не можем вернуть ошибку, но можем логировать
			// TODO: интегрировать с системой логирования
			_ = closeErr // Подавляем предупреждение линтера
		}
	}()

	return s.processFeedbackRows(rows), nil
}

// getFeedbackQuery возвращает SQL запрос для получения всех отзывов.
func getFeedbackQuery() string {
	return `
        SELECT uf.id, uf.feedback_text, uf.contact_info, uf.created_at,
               uf.is_processed, u.username, u.telegram_id, u.first_name,
               uf.admin_response
        FROM user_feedback uf
        JOIN users u ON uf.user_id = u.id
        ORDER BY uf.created_at DESC
    `
}

// processFeedbackRows обрабатывает строки результата запроса отзывов.
func (s *BotService) processFeedbackRows(rows *sql.Rows) []map[string]interface{} {
	var feedbacks []map[string]interface{}

	for rows.Next() {
		feedback, err := s.scanFeedbackRow(rows)
		if err != nil {
			continue // Пропускаем ошибочные записи
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks
}

// scanFeedbackRow сканирует одну строку результата запроса отзывов.
func (s *BotService) scanFeedbackRow(rows *sql.Rows) (map[string]interface{}, error) {
	var (
		feedbackID   int
		feedbackText string
		contactInfo  sql.NullString
		createdAt    sql.NullTime
		isProcessed  bool
		username     sql.NullString
		telegramID   int64
		firstName    string
		adminResp    sql.NullString
	)

	err := rows.Scan(&feedbackID, &feedbackText, &contactInfo, &createdAt, &isProcessed,
		&username, &telegramID, &firstName, &adminResp)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	feedback := map[string]interface{}{
		"id":            feedbackID,
		"feedback_text": feedbackText,
		"created_at":    createdAt.Time,
		"telegram_id":   telegramID,
		"first_name":    firstName,
		"is_processed":  isProcessed,
	}

	// Добавляем опциональные поля
	feedback["username"] = getStringValue(username)
	feedback["contact_info"] = getStringValue(contactInfo)
	feedback["admin_response"] = getStringValue(adminResp)

	return feedback, nil
}

// getStringValue возвращает строковое значение из sql.NullString.
func getStringValue(nullStr sql.NullString) interface{} {
	if nullStr.Valid {
		return nullStr.String
	}

	return nil
}

// UpdateFeedbackStatus обновляет статус отзыва (обработан/не обработан).
func (s *BotService) UpdateFeedbackStatus(feedbackID int, isProcessed bool) error {
	query := `
		UPDATE user_feedback
		SET is_processed = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.DB.GetConnection().ExecContext(context.Background(), query, isProcessed, feedbackID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса отзыва: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить количество измененных строк: %w", err)
	}

	if rowsAffected == 0 {
		return errorsPkg.ErrFeedbackNotFound
	}

	return nil
}

// ArchiveFeedback архивирует отзыв.
func (s *BotService) ArchiveFeedback(feedbackID int) error {
	query := `
		UPDATE user_feedback
		SET is_processed = true, updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.DB.GetConnection().ExecContext(context.Background(), query, feedbackID)
	if err != nil {
		return fmt.Errorf("ошибка архивирования отзыва: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return errorsPkg.ErrFeedbackNotFound
	}

	return nil
}

// DeleteFeedback удаляет отзыв из базы данных.
func (s *BotService) DeleteFeedback(feedbackID int) error {
	query := `DELETE FROM user_feedback WHERE id = $1`

	result, err := s.DB.GetConnection().ExecContext(context.Background(), query, feedbackID)
	if err != nil {
		return fmt.Errorf("ошибка удаления отзыва: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить количество удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return errorsPkg.ErrFeedbackNotFound
	}

	return nil
}

// MarkFeedbackProcessed помечает отзыв как обработанный с ответом.
func (s *BotService) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	err := s.DB.MarkFeedbackProcessed(feedbackID, adminResponse)
	if err != nil {
		return fmt.Errorf("failed to mark feedback processed: %w", err)
	}

	return nil
}

// DeleteAllProcessedFeedbacks удаляет все обработанные отзывы.
func (s *BotService) DeleteAllProcessedFeedbacks() (int, error) {
	query := `DELETE FROM user_feedback WHERE is_processed = true`

	result, err := s.DB.GetConnection().ExecContext(context.Background(), query)
	if err != nil {
		return 0, fmt.Errorf("ошибка удаления обработанных отзывов: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества удаленных строк: %w", err)
	}

	return int(rowsAffected), nil
}

// UnarchiveFeedback возвращает отзыв в активные (убирает флаг is_processed).
func (s *BotService) UnarchiveFeedback(feedbackID int) error {
	query := `
		UPDATE user_feedback
		SET is_processed = false, updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.DB.GetConnection().ExecContext(context.Background(), query, feedbackID)
	if err != nil {
		return fmt.Errorf("ошибка возврата отзыва в активные: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return errorsPkg.ErrFeedbackNotFound
	}

	return nil
}

// ===== КЭШИРОВАННЫЕ МЕТОДЫ =====

// GetCachedLanguages получает языки из кэша или загружает из БД.
func (s *BotService) GetCachedLanguages(lang string) ([]*models.Language, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if languages, found := s.Cache.GetLanguages(context.Background(), lang); found {
		return languages, nil
	}

	// Загружаем из БД
	languages, err := s.DB.GetLanguages()
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	// Сохраняем в кэш
	s.Cache.SetLanguages(context.Background(), lang, languages)

	return languages, nil
}

// GetCachedInterests получает интересы из кэша или загружает из БД.
func (s *BotService) GetCachedInterests(lang string) (map[int]string, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if interests, found := s.Cache.GetInterests(context.Background(), lang); found {
		return interests, nil
	}

	// Загружаем из БД и локализуем
	interests, err := s.Localizer.GetInterests(lang)
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	// Сохраняем в кэш
	s.Cache.SetInterests(context.Background(), lang, interests)

	return interests, nil
}

// GetCachedUser получает пользователя из кэша или загружает из БД.
func (s *BotService) GetCachedUser(telegramID int64) (*models.User, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if user, found := s.Cache.GetUser(context.Background(), telegramID); found {
		return user, nil
	}

	// Загружаем из БД
	user, err := s.DB.GetUserByTelegramID(telegramID)
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	// Сохраняем в кэш
	s.Cache.SetUser(context.Background(), user)

	return user, nil
}

// GetCachedTranslations получает переводы из кэша или загружает из файлов.
func (s *BotService) GetCachedTranslations(lang string) (map[string]string, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if translations, found := s.Cache.GetTranslations(context.Background(), lang); found {
		return translations, nil
	}

	// Загружаем из файлов локализации
	// Здесь нужно будет добавить метод в Localizer для получения всех переводов
	// Пока что возвращаем пустую карту
	translations := make(map[string]string)

	// Сохраняем в кэш
	s.Cache.SetTranslations(context.Background(), lang, translations)

	return translations, nil
}

// UpdateCachedUser обновляет пользователя в БД и кэше.
func (s *BotService) UpdateCachedUser(user *models.User) error {
	// Обновляем в БД
	err := s.DB.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Обновляем в кэше
	s.Cache.SetUser(context.Background(), user)

	return nil
}

// InvalidateUserCache инвалидирует кэш пользователя.
func (s *BotService) InvalidateUserCache(userID int64) {
	s.InvalidationService.InvalidateUserData(userID)
}

// InvalidateStaticDataCache инвалидирует кэш статических данных.
func (s *BotService) InvalidateStaticDataCache() {
	s.InvalidationService.InvalidateStaticData()
}

// GetCacheStats возвращает статистику кэша.
func (s *BotService) GetCacheStats() map[string]interface{} {
	return s.MetricsService.GetMetrics()
}

// StopCache останавливает кэш-сервис.
func (s *BotService) StopCache() {
	s.Cache.Stop()
}

// ===== BATCH LOADING МЕТОДЫ =====

// GetUserWithAllData получает пользователя со всеми связанными данными одним запросом.
func (s *BotService) GetUserWithAllData(telegramID int64) (*database.UserWithAllData, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Пытаемся получить из кэша
	if userData, found := s.Cache.GetUser(context.Background(), telegramID); found {
		// Если пользователь есть в кэше, но нет полных данных, загружаем их
		if userData != nil {
			// Загружаем полные данные
			userData, err := s.BatchLoader.GetUserWithAllData(telegramID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user with all data: %w", err)
			}

			return userData, nil
		}
	}

	// Загружаем из БД одним запросом
	userData, err := s.BatchLoader.GetUserWithAllData(telegramID)
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	// Сохраняем в кэш
	s.Cache.SetUser(context.Background(), userData.User)

	return userData, nil
}

// BatchLoadUsersWithInterests загружает нескольких пользователей с их интересами одним запросом.
func (s *BotService) BatchLoadUsersWithInterests(telegramIDs []int64) (map[int64]*database.UserWithInterests, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	users, err := s.BatchLoader.BatchLoadUsersWithInterests(telegramIDs)
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	// Сохраняем пользователей в кэш
	for _, userData := range users {
		s.Cache.SetUser(context.Background(), userData.User)
	}

	return users, nil
}

// BatchLoadInterestsWithTranslations загружает интересы с переводами для нескольких языков.
func (s *BotService) BatchLoadInterestsWithTranslations(languages []string) (map[string]map[int]string, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	interests, err := s.BatchLoader.BatchLoadInterestsWithTranslations(languages)
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	// Сохраняем в кэш
	for lang, langInterests := range interests {
		s.Cache.SetInterests(context.Background(), lang, langInterests)
	}

	return interests, nil
}

// BatchLoadLanguagesWithTranslations загружает языки с переводами для нескольких языков.
func (s *BotService) BatchLoadLanguagesWithTranslations(languages []string) (map[string][]*models.Language, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	langs, err := s.BatchLoader.BatchLoadLanguagesWithTranslations(languages)
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	// Сохраняем в кэш
	for lang, langList := range langs {
		s.Cache.SetLanguages(context.Background(), lang, langList)
	}

	return langs, nil
}

// BatchLoadUserInterests загружает интересы для нескольких пользователей одним запросом.
func (s *BotService) BatchLoadUserInterests(userIDs []int) (map[int][]int, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	interests, err := s.BatchLoader.BatchLoadUserInterests(userIDs)
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return interests, nil
}

// BatchLoadUsers загружает пользователей по Telegram ID одним запросом.
func (s *BotService) BatchLoadUsers(telegramIDs []int64) (map[int64]*models.User, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	users, err := s.BatchLoader.BatchLoadUsers(telegramIDs)
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	// Сохраняем в кэш
	for _, user := range users {
		s.Cache.SetUser(context.Background(), user)
	}

	return users, nil
}

// BatchLoadStats загружает статистику для нескольких типов одним запросом.
func (s *BotService) BatchLoadStats(statTypes []string) (map[string]map[string]interface{}, error) {
	start := time.Now()

	defer func() {
		s.MetricsService.RecordRequest(time.Since(start), true)
	}()

	// Загружаем из БД одним запросом
	stats, err := s.BatchLoader.BatchLoadStats(statTypes)
	if err != nil {
		s.MetricsService.RecordError()

		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return stats, nil
}

// getLanguageFlag возвращает флаг для языка.
func (s *BotService) getLanguageFlag(languageCode string) string {
	switch languageCode {
	case "ru":
		return "🇷🇺"
	case "en":
		return "🇺🇸"
	case "es":
		return "🇪🇸"
	case "zh":
		return "🇨🇳"
	default:
		return "🌍"
	}
}
