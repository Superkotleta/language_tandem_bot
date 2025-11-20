package service

import (
	"context"
	"language-exchange-bot/internal/domain"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	GetBySocialID(ctx context.Context, socialID, platform string) (*domain.User, error)
	CreateOrUpdate(ctx context.Context, user *domain.User) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, socialID, platform, firstName, username, languageCode string) (*domain.User, error) {
	// Normalizing language code if needed
	if languageCode == "" {
		languageCode = "en"
	}

	user := &domain.User{
		SocialID:  socialID,
		Platform:  platform,
		FirstName: firstName,
		Username:  username,
		Language:  languageCode,
	}

	if err := s.repo.CreateOrUpdate(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}


