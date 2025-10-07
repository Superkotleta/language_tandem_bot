package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"profile/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPreferenceRepository struct {
	db *pgxpool.Pool
}

func NewUserPreferenceRepository(db *pgxpool.Pool) *UserPreferenceRepository {
	return &UserPreferenceRepository{db: db}
}

// Create creates a new user preference.
func (r *UserPreferenceRepository) Create(ctx context.Context, req *models.CreateUserPreferenceRequest, userID int64) (*models.UserPreference, error) {
	query := `
		INSERT INTO user_preferences (
			user_id, min_age, max_age, preferred_gender, preferred_countries,
			preferred_languages, max_distance, timezone_offset, availability_start,
			availability_end, is_online_only
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, created_at, updated_at`

	var userPreference models.UserPreference
	err := r.db.QueryRow(ctx, query,
		userID, req.MinAge, req.MaxAge, req.PreferredGender, req.PreferredCountries,
		req.PreferredLanguages, req.MaxDistance, req.TimezoneOffset, req.AvailabilityStart,
		req.AvailabilityEnd, req.IsOnlineOnly,
	).Scan(&userPreference.ID, &userPreference.CreatedAt, &userPreference.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user preference: %w", err)
	}

	userPreference.UserID = userID
	userPreference.MinAge = req.MinAge
	userPreference.MaxAge = req.MaxAge
	userPreference.PreferredGender = req.PreferredGender
	userPreference.PreferredCountries = req.PreferredCountries
	userPreference.PreferredLanguages = req.PreferredLanguages
	userPreference.MaxDistance = req.MaxDistance
	userPreference.TimezoneOffset = req.TimezoneOffset
	userPreference.AvailabilityStart = req.AvailabilityStart
	userPreference.AvailabilityEnd = req.AvailabilityEnd
	userPreference.IsOnlineOnly = req.IsOnlineOnly

	return &userPreference, nil
}

// GetByUserID retrieves user preferences by user ID.
func (r *UserPreferenceRepository) GetByUserID(ctx context.Context, userID int64) (*models.UserPreference, error) {
	query := `
		SELECT id, user_id, min_age, max_age, preferred_gender, preferred_countries,
			   preferred_languages, max_distance, timezone_offset, availability_start,
			   availability_end, is_online_only, created_at, updated_at
		FROM user_preferences WHERE user_id = $1`

	var userPreference models.UserPreference
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&userPreference.ID, &userPreference.UserID, &userPreference.MinAge, &userPreference.MaxAge,
		&userPreference.PreferredGender, &userPreference.PreferredCountries, &userPreference.PreferredLanguages,
		&userPreference.MaxDistance, &userPreference.TimezoneOffset, &userPreference.AvailabilityStart,
		&userPreference.AvailabilityEnd, &userPreference.IsOnlineOnly, &userPreference.CreatedAt, &userPreference.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user preference not found")
		}
		return nil, fmt.Errorf("failed to get user preference: %w", err)
	}

	return &userPreference, nil
}

// Update updates user preferences.
func (r *UserPreferenceRepository) Update(ctx context.Context, userID int64, req *models.UpdateUserPreferenceRequest) (*models.UserPreference, error) {
	// Build dynamic query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.MinAge != nil {
		setParts = append(setParts, fmt.Sprintf("min_age = $%d", argIndex))
		args = append(args, *req.MinAge)
		argIndex++
	}
	if req.MaxAge != nil {
		setParts = append(setParts, fmt.Sprintf("max_age = $%d", argIndex))
		args = append(args, *req.MaxAge)
		argIndex++
	}
	if req.PreferredGender != nil {
		setParts = append(setParts, fmt.Sprintf("preferred_gender = $%d", argIndex))
		args = append(args, *req.PreferredGender)
		argIndex++
	}
	if req.PreferredCountries != nil {
		setParts = append(setParts, fmt.Sprintf("preferred_countries = $%d", argIndex))
		args = append(args, req.PreferredCountries)
		argIndex++
	}
	if req.PreferredLanguages != nil {
		setParts = append(setParts, fmt.Sprintf("preferred_languages = $%d", argIndex))
		args = append(args, req.PreferredLanguages)
		argIndex++
	}
	if req.MaxDistance != nil {
		setParts = append(setParts, fmt.Sprintf("max_distance = $%d", argIndex))
		args = append(args, *req.MaxDistance)
		argIndex++
	}
	if req.TimezoneOffset != nil {
		setParts = append(setParts, fmt.Sprintf("timezone_offset = $%d", argIndex))
		args = append(args, *req.TimezoneOffset)
		argIndex++
	}
	if req.AvailabilityStart != nil {
		setParts = append(setParts, fmt.Sprintf("availability_start = $%d", argIndex))
		args = append(args, *req.AvailabilityStart)
		argIndex++
	}
	if req.AvailabilityEnd != nil {
		setParts = append(setParts, fmt.Sprintf("availability_end = $%d", argIndex))
		args = append(args, *req.AvailabilityEnd)
		argIndex++
	}
	if req.IsOnlineOnly != nil {
		setParts = append(setParts, fmt.Sprintf("is_online_only = $%d", argIndex))
		args = append(args, *req.IsOnlineOnly)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByUserID(ctx, userID)
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE user_preferences SET %s
		WHERE user_id = $%d
		RETURNING id, user_id, min_age, max_age, preferred_gender, preferred_countries,
				  preferred_languages, max_distance, timezone_offset, availability_start,
				  availability_end, is_online_only, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	var userPreference models.UserPreference
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&userPreference.ID, &userPreference.UserID, &userPreference.MinAge, &userPreference.MaxAge,
		&userPreference.PreferredGender, &userPreference.PreferredCountries, &userPreference.PreferredLanguages,
		&userPreference.MaxDistance, &userPreference.TimezoneOffset, &userPreference.AvailabilityStart,
		&userPreference.AvailabilityEnd, &userPreference.IsOnlineOnly, &userPreference.CreatedAt, &userPreference.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user preference not found")
		}
		return nil, fmt.Errorf("failed to update user preference: %w", err)
	}

	return &userPreference, nil
}

// Delete deletes user preferences.
func (r *UserPreferenceRepository) Delete(ctx context.Context, userID int64) error {
	query := `DELETE FROM user_preferences WHERE user_id = $1`
	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user preference: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user preference not found")
	}

	return nil
}

// Exists checks if user preferences exist.
func (r *UserPreferenceRepository) Exists(ctx context.Context, userID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_preferences WHERE user_id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user preference existence: %w", err)
	}
	return exists, nil
}
