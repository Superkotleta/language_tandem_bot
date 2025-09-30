package database

import (
	"database/sql"

	"language-exchange-bot/internal/models"
)

// Database интерфейс для работы с базой данных
type Database interface {
	// Пользователи
	FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error)
	GetUserByTelegramID(telegramID int64) (*models.User, error)
	UpdateUser(user *models.User) error
	UpdateUserInterfaceLanguage(userID int, language string) error
	UpdateUserState(userID int, state string) error
	UpdateUserStatus(userID int, status string) error
	UpdateUserNativeLanguage(userID int, langCode string) error
	UpdateUserTargetLanguage(userID int, langCode string) error
	UpdateUserTargetLanguageLevel(userID int, level string) error
	ResetUserProfile(userID int) error

	// Языки
	GetLanguages() ([]*models.Language, error)
	GetLanguageByCode(code string) (*models.Language, error)

	// Интересы
	GetInterests() ([]*models.Interest, error)
	GetUserSelectedInterests(userID int) ([]int, error)
	SaveUserInterests(userID int64, interestIDs []int) error
	SaveUserInterest(userID, interestID int, isPrimary bool) error
	RemoveUserInterest(userID, interestID int) error
	ClearUserInterests(userID int) error

	// Обратная связь
	SaveUserFeedback(userID int, feedbackText string, contactInfo *string) error
	GetUnprocessedFeedback() ([]map[string]interface{}, error)
	MarkFeedbackProcessed(feedbackID int, adminResponse string) error

	// Соединение
	GetConnection() *sql.DB
	Close() error
}
