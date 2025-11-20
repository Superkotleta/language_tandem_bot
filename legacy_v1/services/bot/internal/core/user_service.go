package core

import (
	"fmt"
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
	user, err := s.db.FindOrCreateUser(telegramID, username, firstName)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	return user, nil
}

// UpdateUserState обновляет состояние пользователя.
func (s *UserService) UpdateUserState(userID int, state string) error {
	err := s.db.UpdateUserState(userID, state)
	if err != nil {
		return fmt.Errorf("failed to update user state: %w", err)
	}

	return nil
}

// UpdateUserStatus обновляет статус пользователя.
func (s *UserService) UpdateUserStatus(userID int, status string) error {
	err := s.db.UpdateUserStatus(userID, status)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

// SaveLanguages сохраняет выбранные пользователем языки.
func (s *UserService) SaveLanguages(userID int, nativeLang, targetLang string) error {
	err := s.db.SaveNativeLanguage(userID, nativeLang)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	if err := s.db.SaveTargetLanguage(userID, targetLang); err != nil {
		return fmt.Errorf("failed to save target language: %w", err)
	}

	return nil
}

// SaveInterest сохраняет интерес пользователя.
func (s *UserService) SaveInterest(userID, interestID int, isMain bool) error {
	err := s.db.SaveUserInterest(userID, interestID, isMain)
	if err != nil {
		return fmt.Errorf("failed to save user interest: %w", err)
	}

	return nil
}
