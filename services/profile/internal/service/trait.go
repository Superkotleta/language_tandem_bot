package service

import (
	"context"
	"fmt"

	"profile/internal/models"
	"profile/internal/repository"
)

type TraitService struct {
	traitRepo *repository.UserTraitRepository
	userRepo  *repository.UserRepository
}

func NewTraitService(
	traitRepo *repository.UserTraitRepository,
	userRepo *repository.UserRepository,
) *TraitService {
	return &TraitService{
		traitRepo: traitRepo,
		userRepo:  userRepo,
	}
}

// CreateUserTrait creates a new user trait.
func (s *TraitService) CreateUserTrait(ctx context.Context, userID int64, req *models.CreateUserTraitRequest) (*models.UserTraitResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Validate trait type
	if !models.IsValidTraitType(req.TraitType) {
		return nil, fmt.Errorf("invalid trait type: %s", req.TraitType)
	}

	// Check if trait already exists for this user
	exists, err = s.traitRepo.Exists(ctx, userID, req.TraitType)
	if err != nil {
		return nil, fmt.Errorf("failed to check trait existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user already has this trait type")
	}

	// Create trait
	trait, err := s.traitRepo.Create(ctx, req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user trait: %w", err)
	}

	response := trait.ToResponse()
	return &response, nil
}

// GetUserTrait retrieves a user trait by ID.
func (s *TraitService) GetUserTrait(ctx context.Context, userID, traitID int64) (*models.UserTraitResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get trait
	trait, err := s.traitRepo.GetByID(ctx, traitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user trait: %w", err)
	}

	// Check if it belongs to the user
	if trait.UserID != userID {
		return nil, fmt.Errorf("user trait not found")
	}

	response := trait.ToResponse()
	return &response, nil
}

// GetUserTraitByType retrieves a user trait by type.
func (s *TraitService) GetUserTraitByType(ctx context.Context, userID int64, traitType string) (*models.UserTraitResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get trait
	trait, err := s.traitRepo.GetByUserIDAndTraitType(ctx, userID, traitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get user trait: %w", err)
	}

	response := trait.ToResponse()
	return &response, nil
}

// ListUserTraits retrieves all traits for a user.
func (s *TraitService) ListUserTraits(ctx context.Context, userID int64, page, perPage int) (*models.UserTraitListResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// List traits
	response, err := s.traitRepo.List(ctx, userID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get user traits: %w", err)
	}

	return response, nil
}

// UpdateUserTrait updates a user trait.
func (s *TraitService) UpdateUserTrait(ctx context.Context, userID, traitID int64, req *models.UpdateUserTraitRequest) (*models.UserTraitResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get trait
	trait, err := s.traitRepo.GetByID(ctx, traitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user trait: %w", err)
	}

	// Check if it belongs to the user
	if trait.UserID != userID {
		return nil, fmt.Errorf("user trait not found")
	}

	// Update trait
	updatedTrait, err := s.traitRepo.Update(ctx, traitID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user trait: %w", err)
	}

	response := updatedTrait.ToResponse()
	return &response, nil
}

// DeleteUserTrait deletes a user trait.
func (s *TraitService) DeleteUserTrait(ctx context.Context, userID, traitID int64) error {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Get trait
	trait, err := s.traitRepo.GetByID(ctx, traitID)
	if err != nil {
		return fmt.Errorf("failed to get user trait: %w", err)
	}

	// Check if it belongs to the user
	if trait.UserID != userID {
		return fmt.Errorf("user trait not found")
	}

	// Delete trait
	err = s.traitRepo.Delete(ctx, traitID)
	if err != nil {
		return fmt.Errorf("failed to delete user trait: %w", err)
	}

	return nil
}

// DeleteUserTraitByType deletes a user trait by type.
func (s *TraitService) DeleteUserTraitByType(ctx context.Context, userID int64, traitType string) error {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Delete trait
	err = s.traitRepo.DeleteByUserIDAndTraitType(ctx, userID, traitType)
	if err != nil {
		return fmt.Errorf("failed to delete user trait: %w", err)
	}

	return nil
}
