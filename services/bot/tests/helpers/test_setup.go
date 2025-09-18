package helpers

import (
	"database/sql"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/localization"
	"language-exchange-bot/internal/models"
	"language-exchange-bot/tests/mocks"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SetupTestBot создает тестовый бот с моками
func SetupTestBot(mockDB database.Database) (*mocks.TelegramHandlerWrapper, *core.BotService) {
	// Создаем простой локализатор для тестов - используем реальный с nil DB
	localizer := localization.NewLocalizer(nil)

	// Создаем тестовый сервис
	service := core.NewBotServiceWithInterface(mockDB, localizer)

	// Создаем обертку хендлера для тестирования
	wrapper := &mocks.TelegramHandlerWrapper{
		Service:        service,
		SentMessages:   make([]tgbotapi.MessageConfig, 0),
		SentCallbacks:  make([]tgbotapi.CallbackConfig, 0),
		EditedMessages: make([]tgbotapi.EditMessageTextConfig, 0),
	}

	return wrapper, service
}

// CreateTestMessage создает тестовое сообщение Telegram
func CreateTestMessage(text string, userID int64, username string) *tgbotapi.Message {
	return &tgbotapi.Message{
		MessageID: int(time.Now().Unix()),
		From: &tgbotapi.User{
			ID:           userID,
			UserName:     username,
			FirstName:    "Test",
			LastName:     "User",
			LanguageCode: "en",
		},
		Chat: &tgbotapi.Chat{
			ID:   userID,
			Type: "private",
		},
		Date: int(time.Now().Unix()),
		Text: text,
	}
}

// CreateTestCommand создает тестовую команду
func CreateTestCommand(command string, userID int64, username string) *tgbotapi.Message {
	msg := CreateTestMessage("/"+command, userID, username)
	// Дополнительно устанавливаем поля для команды
	entities := []tgbotapi.MessageEntity{
		{
			Type:   "bot_command",
			Offset: 0,
			Length: len(command) + 1,
		},
	}
	msg.Entities = entities
	return msg
}

// CreateTestCallback создает тестовый callback query
func CreateTestCallback(data string, userID int64) *tgbotapi.CallbackQuery {
	return &tgbotapi.CallbackQuery{
		ID: "test_callback_" + data,
		From: &tgbotapi.User{
			ID:           userID,
			UserName:     "testuser",
			FirstName:    "Test",
			LastName:     "User",
			LanguageCode: "en",
		},
		Message: &tgbotapi.Message{
			MessageID: int(time.Now().Unix()),
			Chat: &tgbotapi.Chat{
				ID:   userID,
				Type: "private",
			},
		},
		Data: data,
	}
}

// CreateUpdateWithMessage создает Update с сообщением
func CreateUpdateWithMessage(message *tgbotapi.Message) tgbotapi.Update {
	return tgbotapi.Update{
		UpdateID: int(time.Now().Unix()),
		Message:  message,
	}
}

// CreateUpdateWithCallback создает Update с callback query
func CreateUpdateWithCallback(callback *tgbotapi.CallbackQuery) tgbotapi.Update {
	return tgbotapi.Update{
		UpdateID:      int(time.Now().Unix()),
		CallbackQuery: callback,
	}
}

// CreateTestUser создает тестового пользователя
func CreateTestUser() *models.User {
	return &models.User{
		ID:                     1,
		TelegramID:             555666777,
		Username:               "testuser",
		FirstName:              "Test",
		NativeLanguageCode:     "",
		TargetLanguageCode:     "",
		InterfaceLanguageCode:  "en",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		State:                  "new",
		ProfileCompletionLevel: 0,
		Status:                 "new",
	}
}

// CreateTestAdminUser создает тестового администратора
func CreateTestAdminUser() *models.User {
	return &models.User{
		ID:                     2,
		TelegramID:             123456789,
		Username:               "testadmin",
		FirstName:              "Admin",
		NativeLanguageCode:     "en",
		TargetLanguageCode:     "ru",
		InterfaceLanguageCode:  "en",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		State:                  "ready",
		ProfileCompletionLevel: 100,
		Status:                 "active",
	}
}

// CreateTestLocalizer создает простой локализатор для тестов
func CreateTestLocalizer() LocalizerInterface {
	// Возвращаем мок локализатора вместо реального
	return mocks.NewLocalizerMock()
}

// LocalizerInterface интерфейс для локализатора (чтобы можно было использовать мок)
type LocalizerInterface interface {
	Get(langCode, key string) string
	GetWithParams(langCode, key string, params map[string]string) string
	GetLanguageName(langCode, interfaceLangCode string) string
	GetInterests(langCode string) (map[int]string, error)
}

// GetTestDatabaseURL возвращает URL тестовой базы данных
func GetTestDatabaseURL() string {
	// Проверяем переменную окружения для тестовой БД
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL != "" {
		return testDBURL
	}

	// Fallback на основную БД с тестовой схемой
	mainDBURL := os.Getenv("DATABASE_URL")
	if mainDBURL != "" {
		return mainDBURL
	}

	// Дефолтная тестовая БД
	return "postgres://postgres:password@localhost:5432/language_exchange_test?sslmode=disable"
}

// CleanupTestData очищает тестовые данные из базы данных
func CleanupTestData(db *sql.DB) {
	// Удаляем тестовые данные в правильном порядке (с учетом foreign keys)
	tables := []string{
		"user_feedback",
		"user_interests",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table + " WHERE telegram_id >= 12345 AND telegram_id <= 99999")
		if err != nil {
			// Игнорируем ошибки очистки в тестах
			continue
		}
	}
}

// GetTestRegularUserID возвращает тестового обычного пользователя
func GetTestRegularUserID() (int64, string) {
	return 12345, "testuser"
}

// GetTestAdminUserID возвращает тестового администратора
func GetTestAdminUserID() (int64, string) {
	return 123456, "testadmin"
}
