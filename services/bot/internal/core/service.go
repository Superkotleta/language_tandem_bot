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
	DB                       *database.DB
	Localizer                *localization.Localizer
	FeedbackNotificationFunc func(data map[string]interface{}) error // функция для отправки уведомлений
}

func NewBotService(db *database.DB) *BotService {
	return &BotService{
		DB:        db,
		Localizer: localization.NewLocalizer(db.GetConnection()),
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
		int64(feedbackData["telegram_id"].(int)),
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
		result["username"] = username
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
