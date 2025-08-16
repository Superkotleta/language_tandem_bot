package database

import (
	"database/sql"
	"language-exchange-bot/internal/models"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func Connect(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	db := &DB{conn: conn}

	// Создаем таблицы если не существуют
	if err := db.createTables(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		telegram_id BIGINT UNIQUE NOT NULL,
		username TEXT,
		first_name TEXT,
		created_at TIMESTAMP DEFAULT NOW(),
		state TEXT DEFAULT '',
		profile_completion_level INT DEFAULT 0,
		status TEXT DEFAULT 'not_started' CHECK (status IN ('not_started', 'filling', 'ready', 'matched', 'waiting'))
	);
	
	CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
	`

	_, err := db.conn.Exec(query)
	return err
}

// CRUD операции для пользователей
func (db *DB) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	user := &models.User{}

	// Сначала пытаемся найти
	err := db.conn.QueryRow(
		"SELECT id, telegram_id, username, first_name, created_at, state, profile_completion_level, status FROM users WHERE telegram_id = $1",
		telegramID,
	).Scan(&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.CreatedAt, &user.State, &user.ProfileCompletionLevel, &user.Status)

	if err == sql.ErrNoRows {
		// Создаем нового пользователя
		err = db.conn.QueryRow(
			"INSERT INTO users (telegram_id, username, first_name) VALUES ($1, $2, $3) RETURNING id, created_at, state, profile_completion_level, status",
			telegramID, username, firstName,
		).Scan(&user.ID, &user.CreatedAt, &user.State, &user.ProfileCompletionLevel, &user.Status)

		if err != nil {
			return nil, err
		}

		user.TelegramID = telegramID
		user.Username = username
		user.FirstName = firstName
	} else if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *DB) UpdateUserState(userID int, state string) error {
	_, err := db.conn.Exec("UPDATE users SET state = $1 WHERE id = $2", state, userID)
	return err
}

func (db *DB) UpdateUserStatus(userID int, status string) error {
	_, err := db.conn.Exec("UPDATE users SET status = $1 WHERE id = $2", status, userID)
	return err
}
