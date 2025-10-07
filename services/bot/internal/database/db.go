package database

import (
	"database/sql"
	"fmt"
	"language-exchange-bot/internal/models"

	_ "github.com/lib/pq" // PostgreSQL driver
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

// User operations.
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

// ✅ Новые методы для работы с языками пользователя.
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

// ✅ Методы-обертки для совместимости.
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

// SaveUserFeedback сохраняет отзыв пользователя в базу данных.
func (db *DB) SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error {
	query := `
        INSERT INTO user_feedback (user_id, feedback_text, contact_info, created_at, is_processed)
        VALUES ($1, $2, $3, NOW(), false)
    `

	_, err := db.conn.Exec(query, userID, feedbackText, contactInfo)
	return err
}

// GetUserFeedbackByUserID получает отзывы пользователя.
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

// GetUnprocessedFeedback получает все необработанные отзывы для администрирования.
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

// GetAllFeedback получает все отзывы для администрирования.
func (db *DB) GetAllFeedback() ([]map[string]interface{}, error) {
	query := `
        SELECT uf.id, uf.feedback_text, uf.contact_info, uf.created_at, uf.is_processed, uf.admin_response,
               u.username, u.telegram_id, u.first_name
        FROM user_feedback uf
        JOIN users u ON uf.user_id = u.id
        ORDER BY uf.created_at DESC
    `

	rows, err := db.conn.Query(query)
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
			username      sql.NullString
			telegramID    int64
			firstName     string
		)

		err := rows.Scan(&id, &feedbackText, &contactInfo, &createdAt, &isProcessed, &adminResponse, &username, &telegramID, &firstName)
		if err != nil {
			continue // Пропускаем ошибочные записи
		}

		feedback := map[string]interface{}{
			"id":            id,
			"feedback_text": feedbackText,
			"created_at":    createdAt.Time,
			"is_processed":  isProcessed,
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

		if adminResponse.Valid {
			feedback["admin_response"] = adminResponse.String
		} else {
			feedback["admin_response"] = nil
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// MarkFeedbackProcessed помечает отзыв как обработанный и добавляет ответ администратора.
func (db *DB) MarkFeedbackProcessed(feedbackID int, adminResponse string) error {
	query := `
        UPDATE user_feedback
        SET is_processed = true, admin_response = $1
        WHERE id = $2
    `

	_, err := db.conn.Exec(query, adminResponse, feedbackID)
	return err
}

// GetUserDataForFeedback получает данные пользователя для формирования уведомления о новом отзыве.
func (db *DB) GetUserDataForFeedback(userID int) (map[string]interface{}, error) {
	query := `
        SELECT telegram_id, username, first_name
        FROM users
        WHERE id = $1
    `

	var (
		telegramID int64
		username   sql.NullString
		firstName  sql.NullString
	)

	err := db.conn.QueryRow(query, userID).Scan(&telegramID, &username, &firstName)
	if err != nil {
		return nil, err
	}

	userData := map[string]interface{}{
		"telegram_id": telegramID,
		"first_name":  firstName.String,
	}

	if username.Valid {
		userData["username"] = &username.String
	} else {
		userData["username"] = nil
	}

	return userData, nil
}

// DeleteFeedback удаляет отзыв из базы данных.
func (db *DB) DeleteFeedback(feedbackID int) error {
	query := `DELETE FROM user_feedback WHERE id = $1`
	_, err := db.conn.Exec(query, feedbackID)
	return err
}

// ArchiveFeedback архивирует отзыв (помечает как обработанный).
func (db *DB) ArchiveFeedback(feedbackID int) error {
	query := `UPDATE user_feedback SET is_processed = true WHERE id = $1`
	_, err := db.conn.Exec(query, feedbackID)
	return err
}

// UnarchiveFeedback разархивирует отзыв (помечает как необработанный).
func (db *DB) UnarchiveFeedback(feedbackID int) error {
	query := `UPDATE user_feedback SET is_processed = false WHERE id = $1`
	_, err := db.conn.Exec(query, feedbackID)
	return err
}

// UpdateFeedbackStatus обновляет статус отзыва.
func (db *DB) UpdateFeedbackStatus(feedbackID int, isProcessed bool) error {
	query := `UPDATE user_feedback SET is_processed = $1 WHERE id = $2`
	_, err := db.conn.Exec(query, isProcessed, feedbackID)
	return err
}

// DeleteAllProcessedFeedbacks удаляет все обработанные отзывы.
func (db *DB) DeleteAllProcessedFeedbacks() (int, error) {
	query := `DELETE FROM user_feedback WHERE is_processed = true`
	result, err := db.conn.Exec(query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}
