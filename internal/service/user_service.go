package service

import (
	"context"
	"language-exchange-bot/internal/domain"
	"strings"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	GetBySocialID(ctx context.Context, socialID, platform string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	AddInterest(ctx context.Context, userID, interestID string) error
	RemoveInterest(ctx context.Context, userID, interestID string) error
	GetUserInterests(ctx context.Context, userID string) ([]domain.Interest, error)
}

// ReferenceRepository defines the interface for static data persistence
type ReferenceRepository interface {
	GetLanguages(ctx context.Context) ([]domain.Language, error)
	GetCategories(ctx context.Context) ([]domain.InterestCategory, error)
	GetInterestsByCategory(ctx context.Context, categoryID string) ([]domain.Interest, error)
}

type UserService struct {
	repo    UserRepository
	refRepo ReferenceRepository
}

func NewUserService(repo UserRepository, refRepo ReferenceRepository) *UserService {
	return &UserService{
		repo:    repo,
		refRepo: refRepo,
	}
}

// GetUserBySocialID retrieves a user by their social ID
func (s *UserService) GetUserBySocialID(ctx context.Context, socialID, platform string) (*domain.User, error) {
	return s.repo.GetBySocialID(ctx, socialID, platform)
}

// RegisterUser initializes a new user profile
func (s *UserService) RegisterUser(ctx context.Context, socialID, platform, firstName, username, languageCode string) (*domain.User, error) {
	if languageCode == "" {
		languageCode = "en"
	}

	// Normalize language code (e.g. "es-ES" -> "es")
	if len(languageCode) > 2 {
		languageCode = strings.Split(languageCode, "-")[0]
	}

	user := &domain.User{
		SocialID:      socialID,
		Platform:      platform,
		FirstName:     firstName,
		Username:      username,
		InterfaceLang: languageCode,
		Status:        domain.StatusFillingProfile,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.repo.Update(ctx, user)
}

// Interest Management
func (s *UserService) AddInterest(ctx context.Context, userID, interestID string) error {
	return s.repo.AddInterest(ctx, userID, interestID)
}

func (s *UserService) RemoveInterest(ctx context.Context, userID, interestID string) error {
	return s.repo.RemoveInterest(ctx, userID, interestID)
}

func (s *UserService) GetUserInterests(ctx context.Context, userID string) ([]domain.Interest, error) {
	return s.repo.GetUserInterests(ctx, userID)
}

// Reference Data
func (s *UserService) GetLanguages(ctx context.Context) ([]domain.Language, error) {
	return s.refRepo.GetLanguages(ctx)
}

func (s *UserService) GetCategories(ctx context.Context) ([]domain.InterestCategory, error) {
	return s.refRepo.GetCategories(ctx)
}

func (s *UserService) GetInterestsByCategory(ctx context.Context, categoryID string) ([]domain.Interest, error) {
	return s.refRepo.GetInterestsByCategory(ctx, categoryID)
}
