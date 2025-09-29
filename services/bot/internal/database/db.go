package database

import (
	"database/sql"
	"fmt"
	"language-exchange-bot/internal/models"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func NewDB(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// Language operations
func (db *DB) GetLanguages() ([]*models.Language, error) {
	query := `
		SELECT id, code, name_native, name_en
		FROM languages
		ORDER BY id
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get languages: %w", err)
	}
	defer rows.Close()

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
			{ID: 1, Code: "en", NameNative: "English", NameEn: "English"},
			{ID: 2, Code: "ru", NameNative: "Русский", NameEn: "Russian"},
			{ID: 3, Code: "es", NameNative: "Español", NameEn: "Spanish"},
			{ID: 4, Code: "zh", NameNative: "中文", NameEn: "Chinese"},
		}, nil
	}

	return languages, nil
}

func (db *DB) GetLanguageByCode(code string) (*models.Language, error) {
	query := `
		SELECT id, code, name_native, name_en
		FROM languages
		WHERE code = $1
	`
	lang := &models.Language{}
	err := db.conn.QueryRow(query, code).Scan(&lang.ID, &lang.Code, &lang.NameNative, &lang.NameEn)
	if err != nil {
		return nil, err
	}
	return lang, nil
}

// Interest operations
func (db *DB) GetInterests() ([]*models.Interest, error) {
	query := `
		SELECT id, key_name, type
		FROM interests
		ORDER BY id
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get interests: %w", err)
	}
	defer rows.Close()

	var interests []*models.Interest
	for rows.Next() {
		interest := &models.Interest{}
		err := rows.Scan(&interest.ID, &interest.Name, &interest.Type)
		if err != nil {
			continue
		}
		interests = append(interests, interest)
	}

	// Fallback если нет данных в БД (для тестов)
	if len(interests) == 0 {
		return []*models.Interest{
			{ID: 1, Name: "movies", Type: "entertainment"},
			{ID: 2, Name: "music", Type: "entertainment"},
			{ID: 3, Name: "sports", Type: "activity"},
			{ID: 4, Name: "travel", Type: "activity"},
		}, nil
	}

	return interests, nil
}

func (db *DB) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	user := &models.User{}

	err := db.conn.QueryRow(`
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
		return nil, err
	}

	return user, nil
}

func (db *DB) UpdateUser(user *models.User) error {
	_, err := db.conn.Exec(`
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

	return err
}

func (db *DB) SaveUserInterests(userID int64, interestIDs []int) error {
	// Сначала удаляем все интересы пользователя
	if err := db.ClearUserInterests(int(userID)); err != nil {
		return err
	}

	// Затем добавляем новые
	for _, interestID := range interestIDs {
		if err := db.SaveUserInterest(int(userID), interestID, false); err != nil {
			return err
		}
	}

	return nil
}

// User operations
func (db *DB) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	user := &models.User{}

	err := db.conn.QueryRow(`
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

	return user, err
}

func (db *DB) UpdateUserState(userID int, state string) error {
	_, err := db.conn.Exec(`
        UPDATE users SET state = $1, updated_at = NOW() WHERE id = $2
    `, state, userID)
	return err
}

func (db *DB) UpdateUserStatus(userID int, status string) error {
	_, err := db.conn.Exec(`
        UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2
    `, status, userID)
	return err
}

func (db *DB) UpdateUserInterfaceLanguage(userID int, langCode string) error {
	_, err := db.conn.Exec(`
        UPDATE users SET interface_language_code = $1, updated_at = NOW() WHERE id = $2
    `, langCode, userID)
	return err
}

// ✅ Новые методы для работы с языками пользователя
func (db *DB) UpdateUserNativeLanguage(userID int, langCode string) error {
	_, err := db.conn.Exec(
		"UPDATE users SET native_language_code = $1, updated_at = NOW() WHERE id = $2",
		langCode, userID,
	)
	return err
}

func (db *DB) UpdateUserTargetLanguage(userID int, langCode string) error {
	_, err := db.conn.Exec(
		"UPDATE users SET target_language_code = $1, updated_at = NOW() WHERE id = $2",
		langCode, userID,
	)
	return err
}

func (db *DB) UpdateUserTargetLanguageLevel(userID int, level string) error {
	_, err := db.conn.Exec(
		"UPDATE users SET target_language_level = $1, updated_at = NOW() WHERE id = $2",
		level, userID,
	)
	return err
}

// ✅ Методы-обертки для совместимости
func (db *DB) SaveNativeLanguage(userID int, langCode string) error {
	return db.UpdateUserNativeLanguage(userID, langCode)
}

func (db *DB) SaveTargetLanguage(userID int, langCode string) error {
	return db.UpdateUserTargetLanguage(userID, langCode)
}

func (db *DB) SaveUserInterest(userID, interestID int, isPrimary bool) error {
	_, err := db.conn.Exec(`
        INSERT INTO user_interests (user_id, interest_id, is_primary, created_at) 
        VALUES ($1, $2, $3, NOW()) 
        ON CONFLICT (user_id, interest_id) DO NOTHING
    `, userID, interestID, isPrimary)
	return err
}

func (db *DB) GetUserSelectedInterests(userID int) ([]int, error) {
	rows, err := db.conn.Query(`
        SELECT interest_id FROM user_interests 
        WHERE user_id = $1
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interests []int
	for rows.Next() {
		var interestID int
		if err := rows.Scan(&interestID); err != nil {
			continue
		}
		interests = append(interests, interestID)
	}
	return interests, nil
}

func (db *DB) RemoveUserInterest(userID, interestID int) error {
	_, err := db.conn.Exec(`
        DELETE FROM user_interests 
        WHERE user_id = $1 AND interest_id = $2
    `, userID, interestID)
	return err
}

func (db *DB) ClearUserInterests(userID int) error {
	_, err := db.conn.Exec(`
        DELETE FROM user_interests WHERE user_id = $1
    `, userID)
	return err
}

// ResetUserProfile очищает языки и интересы, переводит пользователя в начало онбординга.
func (db *DB) ResetUserProfile(userID int) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Удаляем интересы
	if _, err := tx.Exec(`DELETE FROM user_interests WHERE user_id = $1`, userID); err != nil {
		return err
	}

	// Сбрасываем языки и состояние (интерфейсный язык не трогаем)
	if _, err := tx.Exec(`
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
		return err
	}

	return tx.Commit()
}

// Методы работы с отзывами пользователей

// SaveUserFeedback сохраняет отзыв пользователя в базу данных
func (db *DB) SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error {
	query := `
        INSERT INTO user_feedback (user_id, feedback_text, contact_info, created_at, is_processed)
        VALUES ($1, $2, $3, NOW(), false)
    `

	_, err := db.conn.Exec(query, userID, feedbackText, contactInfo)
	return err
}

// GetUserFeedbackByUserID получает отзывы пользователя
func (db *DB) GetUserFeedbackByUserID(userID int) ([]map[string]interface{}, error) {
	query := `
        SELECT id, feedback_text, contact_info, created_at, is_processed, admin_response
        FROM user_feedback
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := db.conn.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

// GetUnprocessedFeedback получает все необработанные отзывы для администрирования
func (db *DB) GetUnprocessedFeedback() ([]map[string]interface{}, error) {
	query := `
        SELECT uf.id, uf.feedback_text, uf.contact_info, uf.created_at,
               u.username, u.telegram_id, u.first_name
        FROM user_feedback uf
        JOIN users u ON uf.user_id = u.id
        WHERE uf.is_processed = false
        ORDER BY uf.created_at ASC
    `

	rows, err := db.conn.Query(query)
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
			username     sql.NullString
			telegramID   int64
			firstName    string
		)

		err := rows.Scan(&id, &feedbackText, &contactInfo, &createdAt, &username, &telegramID, &firstName)
		if err != nil {
			continue // Пропускаем ошибочные записи
		}

		feedback := map[string]interface{}{
			"id":            id,
			"feedback_text": feedbackText,
			"created_at":    createdAt.Time,
			"telegram_id":   telegramID,
			"first_name":    firstName,
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

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// MarkFeedbackProcessed помечает отзыв как обработанный и добавляет ответ администратора
func (db *DB) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	query := `
        UPDATE user_feedback
        SET is_processed = true, admin_response = $1
        WHERE id = $2
    `

	_, err := db.conn.Exec(query, adminResponse, feedbackID)
	return err
}
