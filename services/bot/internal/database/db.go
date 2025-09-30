// Package database provides database operations and connection management.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"language-exchange-bot/internal/models"

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
	conn *sql.DB
}

// NewDB создает новое подключение к базе данных.
func NewDB(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := conn.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn}, nil
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
func (db *DB) GetLanguages() ([]*models.Language, error) {
	query := `
		SELECT id, code, name_native, name_en
		FROM languages
		ORDER BY id
	`

	rows, err := db.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get languages: %w", err)
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			// Логируем ошибку закрытия, но не возвращаем её
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	var languages []*models.Language

	for rows.Next() {
		lang := &models.Language{}

		err := rows.Scan(&lang.ID, &lang.Code, &lang.NameNative, &lang.NameEn)
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
		SELECT id, code, name_native, name_en
		FROM languages
		WHERE code = $1
	`
	lang := &models.Language{}

	err := db.conn.QueryRowContext(context.Background(), query, code).Scan(
		&lang.ID, &lang.Code, &lang.NameNative, &lang.NameEn,
	)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	return lang, nil
}

// GetInterests возвращает список всех интересов.
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

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			// Логируем ошибку закрытия, но не возвращаем её
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	var interests []*models.Interest

	for rows.Next() {
		interest := &models.Interest{}

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
	user := &models.User{}

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
		    profile_completion_level = $9, updated_at = NOW()
		WHERE id = $10
	`, user.Username, user.FirstName, user.NativeLanguageCode,
		user.TargetLanguageCode, user.TargetLanguageLevel,
		user.InterfaceLanguageCode, user.State, user.Status,
		user.ProfileCompletionLevel, user.ID)

	return fmt.Errorf("operation failed: %w", err)
}

// SaveUserInterests сохраняет интересы пользователя.
func (db *DB) SaveUserInterests(userID int64, interestIDs []int) error {
	// Сначала удаляем все интересы пользователя
	err := db.ClearUserInterests(int(userID))
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Затем добавляем новые
	for _, interestID := range interestIDs {
		err := db.SaveUserInterest(int(userID), interestID, false)
		if err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}
	}

	return nil
}

// FindOrCreateUser находит или создает пользователя.
func (db *DB) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	user := &models.User{}

	err := db.conn.QueryRowContext(context.Background(), `
        INSERT INTO users (telegram_id, username, first_name, interface_language_code)
        VALUES ($1, $2, $3, 'en')
        ON CONFLICT (telegram_id) DO UPDATE SET
            username = EXCLUDED.username,
            first_name = EXCLUDED.first_name,
            updated_at = NOW()
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

	return user, fmt.Errorf("operation failed: %w", err)
}

// UpdateUserState обновляет состояние пользователя.
func (db *DB) UpdateUserState(userID int, state string) error {
	_, err := db.conn.ExecContext(context.Background(), `
        UPDATE users SET state = $1, updated_at = NOW() WHERE id = $2
    `, state, userID)

	return fmt.Errorf("operation failed: %w", err)
}

// UpdateUserStatus обновляет статус пользователя.
func (db *DB) UpdateUserStatus(userID int, status string) error {
	_, err := db.conn.ExecContext(context.Background(), `
        UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2
    `, status, userID)

	return fmt.Errorf("operation failed: %w", err)
}

// UpdateUserInterfaceLanguage обновляет язык интерфейса пользователя.
func (db *DB) UpdateUserInterfaceLanguage(userID int, langCode string) error {
	_, err := db.conn.ExecContext(context.Background(), `
        UPDATE users SET interface_language_code = $1, updated_at = NOW() WHERE id = $2
    `, langCode, userID)

	return fmt.Errorf("operation failed: %w", err)
}

// UpdateUserNativeLanguage обновляет родной язык пользователя.
func (db *DB) UpdateUserNativeLanguage(userID int, langCode string) error {
	_, err := db.conn.ExecContext(context.Background(),
		"UPDATE users SET native_language_code = $1, updated_at = NOW() WHERE id = $2",
		langCode, userID,
	)

	return fmt.Errorf("operation failed: %w", err)
}

// UpdateUserTargetLanguage обновляет целевой язык пользователя.
func (db *DB) UpdateUserTargetLanguage(userID int, langCode string) error {
	_, err := db.conn.ExecContext(context.Background(),
		"UPDATE users SET target_language_code = $1, updated_at = NOW() WHERE id = $2",
		langCode, userID,
	)

	return fmt.Errorf("operation failed: %w", err)
}

// UpdateUserTargetLanguageLevel обновляет уровень целевого языка пользователя.
func (db *DB) UpdateUserTargetLanguageLevel(userID int, level string) error {
	_, err := db.conn.ExecContext(context.Background(),
		"UPDATE users SET target_language_level = $1, updated_at = NOW() WHERE id = $2",
		level, userID,
	)

	return fmt.Errorf("operation failed: %w", err)
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
        VALUES ($1, $2, $3, NOW()) 
        ON CONFLICT (user_id, interest_id) DO NOTHING
    `, userID, interestID, isPrimary)

	return fmt.Errorf("operation failed: %w", err)
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

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			// Логируем ошибку закрытия, но не возвращаем её
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
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

	return fmt.Errorf("operation failed: %w", err)
}

// ClearUserInterests удаляет все интересы пользователя.
func (db *DB) ClearUserInterests(userID int) error {
	_, err := db.conn.ExecContext(context.Background(), `
        DELETE FROM user_interests WHERE user_id = $1
    `, userID)

	return fmt.Errorf("operation failed: %w", err)
}

// ResetUserProfile очищает языки и интересы, переводит пользователя в начало онбординга.
func (db *DB) ResetUserProfile(userID int) error {
	tx, err := db.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	// Удаляем интересы
	query := `DELETE FROM user_interests WHERE user_id = $1`
	if _, err := tx.ExecContext(context.Background(), query, userID); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Сбрасываем языки и состояние (интерфейсный язык не трогаем)
	if _, err := tx.ExecContext(context.Background(), `
		UPDATE users
		SET native_language_code = NULL,
		    target_language_code = NULL,
		    target_language_level = '',
		    state = $1,
		    status = $2,
		    profile_completion_level = 0,
		    updated_at = NOW()
		WHERE id = $3
	`, models.StateWaitingLanguage, models.StatusFilling, userID); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Методы работы с отзывами пользователей

// SaveUserFeedback сохраняет отзыв пользователя в базу данных.
func (db *DB) SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error {
	query := `
        INSERT INTO user_feedback (user_id, feedback_text, contact_info, created_at, is_processed)
        VALUES ($1, $2, $3, NOW(), false)
    `

	_, err := db.conn.ExecContext(context.Background(), query, userID, feedbackText, contactInfo)

	return fmt.Errorf("operation failed: %w", err)
}

// GetUserFeedbackByUserID получает отзывы пользователя.
func (db *DB) GetUserFeedbackByUserID(userID int) ([]map[string]interface{}, error) {
	query := `
        SELECT id, feedback_text, contact_info, created_at, is_processed, admin_response
        FROM user_feedback
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := db.conn.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			// Логируем ошибку закрытия, но не возвращаем её
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	var feedbacks []map[string]interface{}

	for rows.Next() {
		var (
			id            int
			feedbackText  string
			contactInfo   sql.NullString
			createdAt     sql.NullTime
			isProcessed   bool
			adminResponse sql.NullString
		)

		err := rows.Scan(&id, &feedbackText, &contactInfo, &createdAt, &isProcessed, &adminResponse)
		if err != nil {
			continue // Пропускаем ошибочные записи
		}

		feedback := map[string]interface{}{
			"id":            id,
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

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// GetUnprocessedFeedback получает все необработанные отзывы для администрирования.
func (db *DB) GetUnprocessedFeedback() ([]map[string]interface{}, error) {
	query := getUnprocessedFeedbackQuery()

	rows, err := db.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	return db.processFeedbackRows(rows), nil
}

// getUnprocessedFeedbackQuery возвращает SQL запрос для получения необработанных отзывов.
func getUnprocessedFeedbackQuery() string {
	return `
        SELECT uf.id, uf.feedback_text, uf.contact_info, uf.created_at,
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
		id           int
		feedbackText string
		contactInfo  sql.NullString
		createdAt    sql.NullTime
		username     sql.NullString
		telegramID   int64
		firstName    string
	)

	err := rows.Scan(&id, &feedbackText, &contactInfo, &createdAt, &username, &telegramID, &firstName)
	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	feedback := map[string]interface{}{
		"id":            id,
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
        SET is_processed = true, admin_response = $1
        WHERE id = $2
    `

	_, err := db.conn.ExecContext(context.Background(), query, adminResponse, feedbackID)

	return fmt.Errorf("operation failed: %w", err)
}
