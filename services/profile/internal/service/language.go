package service

import (
	"context"
	"fmt"

	"profile/internal/models"
	"profile/internal/repository"
)

type LanguageService struct {
	languageRepo     *repository.LanguageRepository
	userLanguageRepo *repository.UserLanguageRepository
}

func NewLanguageService(
	languageRepo *repository.LanguageRepository,
	userLanguageRepo *repository.UserLanguageRepository,
) *LanguageService {
	return &LanguageService{
		languageRepo:     languageRepo,
		userLanguageRepo: userLanguageRepo,
	}
}

// CreateLanguage creates a new language.
func (s *LanguageService) CreateLanguage(ctx context.Context, code, name, nativeName string) (*models.Language, error) {
	// Check if language already exists
	exists, err := s.languageRepo.ExistsByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to check language existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("language with code %s already exists", code)
	}

	// Create language
	language, err := s.languageRepo.Create(ctx, code, name, nativeName)
	if err != nil {
		return nil, fmt.Errorf("failed to create language: %w", err)
	}

	return language, nil
}

// GetLanguage retrieves a language by ID.
func (s *LanguageService) GetLanguage(ctx context.Context, id int64) (*models.Language, error) {
	language, err := s.languageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get language: %w", err)
	}

	return language, nil
}

// GetLanguageByCode retrieves a language by code.
func (s *LanguageService) GetLanguageByCode(ctx context.Context, code string) (*models.Language, error) {
	language, err := s.languageRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get language by code: %w", err)
	}

	return language, nil
}

// ListLanguages retrieves a list of languages with pagination.
func (s *LanguageService) ListLanguages(ctx context.Context, req *models.LanguageSearchRequest) (*models.LanguageListResponse, error) {
	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}

	// List languages
	response, err := s.languageRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list languages: %w", err)
	}

	return response, nil
}

// UpdateLanguage updates a language.
func (s *LanguageService) UpdateLanguage(ctx context.Context, id int64, name, nativeName string, isActive bool) (*models.Language, error) {
	// Check if language exists
	exists, err := s.languageRepo.Exists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check language existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("language not found")
	}

	// Update language
	language, err := s.languageRepo.Update(ctx, id, name, nativeName, isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to update language: %w", err)
	}

	return language, nil
}

// DeleteLanguage deletes a language.
func (s *LanguageService) DeleteLanguage(ctx context.Context, id int64) error {
	// Check if language exists
	exists, err := s.languageRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check language existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("language not found")
	}

	// Delete language
	err = s.languageRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete language: %w", err)
	}

	return nil
}

// AddUserLanguage adds a language to a user.
func (s *LanguageService) AddUserLanguage(ctx context.Context, userID int64, req *models.CreateUserLanguageRequest) (*models.UserLanguageResponse, error) {
	// Check if language exists
	exists, err := s.languageRepo.Exists(ctx, req.LanguageID)
	if err != nil {
		return nil, fmt.Errorf("failed to check language existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("language not found")
	}

	// Check if user already has this language
	exists, err = s.userLanguageRepo.Exists(ctx, userID, req.LanguageID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user language existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user already has this language")
	}

	// Add language to user
	userLanguage, err := s.userLanguageRepo.Create(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to add language to user: %w", err)
	}

	response := userLanguage.ToResponse()
	return &response, nil
}

// UpdateUserLanguage updates a user's language.
func (s *LanguageService) UpdateUserLanguage(ctx context.Context, userID, userLanguageID int64, req *models.UpdateUserLanguageRequest) (*models.UserLanguageResponse, error) {
	// Get user language
	userLanguage, err := s.userLanguageRepo.GetByID(ctx, userLanguageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user language: %w", err)
	}

	// Check if it belongs to the user
	if userLanguage.UserID != userID {
		return nil, fmt.Errorf("user language not found")
	}

	// Update user language
	updatedUserLanguage, err := s.userLanguageRepo.Update(ctx, userLanguageID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user language: %w", err)
	}

	response := updatedUserLanguage.ToResponse()
	return &response, nil
}

// RemoveUserLanguage removes a language from a user.
func (s *LanguageService) RemoveUserLanguage(ctx context.Context, userID, userLanguageID int64) error {
	// Get user language
	userLanguage, err := s.userLanguageRepo.GetByID(ctx, userLanguageID)
	if err != nil {
		return fmt.Errorf("failed to get user language: %w", err)
	}

	// Check if it belongs to the user
	if userLanguage.UserID != userID {
		return fmt.Errorf("user language not found")
	}

	// Remove user language
	err = s.userLanguageRepo.Delete(ctx, userLanguageID)
	if err != nil {
		return fmt.Errorf("failed to remove user language: %w", err)
	}

	return nil
}

// GetUserLanguages retrieves all languages for a user.
func (s *LanguageService) GetUserLanguages(ctx context.Context, userID int64, page, perPage int) (*models.UserLanguageListResponse, error) {
	// List user languages
	response, err := s.userLanguageRepo.List(ctx, userID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get user languages: %w", err)
	}

	return response, nil
}
