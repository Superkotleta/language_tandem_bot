package core

import (
	"language-exchange-bot/internal/database"
	"language-exchange-bot/internal/models"
)

// UserService обрабатывает операции с пользователями
type UserService struct {
	db *database.DB
}

func NewUserService(db *database.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (s *UserService) FindOrCreateUser(telegramID int64, username, firstName string) (*models.User, error) {
	return s.db.FindOrCreateUser(telegramID, username, firstName)
}

func (s *UserService) UpdateUserState(userID int, state string) error {
	return s.db.UpdateUserState(userID, state)
}

func (s *UserService) UpdateUserStatus(userID int, status string) error {
	return s.db.UpdateUserStatus(userID, status)
}

func (s *UserService) SaveLanguages(userID int, nativeLang, targetLang string) error {
	if err := s.db.SaveNativeLanguage(userID, nativeLang); err != nil {
		return err
	}
	return s.db.SaveTargetLanguage(userID, targetLang)
}

func (s *UserService) SaveInterest(userID int, interestID int, isMain bool) error {
	return s.db.SaveUserInterest(userID, interestID, isMain)
}
