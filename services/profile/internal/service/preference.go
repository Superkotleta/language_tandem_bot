package service

import (
	"context"
	"fmt"

	"profile/internal/models"
	"profile/internal/repository"
)

type PreferenceService struct {
	preferenceRepo *repository.UserPreferenceRepository
	userRepo       *repository.UserRepository
}

func NewPreferenceService(
	preferenceRepo *repository.UserPreferenceRepository,
	userRepo *repository.UserRepository,
) *PreferenceService {
	return &PreferenceService{
		preferenceRepo: preferenceRepo,
		userRepo:       userRepo,
	}
}

// CreateUserPreferences creates user preferences.
func (s *PreferenceService) CreateUserPreferences(ctx context.Context, userID int64, req *models.CreateUserPreferenceRequest) (*models.UserPreferenceResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Check if preferences already exist
	exists, err = s.preferenceRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check preferences existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user preferences already exist")
	}

	// Validate age range
	if req.MinAge != nil && req.MaxAge != nil && *req.MinAge > *req.MaxAge {
		return nil, fmt.Errorf("min_age cannot be greater than max_age")
	}

	// Validate gender preference
	if req.PreferredGender != nil && !models.IsValidGenderPreference(*req.PreferredGender) {
		return nil, fmt.Errorf("invalid preferred gender: %s", *req.PreferredGender)
	}

	// Create preferences
	preferences, err := s.preferenceRepo.Create(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user preferences: %w", err)
	}

	response := preferences.ToResponse()
	return &response, nil
}

// GetUserPreferences retrieves user preferences.
func (s *PreferenceService) GetUserPreferences(ctx context.Context, userID int64) (*models.UserPreferenceResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get preferences
	preferences, err := s.preferenceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	response := preferences.ToResponse()
	return &response, nil
}

// UpdateUserPreferences updates user preferences.
func (s *PreferenceService) UpdateUserPreferences(ctx context.Context, userID int64, req *models.UpdateUserPreferenceRequest) (*models.UserPreferenceResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Check if preferences exist
	exists, err = s.preferenceRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check preferences existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user preferences not found")
	}

	// Validate age range
	if req.MinAge != nil && req.MaxAge != nil && *req.MinAge > *req.MaxAge {
		return nil, fmt.Errorf("min_age cannot be greater than max_age")
	}

	// Validate gender preference
	if req.PreferredGender != nil && !models.IsValidGenderPreference(*req.PreferredGender) {
		return nil, fmt.Errorf("invalid preferred gender: %s", *req.PreferredGender)
	}

	// Update preferences
	preferences, err := s.preferenceRepo.Update(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user preferences: %w", err)
	}

	response := preferences.ToResponse()
	return &response, nil
}

// DeleteUserPreferences deletes user preferences.
func (s *PreferenceService) DeleteUserPreferences(ctx context.Context, userID int64) error {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Check if preferences exist
	exists, err = s.preferenceRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check preferences existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user preferences not found")
	}

	// Delete preferences
	err = s.preferenceRepo.Delete(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user preferences: %w", err)
	}

	return nil
}
