package core

import (
	"database/sql"
	"fmt"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
	"log"
	"strings"
)

type BotService struct {
	DB                       database.Database
	Localizer                *localization.Localizer
	FeedbackNotificationFunc func(data map[string]interface{}) error // функция для отправки уведомлений
}

func NewBotService(db *database.DB) *BotService {
	return &BotService{
		DB:        &databaseAdapter{db: db}, // Оборачиваем в адаптер
		Localizer: localization.NewLocalizer(db.GetConnection()),
	}
}

// NewBotServiceWithInterface создает BotService с уже готовым интерфейсом Database (для тестов).
func NewBotServiceWithInterface(db database.Database, localizer *localization.Localizer) *BotService {
	return &BotService{
		DB:        db,
		Localizer: localizer,
	}
}

// databaseAdapter адаптер для совместимости с интерфейсом Database.
type databaseAdapter struct {
	db *database.DB
}

// Реализуем все методы интерфейса, делегируя к db или создавая заглушки

func (a *databaseAdapter) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	return a.db.FindOrCreateUser(telegramID, username, firstName)
}

func (a *databaseAdapter) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	// Заглушка - используем FindOrCreateUser
	return a.db.FindOrCreateUser(telegramID, "", "")
}

func (a *databaseAdapter) UpdateUser(user *models.User) error {
	// Заглушка - обновляем основные поля
	a.db.UpdateUserState(user.ID, user.State)
	a.db.UpdateUserStatus(user.ID, user.Status)
	a.db.UpdateUserInterfaceLanguage(user.ID, user.InterfaceLanguageCode)
	return nil
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

func (a *databaseAdapter) UpdateUserProfileCompletionLevel(userID int, level int) error {
	// Заглушка - в реальной БД здесь было бы обновление
	return nil
}

func (a *databaseAdapter) ResetUserProfile(userID int) error {
	return a.db.ResetUserProfile(userID)
}

func (a *databaseAdapter) GetLanguages() ([]*models.Language, error) {
	// Заглушка - возвращаем базовые языки
	return []*models.Language{
		{ID: 1, Code: "en", NameNative: "English", NameEn: "English"},
		{ID: 2, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
		{ID: 3, Code: "es", NameNative: "Español", NameEn: "Spanish"},
		{ID: 4, Code: "zh", NameNative: "中文", NameEn: "Chinese"},
	}, nil
}

func (a *databaseAdapter) GetLanguageByCode(code string) (*models.Language, error) {
	languages, _ := a.GetLanguages()
	for _, lang := range languages {
		if lang.Code == code {
			return lang, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (a *databaseAdapter) GetInterests() ([]*models.Interest, error) {
	// Заглушка - возвращаем базовые интересы
	return []*models.Interest{
		{ID: 1, Name: "movies", Type: "entertainment"},
		{ID: 2, Name: "music", Type: "entertainment"},
		{ID: 3, Name: "sports", Type: "activity"},
		{ID: 4, Name: "travel", Type: "activity"},
	}, nil
}

func (a *databaseAdapter) GetUserSelectedInterests(userID int) ([]int, error) {
	return a.db.GetUserSelectedInterests(userID)
}

func (a *databaseAdapter) SaveUserInterests(userID int64, interestIDs []int) error {
	// Заглушка - сохраняем по одному
	for _, id := range interestIDs {
		if err := a.db.SaveUserInterest(int(userID), id, false); err != nil {
			return err
		}
	}
	return nil
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

func (a *databaseAdapter) GetUserFeedbackByUserID(userID int) ([]map[string]interface{}, error) {
	return a.db.GetUserFeedbackByUserID(userID)
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

func (a *databaseAdapter) GetUserDataForFeedback(userID int) (map[string]interface{}, error) {
	// Получаем пользователя по ID
	var telegramID int64
	var username, firstName string
	err := a.db.GetConnection().QueryRow(`
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

func (a *databaseAdapter) GetAllFeedback() ([]map[string]interface{}, error) {
	// Заглушка - возвращаем пустой список
	return []map[string]interface{}{}, nil
}

func (a *databaseAdapter) DeleteFeedback(feedbackID int) error {
	// Заглушка - ничего не делаем
	return nil
}

func (a *databaseAdapter) ArchiveFeedback(feedbackID int) error {
	// Заглушка - ничего не делаем
	return nil
}

func (a *databaseAdapter) UnarchiveFeedback(feedbackID int) error {
	// Заглушка - ничего не делаем
	return nil
}

func (a *databaseAdapter) UpdateFeedbackStatus(feedbackID int, isProcessed bool) error {
	// Заглушка - в реальной БД здесь было бы обновление статуса
	return nil
}

func (a *databaseAdapter) DeleteAllProcessedFeedbacks() (int, error) {
	// Заглушка - в реальной БД здесь было бы удаление обработанных отзывов
	return 0, nil
}

// SetFeedbackNotificationFunc устанавливает функцию для отправки уведомлений о новых отзывах.
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

// ValidateFeedback проверяет корректность отзыва по длине.
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

// SaveUserFeedback сохраняет отзыв пользователя и отправляет уведомления.
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

// GetUserDataForFeedback получает данные пользователя для формирования уведомления о новом отзыве.
func (s *BotService) GetUserDataForFeedback(userID int) (map[string]interface{}, error) {
	return s.DB.GetUserDataForFeedback(userID)
}

// GetAllUnprocessedFeedback получает все необработанные отзывы для администратора.
func (s *BotService) GetAllUnprocessedFeedback() ([]map[string]interface{}, error) {
	return s.DB.GetUnprocessedFeedback()
}

// GetAllFeedback получает все отзывы для администратора.
func (s *BotService) GetAllFeedback() ([]map[string]interface{}, error) {
	return s.DB.GetAllFeedback()
}

// UpdateFeedbackStatus обновляет статус отзыва (обработан/не обработан).
func (s *BotService) UpdateFeedbackStatus(feedbackID int, isProcessed bool) error {
	return s.DB.UpdateFeedbackStatus(feedbackID, isProcessed)
}

// ArchiveFeedback архивирует отзыв.
func (s *BotService) ArchiveFeedback(feedbackID int) error {
	return s.DB.ArchiveFeedback(feedbackID)
}

// DeleteFeedback удаляет отзыв из базы данных.
func (s *BotService) DeleteFeedback(feedbackID int) error {
	return s.DB.DeleteFeedback(feedbackID)
}

// MarkFeedbackProcessed помечает отзыв как обработанный с ответом.
func (s *BotService) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	return s.DB.MarkFeedbackProcessed(feedbackID, adminResponse)
}

// DeleteAllProcessedFeedbacks удаляет все обработанные отзывы.
func (s *BotService) DeleteAllProcessedFeedbacks() (int, error) {
	return s.DB.DeleteAllProcessedFeedbacks()
}

// UnarchiveFeedback возвращает отзыв в активные (убирает флаг is_processed).
func (s *BotService) UnarchiveFeedback(feedbackID int) error {
	return s.DB.UnarchiveFeedback(feedbackID)
}
