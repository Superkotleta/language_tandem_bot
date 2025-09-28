package service

import (
	"context"
	"fmt"

	"profile/internal/models"
	"profile/internal/repository"
)

type AvailabilityService struct {
	availabilityRepo *repository.UserTimeAvailabilityRepository
	userRepo         *repository.UserRepository
}

func NewAvailabilityService(
	availabilityRepo *repository.UserTimeAvailabilityRepository,
	userRepo *repository.UserRepository,
) *AvailabilityService {
	return &AvailabilityService{
		availabilityRepo: availabilityRepo,
		userRepo:         userRepo,
	}
}

// CreateUserTimeAvailability creates a new user time availability.
func (s *AvailabilityService) CreateUserTimeAvailability(ctx context.Context, userID int64, req *models.CreateUserTimeAvailabilityRequest) (*models.UserTimeAvailabilityResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Validate day of week
	if !models.IsValidDayOfWeek(req.DayOfWeek) {
		return nil, fmt.Errorf("invalid day of week: %d", req.DayOfWeek)
	}

	// Validate time format (basic validation)
	if req.StartTime == "" || req.EndTime == "" {
		return nil, fmt.Errorf("start_time and end_time are required")
	}

	// Create time availability
	availability, err := s.availabilityRepo.Create(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user time availability: %w", err)
	}

	response := availability.ToResponse()
	return &response, nil
}

// GetUserTimeAvailability retrieves a user time availability by ID.
func (s *AvailabilityService) GetUserTimeAvailability(ctx context.Context, userID, availabilityID int64) (*models.UserTimeAvailabilityResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get time availability
	availability, err := s.availabilityRepo.GetByID(ctx, availabilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user time availability: %w", err)
	}

	// Check if it belongs to the user
	if availability.UserID != userID {
		return nil, fmt.Errorf("user time availability not found")
	}

	response := availability.ToResponse()
	return &response, nil
}

// GetUserTimeAvailabilityByDay retrieves time availability for a specific day.
func (s *AvailabilityService) GetUserTimeAvailabilityByDay(ctx context.Context, userID int64, dayOfWeek int) ([]models.UserTimeAvailabilityResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Validate day of week
	if !models.IsValidDayOfWeek(dayOfWeek) {
		return nil, fmt.Errorf("invalid day of week: %d", dayOfWeek)
	}

	// Get time availability
	availabilities, err := s.availabilityRepo.GetByUserIDAndDayOfWeek(ctx, userID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get user time availability: %w", err)
	}

	// Convert to responses
	responses := make([]models.UserTimeAvailabilityResponse, len(availabilities))
	for i, availability := range availabilities {
		responses[i] = availability.ToResponse()
	}

	return responses, nil
}

// ListUserTimeAvailability retrieves all time availability for a user.
func (s *AvailabilityService) ListUserTimeAvailability(ctx context.Context, userID int64, page, perPage int) (*models.UserTimeAvailabilityListResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// List time availability
	response, err := s.availabilityRepo.List(ctx, userID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get user time availability: %w", err)
	}

	return response, nil
}

// UpdateUserTimeAvailability updates a user time availability.
func (s *AvailabilityService) UpdateUserTimeAvailability(ctx context.Context, userID, availabilityID int64, req *models.UpdateUserTimeAvailabilityRequest) (*models.UserTimeAvailabilityResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get time availability
	availability, err := s.availabilityRepo.GetByID(ctx, availabilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user time availability: %w", err)
	}

	// Check if it belongs to the user
	if availability.UserID != userID {
		return nil, fmt.Errorf("user time availability not found")
	}

	// Validate time format (basic validation)
	if req.StartTime == "" || req.EndTime == "" {
		return nil, fmt.Errorf("start_time and end_time are required")
	}

	// Update time availability
	updatedAvailability, err := s.availabilityRepo.Update(ctx, availabilityID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user time availability: %w", err)
	}

	response := updatedAvailability.ToResponse()
	return &response, nil
}

// DeleteUserTimeAvailability deletes a user time availability.
func (s *AvailabilityService) DeleteUserTimeAvailability(ctx context.Context, userID, availabilityID int64) error {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Get time availability
	availability, err := s.availabilityRepo.GetByID(ctx, availabilityID)
	if err != nil {
		return fmt.Errorf("failed to get user time availability: %w", err)
	}

	// Check if it belongs to the user
	if availability.UserID != userID {
		return fmt.Errorf("user time availability not found")
	}

	// Delete time availability
	err = s.availabilityRepo.Delete(ctx, availabilityID)
	if err != nil {
		return fmt.Errorf("failed to delete user time availability: %w", err)
	}

	return nil
}

// DeleteAllUserTimeAvailability deletes all time availability for a user.
func (s *AvailabilityService) DeleteAllUserTimeAvailability(ctx context.Context, userID int64) error {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Delete all time availability
	err = s.availabilityRepo.DeleteByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user time availability: %w", err)
	}

	return nil
}
