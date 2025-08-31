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
        interface_language_code, created_at, updated_at, state,
        profile_completion_level, status
    `, telegramID, username, firstName).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.NativeLanguageCode, &user.TargetLanguageCode,
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
