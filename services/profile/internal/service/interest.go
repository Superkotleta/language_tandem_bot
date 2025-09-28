package service

import (
	"context"
	"fmt"

	"profile/internal/models"
	"profile/internal/repository"
)

type InterestService struct {
	interestRepo     *repository.InterestRepository
	userInterestRepo *repository.UserInterestRepository
}

func NewInterestService(
	interestRepo *repository.InterestRepository,
	userInterestRepo *repository.UserInterestRepository,
) *InterestService {
	return &InterestService{
		interestRepo:     interestRepo,
		userInterestRepo: userInterestRepo,
	}
}

// CreateInterest creates a new interest.
func (s *InterestService) CreateInterest(ctx context.Context, name, category string, description *string) (*models.Interest, error) {
	// Check if interest already exists
	exists, err := s.interestRepo.ExistsByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check interest existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("interest with name %s already exists", name)
	}

	// Validate category
	if !models.IsValidInterestCategory(category) {
		return nil, fmt.Errorf("invalid interest category: %s", category)
	}

	// Create interest
	interest, err := s.interestRepo.Create(ctx, name, category, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create interest: %w", err)
	}

	return interest, nil
}

// GetInterest retrieves an interest by ID.
func (s *InterestService) GetInterest(ctx context.Context, id int64) (*models.Interest, error) {
	interest, err := s.interestRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get interest: %w", err)
	}

	return interest, nil
}

// GetInterestByName retrieves an interest by name.
func (s *InterestService) GetInterestByName(ctx context.Context, name string) (*models.Interest, error) {
	interest, err := s.interestRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get interest by name: %w", err)
	}

	return interest, nil
}

// ListInterests retrieves a list of interests with pagination.
func (s *InterestService) ListInterests(ctx context.Context, req *models.InterestSearchRequest) (*models.InterestListResponse, error) {
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

	// List interests
	response, err := s.interestRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list interests: %w", err)
	}

	return response, nil
}

// UpdateInterest updates an interest.
func (s *InterestService) UpdateInterest(ctx context.Context, id int64, name, category string, description *string, isActive bool) (*models.Interest, error) {
	// Check if interest exists
	exists, err := s.interestRepo.Exists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check interest existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("interest not found")
	}

	// Validate category
	if !models.IsValidInterestCategory(category) {
		return nil, fmt.Errorf("invalid interest category: %s", category)
	}

	// Update interest
	interest, err := s.interestRepo.Update(ctx, id, name, category, description, isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to update interest: %w", err)
	}

	return interest, nil
}

// DeleteInterest deletes an interest.
func (s *InterestService) DeleteInterest(ctx context.Context, id int64) error {
	// Check if interest exists
	exists, err := s.interestRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check interest existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("interest not found")
	}

	// Delete interest
	err = s.interestRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete interest: %w", err)
	}

	return nil
}

// AddUserInterest adds an interest to a user.
func (s *InterestService) AddUserInterest(ctx context.Context, userID int64, req *models.CreateUserInterestRequest) (*models.UserInterestResponse, error) {
	// Check if interest exists
	exists, err := s.interestRepo.Exists(ctx, req.InterestID)
	if err != nil {
		return nil, fmt.Errorf("failed to check interest existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("interest not found")
	}

	// Check if user already has this interest
	exists, err = s.userInterestRepo.Exists(ctx, userID, req.InterestID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user interest existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user already has this interest")
	}

	// Add interest to user
	userInterest, err := s.userInterestRepo.Create(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to add interest to user: %w", err)
	}

	response := userInterest.ToResponse()
	return &response, nil
}

// RemoveUserInterest removes an interest from a user.
func (s *InterestService) RemoveUserInterest(ctx context.Context, userID, userInterestID int64) error {
	// Get user interest
	userInterest, err := s.userInterestRepo.GetByID(ctx, userInterestID)
	if err != nil {
		return fmt.Errorf("failed to get user interest: %w", err)
	}

	// Check if it belongs to the user
	if userInterest.UserID != userID {
		return fmt.Errorf("user interest not found")
	}

	// Remove user interest
	err = s.userInterestRepo.Delete(ctx, userInterestID)
	if err != nil {
		return fmt.Errorf("failed to remove user interest: %w", err)
	}

	return nil
}

// GetUserInterests retrieves all interests for a user.
func (s *InterestService) GetUserInterests(ctx context.Context, userID int64, page, perPage int) (*models.UserInterestListResponse, error) {
	// List user interests
	response, err := s.userInterestRepo.List(ctx, userID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get user interests: %w", err)
	}

	return response, nil
}
