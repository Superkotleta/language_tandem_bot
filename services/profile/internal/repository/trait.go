package repository

import (
	"context"
	"errors"
	"fmt"

	"profile/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserTraitRepository struct {
	db *pgxpool.Pool
}

func NewUserTraitRepository(db *pgxpool.Pool) *UserTraitRepository {
	return &UserTraitRepository{db: db}
}

// Create creates a new user trait.
func (r *UserTraitRepository) Create(ctx context.Context, req *models.CreateUserTraitRequest, userID int64) (*models.UserTrait, error) {
	query := `
		INSERT INTO user_traits (user_id, trait_type, trait_value)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	var userTrait models.UserTrait
	err := r.db.QueryRow(ctx, query, userID, req.TraitType, req.TraitValue).Scan(
		&userTrait.ID, &userTrait.CreatedAt, &userTrait.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user trait: %w", err)
	}

	userTrait.UserID = userID
	userTrait.TraitType = req.TraitType
	userTrait.TraitValue = req.TraitValue

	return &userTrait, nil
}

// GetByID retrieves a user trait by ID.
func (r *UserTraitRepository) GetByID(ctx context.Context, id int64) (*models.UserTrait, error) {
	query := `
		SELECT id, user_id, trait_type, trait_value, created_at, updated_at
		FROM user_traits WHERE id = $1`

	var userTrait models.UserTrait
	err := r.db.QueryRow(ctx, query, id).Scan(
		&userTrait.ID, &userTrait.UserID, &userTrait.TraitType, &userTrait.TraitValue, &userTrait.CreatedAt, &userTrait.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user trait not found")
		}
		return nil, fmt.Errorf("failed to get user trait: %w", err)
	}

	return &userTrait, nil
}

// GetByUserID retrieves all traits for a user.
func (r *UserTraitRepository) GetByUserID(ctx context.Context, userID int64) ([]models.UserTrait, error) {
	query := `
		SELECT id, user_id, trait_type, trait_value, created_at, updated_at
		FROM user_traits WHERE user_id = $1
		ORDER BY trait_type ASC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user traits: %w", err)
	}
	defer rows.Close()

	var userTraits []models.UserTrait
	for rows.Next() {
		var userTrait models.UserTrait
		err := rows.Scan(
			&userTrait.ID, &userTrait.UserID, &userTrait.TraitType, &userTrait.TraitValue, &userTrait.CreatedAt, &userTrait.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user trait: %w", err)
		}
		userTraits = append(userTraits, userTrait)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user traits: %w", err)
	}

	return userTraits, nil
}

// GetByUserIDAndTraitType retrieves a specific user trait.
func (r *UserTraitRepository) GetByUserIDAndTraitType(ctx context.Context, userID int64, traitType string) (*models.UserTrait, error) {
	query := `
		SELECT id, user_id, trait_type, trait_value, created_at, updated_at
		FROM user_traits WHERE user_id = $1 AND trait_type = $2`

	var userTrait models.UserTrait
	err := r.db.QueryRow(ctx, query, userID, traitType).Scan(
		&userTrait.ID, &userTrait.UserID, &userTrait.TraitType, &userTrait.TraitValue, &userTrait.CreatedAt, &userTrait.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user trait not found")
		}
		return nil, fmt.Errorf("failed to get user trait: %w", err)
	}

	return &userTrait, nil
}

// Update updates a user trait.
func (r *UserTraitRepository) Update(ctx context.Context, id int64, req *models.UpdateUserTraitRequest) (*models.UserTrait, error) {
	query := `
		UPDATE user_traits SET trait_value = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, trait_type, trait_value, created_at, updated_at`

	var userTrait models.UserTrait
	err := r.db.QueryRow(ctx, query, req.TraitValue, id).Scan(
		&userTrait.ID, &userTrait.UserID, &userTrait.TraitType, &userTrait.TraitValue, &userTrait.CreatedAt, &userTrait.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user trait not found")
		}
		return nil, fmt.Errorf("failed to update user trait: %w", err)
	}

	return &userTrait, nil
}

// Delete deletes a user trait.
func (r *UserTraitRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM user_traits WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user trait: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user trait not found")
	}

	return nil
}

// DeleteByUserIDAndTraitType deletes a user trait by user ID and trait type.
func (r *UserTraitRepository) DeleteByUserIDAndTraitType(ctx context.Context, userID int64, traitType string) error {
	query := `DELETE FROM user_traits WHERE user_id = $1 AND trait_type = $2`
	result, err := r.db.Exec(ctx, query, userID, traitType)
	if err != nil {
		return fmt.Errorf("failed to delete user trait: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user trait not found")
	}

	return nil
}

// List retrieves a list of user traits with pagination.
func (r *UserTraitRepository) List(ctx context.Context, userID int64, page, perPage int) (*models.UserTraitListResponse, error) {
	// Set default pagination
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	// Count total
	countQuery := `SELECT COUNT(*) FROM user_traits WHERE user_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count user traits: %w", err)
	}

	// Get user traits
	query := `
		SELECT id, user_id, trait_type, trait_value, created_at, updated_at
		FROM user_traits WHERE user_id = $1
		ORDER BY trait_type ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user traits: %w", err)
	}
	defer rows.Close()

	var userTraits []models.UserTrait
	for rows.Next() {
		var userTrait models.UserTrait
		err := rows.Scan(
			&userTrait.ID, &userTrait.UserID, &userTrait.TraitType, &userTrait.TraitValue, &userTrait.CreatedAt, &userTrait.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user trait: %w", err)
		}
		userTraits = append(userTraits, userTrait)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user traits: %w", err)
	}

	// Calculate total pages
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	// Convert to responses
	userTraitResponses := make([]models.UserTraitResponse, len(userTraits))
	for i, userTrait := range userTraits {
		userTraitResponses[i] = userTrait.ToResponse()
	}

	return &models.UserTraitListResponse{
		Traits:     userTraitResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// Exists checks if a user trait exists.
func (r *UserTraitRepository) Exists(ctx context.Context, userID int64, traitType string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_traits WHERE user_id = $1 AND trait_type = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, userID, traitType).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user trait existence: %w", err)
	}
	return exists, nil
}
