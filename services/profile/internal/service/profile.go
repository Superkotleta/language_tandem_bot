package service

import (
	"context"
	"fmt"

	"profile/internal/models"
	"profile/internal/repository"
)

type ProfileService struct {
	userRepo         *repository.UserRepository
	languageRepo     *repository.LanguageRepository
	interestRepo     *repository.InterestRepository
	userLanguageRepo *repository.UserLanguageRepository
	userInterestRepo *repository.UserInterestRepository
	preferenceRepo   *repository.UserPreferenceRepository
	traitRepo        *repository.UserTraitRepository
	availabilityRepo *repository.UserTimeAvailabilityRepository
}

func NewProfileService(
	userRepo *repository.UserRepository,
	languageRepo *repository.LanguageRepository,
	interestRepo *repository.InterestRepository,
	userLanguageRepo *repository.UserLanguageRepository,
	userInterestRepo *repository.UserInterestRepository,
	preferenceRepo *repository.UserPreferenceRepository,
	traitRepo *repository.UserTraitRepository,
	availabilityRepo *repository.UserTimeAvailabilityRepository,
) *ProfileService {
	return &ProfileService{
		userRepo:         userRepo,
		languageRepo:     languageRepo,
		interestRepo:     interestRepo,
		userLanguageRepo: userLanguageRepo,
		userInterestRepo: userInterestRepo,
		preferenceRepo:   preferenceRepo,
		traitRepo:        traitRepo,
		availabilityRepo: availabilityRepo,
	}
}

// CreateUser creates a new user.
func (s *ProfileService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error) {
	// Validate that at least one ID is provided
	if req.TelegramID == nil && req.DiscordID == nil {
		return nil, fmt.Errorf("either telegram_id or discord_id must be provided")
	}

	// Check if user already exists
	if req.TelegramID != nil {
		existingUser, err := s.userRepo.GetByTelegramID(ctx, *req.TelegramID)
		if err == nil && existingUser != nil {
			return nil, fmt.Errorf("user with telegram_id %d already exists", *req.TelegramID)
		}
	}

	if req.DiscordID != nil {
		existingUser, err := s.userRepo.GetByDiscordID(ctx, *req.DiscordID)
		if err == nil && existingUser != nil {
			return nil, fmt.Errorf("user with discord_id %d already exists", *req.DiscordID)
		}
	}

	// Create user
	user, err := s.userRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// GetUser retrieves a user by ID.
func (s *ProfileService) GetUser(ctx context.Context, id int64) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// GetUserByTelegramID retrieves a user by Telegram ID.
func (s *ProfileService) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// GetUserByDiscordID retrieves a user by Discord ID.
func (s *ProfileService) GetUserByDiscordID(ctx context.Context, discordID int64) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByDiscordID(ctx, discordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by discord ID: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// UpdateUser updates a user.
func (s *ProfileService) UpdateUser(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Update user
	user, err := s.userRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// DeleteUser deletes a user.
func (s *ProfileService) DeleteUser(ctx context.Context, id int64) error {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Delete user (cascade will handle related records)
	err = s.userRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves a list of users with pagination.
func (s *ProfileService) ListUsers(ctx context.Context, req *models.UserSearchRequest) (*models.UserListResponse, error) {
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

	// List users
	response, err := s.userRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return response, nil
}

// UpdateLastSeen updates the last seen timestamp for a user.
func (s *ProfileService) UpdateLastSeen(ctx context.Context, id int64) error {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Update last seen
	err = s.userRepo.UpdateLastSeen(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	return nil
}

// GetUserProfileCompletion calculates the profile completion score for a user.
func (s *ProfileService) GetUserProfileCompletion(ctx context.Context, id int64) (float64, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return 0, fmt.Errorf("user not found")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("failed to get user: %w", err)
	}

	// Calculate completion score
	score := 0.0
	totalFields := 0.0

	// Basic profile fields (40% of total score)
	basicFields := []interface{}{
		user.Username, user.FirstName, user.LastName, user.Email, user.Phone,
		user.Bio, user.Age, user.Gender, user.Country, user.City, user.Timezone, user.ProfilePictureURL,
	}
	for _, field := range basicFields {
		totalFields++
		if field != nil {
			switch v := field.(type) {
			case *string:
				if v != nil && *v != "" {
					score++
				}
			case *int:
				if v != nil && *v > 0 {
					score++
				}
			}
		}
	}

	// Languages (30% of total score)
	languages, err := s.GetUserLanguages(ctx, id, 1, 100)
	if err == nil && languages != nil && len(languages.Languages) > 0 {
		score += 3.0 // At least one language
		if len(languages.Languages) >= 2 {
			score += 1.0 // Multiple languages
		}
	}
	totalFields += 4.0

	// Interests (20% of total score)
	interests, err := s.GetUserInterests(ctx, id, 1, 100)
	if err == nil && interests != nil && len(interests.Interests) > 0 {
		score += 2.0 // At least one interest
		if len(interests.Interests) >= 3 {
			score += 1.0 // Multiple interests
		}
	}
	totalFields += 3.0

	// Preferences (10% of total score)
	preferences, err := s.GetUserPreferences(ctx, id)
	if err == nil && preferences != nil {
		score += 1.0 // Has preferences
	}
	totalFields += 1.0

	// Calculate percentage
	if totalFields > 0 {
		return (score / totalFields) * 100, nil
	}

	return 0, nil
}

// GetUserLanguages retrieves all languages for a user.
func (s *ProfileService) GetUserLanguages(ctx context.Context, userID int64, page, perPage int) (*models.UserLanguageListResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get user languages
	response, err := s.userLanguageRepo.List(ctx, userID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get user languages: %w", err)
	}

	return response, nil
}

// GetUserInterests retrieves all interests for a user.
func (s *ProfileService) GetUserInterests(ctx context.Context, userID int64, page, perPage int) (*models.UserInterestListResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get user interests
	response, err := s.userInterestRepo.List(ctx, userID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get user interests: %w", err)
	}

	return response, nil
}

// GetUserPreferences retrieves user preferences.
func (s *ProfileService) GetUserPreferences(ctx context.Context, userID int64) (*models.UserPreferenceResponse, error) {
	// Check if user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Get user preferences
	preferences, err := s.preferenceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	response := preferences.ToResponse()
	return &response, nil
}
