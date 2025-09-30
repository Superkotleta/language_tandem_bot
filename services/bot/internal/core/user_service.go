package core

import (
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/models"
)

// UserService обрабатывает операции с пользователями.
type UserService struct {
	db *database.DB
}

// NewUserService создает новый экземпляр UserService.
func NewUserService(db *database.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// FindOrCreateUser находит или создает пользователя по Telegram ID.
func (s *UserService) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	return s.db.FindOrCreateUser(telegramID, username, firstName)
}

// UpdateUserState обновляет состояние пользователя.
func (s *UserService) UpdateUserState(userID int, state string) error {
	return s.db.UpdateUserState(userID, state)
}

// UpdateUserStatus обновляет статус пользователя.
func (s *UserService) UpdateUserStatus(userID int, status string) error {
	return s.db.UpdateUserStatus(userID, status)
}

// SaveLanguages сохраняет выбранные пользователем языки.
func (s *UserService) SaveLanguages(userID int, nativeLang, targetLang string) error {
	err := s.db.SaveNativeLanguage(userID, nativeLang)
	if err != nil {
		return err
	}

	return s.db.SaveTargetLanguage(userID, targetLang)
}

// SaveInterest сохраняет интерес пользователя.
func (s *UserService) SaveInterest(userID, interestID int, isMain bool) error {
	return s.db.SaveUserInterest(userID, interestID, isMain)
}
