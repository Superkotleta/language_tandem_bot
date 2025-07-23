package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	TelegramID int64
	Username   string
	FirstName  string
	CreatedAt  time.Time
	State      string
}

// Глобальный пул соединений
var db *pgxpool.Pool

func main() {
	// Строка подключения, можете добавить .env и использовать os.Getenv("DB_URL")
	dbURL := "postgres://postgres:root@localhost:5432/language_exchange_db"

	var err error
	db, err = connectDB(dbURL)
	if err != nil {
		fmt.Println("❌ Ошибка подключения к БД:", err)
		return
	}
	defer db.Close()

	// Примерные данные пришедшие из Telegram API
	telegramID := int64(123456789)
	username := "exampleuser"
	firstName := "Иван"

	err = FindOrCreateUser(db, telegramID, username, firstName)
	if err != nil {
		fmt.Println("❌ Ошибка при добавлении пользователя:", err)
	} else {
		fmt.Println("✅ Пользователь добавлен или уже существует.")
	}
}

func connectDB(connStr string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга строки подключения: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул подключений: %w", err)
	}

	// Тестовое подключение
	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	return pool, nil
}

func FindOrCreateUser(pool *pgxpool.Pool, telegramID int64, username, firstName string) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (telegram_id) DO NOTHING
	`
	_, err := pool.Exec(context.Background(), query, telegramID, username, firstName)
	return err
}
