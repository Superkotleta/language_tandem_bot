package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type User struct {
	TelegramID int64
	Username   string
	FirstName  string
	State      string
}

var db *pgxpool.Pool

func main() {
	_ = godotenv.Load()

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	dbURL := os.Getenv("DATABASE_URL")

	// Подключение к БД
	var err error
	db, err = connectDB(dbURL)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Запуск Telegram-бота
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatalf("❌ Ошибка запуска Telegram-бота: %v", err)
	}

	bot.Debug = true

	// АЛЬТЕРНАТИВНЫЙ способ удаления webhook
	deleteWebhookConfig := tgbotapi.DeleteWebhookConfig{}
	_, err = bot.Request(deleteWebhookConfig)
	if err != nil {
		log.Printf("⚠️ Предупреждение: не удалось удалить webhook: %v", err)
	}

	fmt.Printf("✅ Бот запущен: @%s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := bot.GetUpdatesChan(u)

	// Главный цикл
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := update.Message
		tgID := msg.From.ID

		if msg.IsCommand() {
			switch msg.Command() {
			case "start":
				_ = FindOrCreateUser(db, tgID, msg.From.UserName, msg.From.FirstName)
				_ = UpdateUserState(db, tgID, "ожидает_ввода_имени")
				reply := tgbotapi.NewMessage(msg.Chat.ID, "Привет! Как тебя зовут?")
				bot.Send(reply)
			}
		} else {
			state, _ := GetUserState(db, tgID)
			if state == "ожидает_ввода_имени" {
				_ = SaveFirstName(db, tgID, msg.Text)
				_ = UpdateUserState(db, tgID, "")
				reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Приятно познакомиться, %s!", msg.Text))
				bot.Send(reply)
			} else {
				bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Напиши /start, чтобы начать."))
			}
		}
	}
}

func connectDB(connStr string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	return pool, pool.Ping(context.Background())
}

func FindOrCreateUser(pool *pgxpool.Pool, telegramID int64, username, firstName string) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (telegram_id) DO NOTHING`
	_, err := pool.Exec(context.Background(), query, telegramID, username, firstName)
	return err
}

func UpdateUserState(pool *pgxpool.Pool, telegramID int64, state string) error {
	_, err := pool.Exec(context.Background(), `
		UPDATE users SET state = $1 WHERE telegram_id = $2`,
		state, telegramID)
	return err
}

func GetUserState(pool *pgxpool.Pool, telegramID int64) (string, error) {
	var state string
	err := pool.QueryRow(context.Background(), `
		SELECT state FROM users WHERE telegram_id = $1`, telegramID).Scan(&state)
	if err != nil {
		return "", err
	}
	return state, nil
}

func SaveFirstName(pool *pgxpool.Pool, telegramID int64, name string) error {
	_, err := pool.Exec(context.Background(), `
		UPDATE users SET first_name = $1 WHERE telegram_id = $2`,
		name, telegramID)
	return err
}
