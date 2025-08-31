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
	// Сначала пытаемся найти
	err := db.conn.QueryRow(`
        SELECT id, telegram_id, username, first_name, 
               COALESCE(native_language_code, '') as native_language_code,
               COALESCE(target_language_code, '') as target_language_code,
               interface_language_code, created_at, updated_at, state, 
               profile_completion_level, status
        FROM users WHERE telegram_id = $1
    `, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.NativeLanguageCode, &user.TargetLanguageCode,
		&user.InterfaceLanguageCode, &user.CreatedAt, &user.UpdatedAt,
		&user.State, &user.ProfileCompletionLevel, &user.Status,
	)

	if err == sql.ErrNoRows {
		// Создаем нового пользователя
		err = db.conn.QueryRow(`
            INSERT INTO users (telegram_id, username, first_name, interface_language_code)
            VALUES ($1, $2, $3, 'en')
            RETURNING id, created_at, updated_at, state, profile_completion_level, status
        `, telegramID, username, firstName).Scan(
			&user.ID, &user.CreatedAt, &user.UpdatedAt,
			&user.State, &user.ProfileCompletionLevel, &user.Status,
		)

		if err != nil {
			return nil, err
		}

		user.TelegramID = telegramID
		user.Username = username
		user.FirstName = firstName
		user.InterfaceLanguageCode = "en"
		user.NativeLanguageCode = ""
		user.TargetLanguageCode = ""
	} else if err != nil {
		return nil, err
	}

	return user, nil
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
