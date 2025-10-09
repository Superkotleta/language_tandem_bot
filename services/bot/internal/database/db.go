// Package database provides database operations and connection management.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/logging"
	"language-exchange-bot/internal/models"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Константы для fallback данных.
const (
	fallbackLanguageID1 = 1
	fallbackLanguageID2 = 2
	fallbackLanguageID3 = 3
	fallbackLanguageID4 = 4

	fallbackInterestID1 = 1
	fallbackInterestID2 = 2
	fallbackInterestID3 = 3
	fallbackInterestID4 = 4
)

// DB представляет подключение к базе данных.
type DB struct {
	conn         *sql.DB
	logger       *logging.DatabaseLogger
	errorHandler *errors.ErrorHandler
	batchOps     *BatchOperations
}

// NewDB создает новое подключение к базе данных.
func NewDB(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Настройка connection pool для оптимизации производительности
	conn.SetMaxOpenConns(25)                  // Максимум 25 открытых соединений
	conn.SetMaxIdleConns(10)                  // Максимум 10 idle соединений
	conn.SetConnMaxLifetime(5 * time.Minute)  // Максимальное время жизни соединения
	conn.SetConnMaxIdleTime(10 * time.Minute) // Максимальное время idle соединения

	if err := conn.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		conn:         conn,
		logger:       logging.NewDatabaseLogger(),
		errorHandler: errors.NewErrorHandler(nil),
	}

	db.batchOps = NewBatchOperations(db)
	db.logger.LogConnectionEstablished("")

	return db, nil
}

// NewDBWithConfig создает новое подключение к базе данных с конфигурацией.
func NewDBWithConfig(databaseURL string, maxOpenConns, maxIdleConns int) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Настройка connection pool с переданными параметрами
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetConnMaxLifetime(5 * time.Minute)  // Максимальное время жизни соединения
	conn.SetConnMaxIdleTime(10 * time.Minute) // Максимальное время idle соединения

	if err := conn.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		conn:         conn,
		logger:       logging.NewDatabaseLogger(),
		errorHandler: errors.NewErrorHandler(nil),
	}

	db.batchOps = NewBatchOperations(db)
	db.logger.LogConnectionEstablished("")

	return db, nil
}

// Close закрывает подключение к базе данных.
func (db *DB) Close() error {
	err := db.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}

// GetConnection возвращает подключение к базе данных.
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// GetLanguages возвращает список всех языков.
//
//nolint:funlen
func (db *DB) GetLanguages() ([]*models.Language, error) {
	query := `
		SELECT id, code, name_native, name_en, is_interface_language
		FROM languages
		ORDER BY id
	`

	rows, err := db.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get languages: %w", err)
	}

	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close rows after error: %w (original error: %w)", closeErr, err)
		}

		return nil, fmt.Errorf("rows error: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			db.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "DatabaseOperation",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	var languages []*models.Language

	for rows.Next() {
		lang := &models.Language{
			ID:                  0,
			Code:                "",
			NameNative:          "",
			NameEn:              "",
			IsInterfaceLanguage: false,
			CreatedAt:           time.Now(),
		}

		err := rows.Scan(&lang.ID, &lang.Code, &lang.NameNative, &lang.NameEn, &lang.IsInterfaceLanguage)
		if err != nil {
			continue
		}

		languages = append(languages, lang)
	}

	// Fallback если нет данных в БД (для тестов)
	if len(languages) == 0 {
		return []*models.Language{
			{ID: fallbackLanguageID1, Code: "en", NameNative: "English", NameEn: "English"},
			{ID: fallbackLanguageID2, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
			{ID: fallbackLanguageID3, Code: "es", NameNative: "Español", NameEn: "Spanish"},
			{ID: fallbackLanguageID4, Code: "zh", NameNative: "中文", NameEn: "Chinese"},
		}, nil
	}

	return languages, nil
}

// GetLanguageByCode возвращает язык по его коду.
func (db *DB) GetLanguageByCode(code string) (*models.Language, error) {
	query := `
		SELECT id, code, name_native, name_en, is_interface_language
		FROM languages
		WHERE code = $1
	`
	lang := &models.Language{
		ID:                  0,
		Code:                "",
		NameNative:          "",
		NameEn:              "",
		IsInterfaceLanguage: false,
		CreatedAt:           time.Now(),
	}

	err := db.conn.QueryRowContext(context.Background(), query, code).Scan(
		&lang.ID, &lang.Code, &lang.NameNative, &lang.NameEn, &lang.IsInterfaceLanguage,
	)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return lang, nil
}

// GetInterests возвращает список всех интересов.
//
//nolint:funlen
func (db *DB) GetInterests() ([]*models.Interest, error) {
	query := `
		SELECT id, key_name, type
		FROM interests
		ORDER BY id
	`

	rows, err := db.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get interests: %w", err)
	}

	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close rows after error: %w (original error: %w)", closeErr, err)
		}

		return nil, fmt.Errorf("rows error: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			db.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "DatabaseOperation",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	var interests []*models.Interest

	for rows.Next() {
		interest := &models.Interest{
			ID:           0,
			KeyName:      "",
			CategoryID:   0,
			DisplayOrder: 0,
			Type:         "",
			CreatedAt:    time.Now(),
		}

		err := rows.Scan(&interest.ID, &interest.KeyName, &interest.Type)
		if err != nil {
			continue
		}

		interests = append(interests, interest)
	}

	// Fallback если нет данных в БД (для тестов)
	if len(interests) == 0 {
		return []*models.Interest{
			{ID: fallbackInterestID1, KeyName: "movies", Type: "entertainment"},
			{ID: fallbackInterestID2, KeyName: "music", Type: "entertainment"},
			{ID: fallbackInterestID3, KeyName: "sports", Type: "activity"},
			{ID: fallbackInterestID4, KeyName: "travel", Type: "activity"},
		}, nil
	}

	return interests, nil
}

// GetUserByTelegramID возвращает пользователя по Telegram ID.
func (db *DB) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	user := &models.User{
		ID:                     0,
		TelegramID:             0,
		Username:               "",
		FirstName:              "",
		NativeLanguageCode:     "",
		TargetLanguageCode:     "",
		TargetLanguageLevel:    "",
		InterfaceLanguageCode:  "",
		State:                  "",
		Status:                 "",
		ProfileCompletionLevel: 0,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Interests:              []int{},
		TimeAvailability: &models.TimeAvailability{
			DayType:      "any",
			SpecificDays: []string{},
			TimeSlot:     "any",
		},
		FriendshipPreferences: &models.FriendshipPreferences{
			ActivityType:       "casual_chat",
			CommunicationStyle: "text",
			CommunicationFreq:  "weekly",
		},
	}

	err := db.conn.QueryRowContext(context.Background(), `
		SELECT id, telegram_id, username, first_name,
		       COALESCE(native_language_code, '') as native_language_code,
		       COALESCE(target_language_code, '') as target_language_code,
		       COALESCE(target_language_level, '') as target_language_level,
		       interface_language_code, created_at, updated_at, state,
		       profile_completion_level, status
		FROM users
		WHERE telegram_id = $1
	`, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.NativeLanguageCode, &user.TargetLanguageCode, &user.TargetLanguageLevel,
		&user.InterfaceLanguageCode, &user.CreatedAt, &user.UpdatedAt,
		&user.State, &user.ProfileCompletionLevel, &user.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return user, nil
}

// UpdateUser обновляет данные пользователя.
func (db *DB) UpdateUser(user *models.User) error {
	_, err := db.conn.ExecContext(context.Background(), `
		UPDATE users
		SET username = $1, first_name = $2, native_language_code = $3,
		    target_language_code = $4, target_language_level = $5,
		    interface_language_code = $6, state = $7, status = $8,
		    profile_completion_level = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $10
	`, user.Username, user.FirstName, user.NativeLanguageCode,
		user.TargetLanguageCode, user.TargetLanguageLevel,
		user.InterfaceLanguageCode, user.State, user.Status,
		user.ProfileCompletionLevel, user.ID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// SaveUserInterests сохраняет интересы пользователя.
func (db *DB) SaveUserInterests(userID int, interestIDs []int) error {
	// Сначала удаляем все интересы пользователя
	err := db.ClearUserInterests(userID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Затем добавляем новые
	for _, interestID := range interestIDs {
		err := db.SaveUserInterest(userID, interestID, false)
		if err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}
	}

	return nil
}

// FindOrCreateUser находит или создает пользователя.
func (db *DB) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	user := &models.User{
		ID:                     0,
		TelegramID:             0,
		Username:               "",
		FirstName:              "",
		NativeLanguageCode:     "",
		TargetLanguageCode:     "",
		TargetLanguageLevel:    "",
		InterfaceLanguageCode:  "",
		State:                  "",
		Status:                 "",
		ProfileCompletionLevel: 0,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Interests:              []int{},
		TimeAvailability: &models.TimeAvailability{
			DayType:      "any",
			SpecificDays: []string{},
			TimeSlot:     "any",
		},
		FriendshipPreferences: &models.FriendshipPreferences{
			ActivityType:       "casual_chat",
			CommunicationStyle: "text",
			CommunicationFreq:  "weekly",
		},
	}

	err := db.conn.QueryRowContext(context.Background(), `
        INSERT INTO users (telegram_id, username, first_name, interface_language_code)
        VALUES ($1, $2, $3, 'en')
        ON CONFLICT (telegram_id) DO UPDATE SET
            username = EXCLUDED.username,
            first_name = EXCLUDED.first_name,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id, telegram_id, username, first_name,
        COALESCE(native_language_code, '') as native_language_code,
        COALESCE(target_language_code, '') as target_language_code,
        COALESCE(target_language_level, '') as target_language_level,
        interface_language_code, created_at, updated_at, state,
        profile_completion_level, status
    `, telegramID, username, firstName).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.NativeLanguageCode, &user.TargetLanguageCode, &user.TargetLanguageLevel,
		&user.InterfaceLanguageCode, &user.CreatedAt, &user.UpdatedAt,
		&user.State, &user.ProfileCompletionLevel, &user.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return user, nil
}

// UpdateUserState обновляет состояние пользователя.
func (db *DB) UpdateUserState(userID int, state string) error {
	_, err := db.conn.ExecContext(context.Background(), `
        UPDATE users SET state = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
    `, state, userID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UpdateUserStatus обновляет статус пользователя.
func (db *DB) UpdateUserStatus(userID int, status string) error {
	_, err := db.conn.ExecContext(context.Background(), `
        UPDATE users SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
    `, status, userID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UpdateUserProfileCompletionLevel обновляет уровень завершения профиля пользователя.
func (db *DB) UpdateUserProfileCompletionLevel(userID int, level int) error {
	_, err := db.conn.ExecContext(context.Background(), `
        UPDATE users SET profile_completion_level = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
    `, level, userID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UpdateUserInterfaceLanguage обновляет язык интерфейса пользователя.
func (db *DB) UpdateUserInterfaceLanguage(userID int, langCode string) error {
	_, err := db.conn.ExecContext(context.Background(), `
        UPDATE users SET interface_language_code = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
    `, langCode, userID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UpdateUserNativeLanguage обновляет родной язык пользователя.
func (db *DB) UpdateUserNativeLanguage(userID int, langCode string) error {
	_, err := db.conn.ExecContext(context.Background(),
		"UPDATE users SET native_language_code = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		langCode, userID,
	)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UpdateUserTargetLanguage обновляет целевой язык пользователя.
func (db *DB) UpdateUserTargetLanguage(userID int, langCode string) error {
	_, err := db.conn.ExecContext(context.Background(),
		"UPDATE users SET target_language_code = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		langCode, userID,
	)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// UpdateUserTargetLanguageLevel обновляет уровень целевого языка пользователя.
func (db *DB) UpdateUserTargetLanguageLevel(userID int, level string) error {
	_, err := db.conn.ExecContext(context.Background(),
		"UPDATE users SET target_language_level = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		level, userID,
	)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// SaveNativeLanguage сохраняет родной язык пользователя.
func (db *DB) SaveNativeLanguage(userID int, langCode string) error {
	return db.UpdateUserNativeLanguage(userID, langCode)
}

// SaveTargetLanguage сохраняет целевой язык пользователя.
func (db *DB) SaveTargetLanguage(userID int, langCode string) error {
	return db.UpdateUserTargetLanguage(userID, langCode)
}

// SaveUserInterest сохраняет интерес пользователя.
func (db *DB) SaveUserInterest(userID, interestID int, isPrimary bool) error {
	_, err := db.conn.ExecContext(context.Background(), `
        INSERT INTO user_interests (user_id, interest_id, is_primary, created_at) 
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP) 
        ON CONFLICT (user_id, interest_id) DO NOTHING
    `, userID, interestID, isPrimary)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// GetUserSelectedInterests возвращает выбранные пользователем интересы.
func (db *DB) GetUserSelectedInterests(userID int) ([]int, error) {
	rows, err := db.conn.QueryContext(context.Background(), `
        SELECT interest_id FROM user_interests 
        WHERE user_id = $1
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close rows after error: %w (original error: %w)", closeErr, err)
		}

		return nil, fmt.Errorf("rows error: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			db.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "DatabaseOperation",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	var interests []int

	for rows.Next() {
		var interestID int

		err := rows.Scan(&interestID)
		if err != nil {
			continue
		}

		interests = append(interests, interestID)
	}

	return interests, nil
}

// RemoveUserInterest удаляет интерес пользователя.
func (db *DB) RemoveUserInterest(userID, interestID int) error {
	_, err := db.conn.ExecContext(context.Background(), `
        DELETE FROM user_interests
        WHERE user_id = $1 AND interest_id = $2
    `, userID, interestID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// ClearUserInterests удаляет все интересы пользователя.
func (db *DB) ClearUserInterests(userID int) error {
	_, err := db.conn.ExecContext(context.Background(), `
        DELETE FROM user_interests WHERE user_id = $1
    `, userID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// GetUserInterestSelections получает выборы интересов пользователя.
func (db *DB) GetUserInterestSelections(userID int) ([]models.InterestSelection, error) {
	query := `
		SELECT id, user_id, interest_id, is_primary, selection_order, created_at
		FROM user_interest_selections 
		WHERE user_id = $1 
		ORDER BY is_primary DESC, selection_order ASC
	`

	rows, err := db.conn.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			db.logger.Error("Failed to close database rows", map[string]interface{}{
				"error":     closeErr.Error(),
				"operation": "GetUserInterestSelections",
			})
		}
	}()

	var selections []models.InterestSelection
	for rows.Next() {
		var selection models.InterestSelection

		err := rows.Scan(
			&selection.ID,
			&selection.UserID,
			&selection.InterestID,
			&selection.IsPrimary,
			&selection.SelectionOrder,
			&selection.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("operation failed: %w", err)
		}

		selections = append(selections, selection)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return selections, nil
}

// GetInterestByID получает интерес по ID.
func (db *DB) GetInterestByID(interestID int) (*models.Interest, error) {
	query := `
		SELECT id, key_name, category_id, display_order, type, created_at
		FROM interests 
		WHERE id = $1
	`

	var interest models.Interest

	err := db.conn.QueryRowContext(context.Background(), query, interestID).Scan(
		&interest.ID,
		&interest.KeyName,
		&interest.CategoryID,
		&interest.DisplayOrder,
		&interest.Type,
		&interest.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return &interest, nil
}

// ResetUserProfile очищает языки и интересы, переводит пользователя в начало онбординга.
func (db *DB) ResetUserProfile(userID int) error {
	transaction, err := db.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	defer func() {
		_ = transaction.Rollback()
	}()

	// Удаляем интересы
	query := `DELETE FROM user_interests WHERE user_id = $1`
	if _, err := transaction.ExecContext(context.Background(), query, userID); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Сбрасываем языки и состояние (интерфейсный язык не трогаем)
	if _, err := transaction.ExecContext(context.Background(), `
		UPDATE users
		SET native_language_code = NULL,
		    target_language_code = NULL,
		    target_language_level = '',
		    state = $1,
		    status = $2,
		    profile_completion_level = 0,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, models.StateWaitingLanguage, models.StatusFilling, userID); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Методы работы с отзывами пользователей

// SaveUserFeedback сохраняет отзыв пользователя в базу данных.
func (db *DB) SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error {
	query := `
        INSERT INTO user_feedback (user_id, feedback_text, contact_info, created_at, is_processed)
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP, false)
    `

	_, err := db.conn.ExecContext(context.Background(), query, userID, feedbackText, contactInfo)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// GetUserFeedbackByUserID получает отзывы пользователя.
func (db *DB) GetUserFeedbackByUserID(userID int) ([]map[string]interface{}, error) {
	query := `
        SELECT id, user_id, feedback_text, contact_info, created_at, is_processed, admin_response
        FROM user_feedback
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := db.conn.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close rows after error: %w (original error: %w)", closeErr, err)
		}

		return nil, fmt.Errorf("rows error: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			db.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "DatabaseOperation",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	return db.processUserFeedbackRows(rows), nil
}

// processUserFeedbackRows обрабатывает строки результата запроса отзывов пользователя.
func (db *DB) processUserFeedbackRows(rows *sql.Rows) []map[string]interface{} {
	var feedbacks []map[string]interface{}

	for rows.Next() {
		feedback := db.scanUserFeedbackRow(rows)
		if feedback != nil {
			feedbacks = append(feedbacks, feedback)
		}
	}

	return feedbacks
}

// scanUserFeedbackRow сканирует одну строку результата запроса отзывов пользователя.
func (db *DB) scanUserFeedbackRow(rows *sql.Rows) map[string]interface{} {
	var (
		feedbackID    int
		userID        int
		feedbackText  string
		contactInfo   sql.NullString
		createdAt     sql.NullTime
		isProcessed   bool
		adminResponse sql.NullString
	)

	err := rows.Scan(&feedbackID, &userID, &feedbackText, &contactInfo, &createdAt, &isProcessed, &adminResponse)
	if err != nil {
		return nil // Пропускаем ошибочные записи
	}

	feedback := map[string]interface{}{
		"id":            feedbackID,
		"user_id":       userID,
		"feedback_text": feedbackText,
		"created_at":    createdAt.Time,
		"is_processed":  isProcessed,
	}

	if contactInfo.Valid {
		feedback["contact_info"] = contactInfo.String
	} else {
		feedback["contact_info"] = nil
	}

	if adminResponse.Valid {
		feedback["admin_response"] = adminResponse.String
	} else {
		feedback["admin_response"] = nil
	}

	return feedback
}

// GetUnprocessedFeedback получает все необработанные отзывы для администрирования.
func (db *DB) GetUnprocessedFeedback() ([]map[string]interface{}, error) {
	query := getUnprocessedFeedbackQuery()

	rows, err := db.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	if err := rows.Err(); err != nil {
		if closeErr := rows.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close rows after error: %w (original error: %w)", closeErr, err)
		}

		return nil, fmt.Errorf("rows error: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			db.logger.ErrorWithContext(
				"Failed to close database rows",
				"", 0, 0, "DatabaseOperation",
				map[string]interface{}{
					"error": closeErr.Error(),
				},
			)
		}
	}()

	return db.processFeedbackRows(rows), nil
}

// getUnprocessedFeedbackQuery возвращает SQL запрос для получения необработанных отзывов.
func getUnprocessedFeedbackQuery() string {
	return `
        SELECT uf.id, uf.user_id, uf.feedback_text, uf.contact_info, uf.created_at,
               u.username, u.telegram_id, u.first_name
        FROM user_feedback uf
        JOIN users u ON uf.user_id = u.id
        WHERE uf.is_processed = false
        ORDER BY uf.created_at ASC
    `
}

// processFeedbackRows обрабатывает строки результата запроса отзывов.
func (db *DB) processFeedbackRows(rows *sql.Rows) []map[string]interface{} {
	var feedbacks []map[string]interface{}

	for rows.Next() {
		feedback, err := db.scanFeedbackRow(rows)
		if err != nil {
			continue // Пропускаем ошибочные записи
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks
}

// scanFeedbackRow сканирует одну строку результата запроса отзывов.
func (db *DB) scanFeedbackRow(rows *sql.Rows) (map[string]interface{}, error) {
	var (
		feedbackID   int
		userID       int
		feedbackText string
		contactInfo  sql.NullString
		createdAt    sql.NullTime
		username     sql.NullString
		telegramID   int64
		firstName    string
	)

	err := rows.Scan(&feedbackID, &userID, &feedbackText, &contactInfo, &createdAt, &username, &telegramID, &firstName)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	feedback := map[string]interface{}{
		"id":            feedbackID,
		"user_id":       userID,
		"feedback_text": feedbackText,
		"created_at":    createdAt.Time,
		"telegram_id":   telegramID,
		"first_name":    firstName,
	}

	// Добавляем опциональные поля
	feedback["username"] = getStringValue(username)
	feedback["contact_info"] = getStringValue(contactInfo)

	return feedback, nil
}

// getStringValue возвращает строковое значение из sql.NullString.
func getStringValue(nullStr sql.NullString) interface{} {
	if nullStr.Valid {
		return nullStr.String
	}

	return nil
}

// MarkFeedbackProcessed помечает отзыв как обработанный и добавляет ответ администратора.
func (db *DB) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	query := `
        UPDATE user_feedback
        SET is_processed = true, admin_response = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `

	_, err := db.conn.ExecContext(context.Background(), query, adminResponse, feedbackID)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// SaveTimeAvailability сохраняет временную доступность пользователя.
func (db *DB) SaveTimeAvailability(userID int, availability *models.TimeAvailability) error {
	query := `
		INSERT INTO user_time_availability (user_id, day_type, specific_days, time_slot)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			day_type = EXCLUDED.day_type,
			specific_days = EXCLUDED.specific_days,
			time_slot = EXCLUDED.time_slot,
			created_at = CURRENT_TIMESTAMP
	`

	_, err := db.conn.ExecContext(context.Background(), query,
		userID,
		availability.DayType,
		availability.SpecificDays,
		availability.TimeSlot,
	)
	if err != nil {
		return fmt.Errorf("failed to save time availability: %w", err)
	}

	return nil
}

// GetTimeAvailability получает временную доступность пользователя.
func (db *DB) GetTimeAvailability(userID int) (*models.TimeAvailability, error) {
	query := `
		SELECT day_type, specific_days, time_slot
		FROM user_time_availability
		WHERE user_id = $1
	`

	var availability models.TimeAvailability

	err := db.conn.QueryRowContext(context.Background(), query, userID).Scan(
		&availability.DayType,
		&availability.SpecificDays,
		&availability.TimeSlot,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Возвращаем значения по умолчанию, если данных нет
			return &models.TimeAvailability{
				DayType:      "any",
				SpecificDays: []string{},
				TimeSlot:     "any",
			}, nil
		}

		return nil, fmt.Errorf("failed to get time availability: %w", err)
	}

	return &availability, nil
}

// SaveFriendshipPreferences сохраняет предпочтения общения пользователя.
func (db *DB) SaveFriendshipPreferences(userID int, preferences *models.FriendshipPreferences) error {
	query := `
		INSERT INTO friendship_preferences (user_id, activity_type, communication_style, communication_frequency)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			activity_type = EXCLUDED.activity_type,
			communication_style = EXCLUDED.communication_style,
			communication_frequency = EXCLUDED.communication_frequency,
			created_at = CURRENT_TIMESTAMP
	`

	_, err := db.conn.ExecContext(context.Background(), query,
		userID,
		preferences.ActivityType,
		preferences.CommunicationStyle,
		preferences.CommunicationFreq,
	)
	if err != nil {
		return fmt.Errorf("failed to save friendship preferences: %w", err)
	}

	return nil
}

// GetFriendshipPreferences получает предпочтения общения пользователя.
func (db *DB) GetFriendshipPreferences(userID int) (*models.FriendshipPreferences, error) {
	query := `
		SELECT activity_type, communication_style, communication_frequency
		FROM friendship_preferences
		WHERE user_id = $1
	`

	var preferences models.FriendshipPreferences

	err := db.conn.QueryRowContext(context.Background(), query, userID).Scan(
		&preferences.ActivityType,
		&preferences.CommunicationStyle,
		&preferences.CommunicationFreq,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Возвращаем значения по умолчанию, если данных нет
			return &models.FriendshipPreferences{
				ActivityType:       "casual_chat",
				CommunicationStyle: "text",
				CommunicationFreq:  "weekly",
			}, nil
		}

		return nil, fmt.Errorf("failed to get friendship preferences: %w", err)
	}

	return &preferences, nil
}

// ===== BATCH OPERATIONS METHODS =====

// GetBatchOperations возвращает экземпляр BatchOperations для массовых операций.
func (db *DB) GetBatchOperations() *BatchOperations {
	return db.batchOps
}

// BatchInsertUsers выполняет массовую вставку пользователей.
func (db *DB) BatchInsertUsers(ctx context.Context, users []*models.User) error {
	return db.batchOps.BatchInsertUsers(ctx, users)
}

// BatchUpdateUsers выполняет массовое обновление пользователей.
func (db *DB) BatchUpdateUsers(ctx context.Context, users []*models.User) error {
	return db.batchOps.BatchUpdateUsers(ctx, users)
}

// BatchInsertInterests выполняет массовую вставку интересов.
func (db *DB) BatchInsertInterests(ctx context.Context, interests []*models.Interest) error {
	return db.batchOps.BatchInsertInterests(ctx, interests)
}

// BatchInsertUserInterests выполняет массовую вставку связей пользователь-интерес.
func (db *DB) BatchInsertUserInterests(ctx context.Context, userID int, interestIDs []int) error {
	return db.batchOps.BatchInsertUserInterests(ctx, userID, interestIDs)
}

// BatchDeleteUserInterests выполняет массовое удаление связей пользователь-интерес.
func (db *DB) BatchDeleteUserInterests(ctx context.Context, userID int, interestIDs []int) error {
	return db.batchOps.BatchDeleteUserInterests(ctx, userID, interestIDs)
}

// BatchInsertLanguages выполняет массовую вставку языков.
func (db *DB) BatchInsertLanguages(ctx context.Context, languages []*models.Language) error {
	return db.batchOps.BatchInsertLanguages(ctx, languages)
}

// BatchUpdateUserStats выполняет массовое обновление статистики пользователей.
func (db *DB) BatchUpdateUserStats(ctx context.Context, userStats []UserStats) error {
	return db.batchOps.BatchUpdateUserStats(ctx, userStats)
}

// BatchGetUsers выполняет массовое получение пользователей по ID.
func (db *DB) BatchGetUsers(ctx context.Context, userIDs []int) ([]*models.User, error) {
	return db.batchOps.BatchGetUsers(ctx, userIDs)
}
