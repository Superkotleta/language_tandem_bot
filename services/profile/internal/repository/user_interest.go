package repository

import (
	"context"
	"fmt"

	"profile/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserInterestRepository struct {
	db *pgxpool.Pool
}

func NewUserInterestRepository(db *pgxpool.Pool) *UserInterestRepository {
	return &UserInterestRepository{db: db}
}

// Create creates a new user interest.
func (r *UserInterestRepository) Create(ctx context.Context, req *models.CreateUserInterestRequest, userID int64) (*models.UserInterest, error) {
	query := `
		INSERT INTO user_interests (user_id, interest_id)
		VALUES ($1, $2)
		RETURNING id, created_at`

	var userInterest models.UserInterest
	err := r.db.QueryRow(ctx, query, userID, req.InterestID).Scan(
		&userInterest.ID, &userInterest.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user interest: %w", err)
	}

	userInterest.UserID = userID
	userInterest.InterestID = req.InterestID

	return &userInterest, nil
}

// GetByID retrieves a user interest by ID.
func (r *UserInterestRepository) GetByID(ctx context.Context, id int64) (*models.UserInterest, error) {
	query := `
		SELECT ui.id, ui.user_id, ui.interest_id, ui.created_at,
			   i.id, i.name, i.category, i.description, i.is_active, i.created_at
		FROM user_interests ui
		JOIN interests i ON ui.interest_id = i.id
		WHERE ui.id = $1`

	var userInterest models.UserInterest
	var interest models.Interest
	err := r.db.QueryRow(ctx, query, id).Scan(
		&userInterest.ID, &userInterest.UserID, &userInterest.InterestID, &userInterest.CreatedAt,
		&interest.ID, &interest.Name, &interest.Category, &interest.Description, &interest.IsActive, &interest.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user interest not found")
		}
		return nil, fmt.Errorf("failed to get user interest: %w", err)
	}

	userInterest.Interest = &interest
	return &userInterest, nil
}

// GetByUserID retrieves all interests for a user.
func (r *UserInterestRepository) GetByUserID(ctx context.Context, userID int64) ([]models.UserInterest, error) {
	query := `
		SELECT ui.id, ui.user_id, ui.interest_id, ui.created_at,
			   i.id, i.name, i.category, i.description, i.is_active, i.created_at
		FROM user_interests ui
		JOIN interests i ON ui.interest_id = i.id
		WHERE ui.user_id = $1
		ORDER BY ui.created_at ASC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user interests: %w", err)
	}
	defer rows.Close()

	var userInterests []models.UserInterest
	for rows.Next() {
		var userInterest models.UserInterest
		var interest models.Interest
		err := rows.Scan(
			&userInterest.ID, &userInterest.UserID, &userInterest.InterestID, &userInterest.CreatedAt,
			&interest.ID, &interest.Name, &interest.Category, &interest.Description, &interest.IsActive, &interest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user interest: %w", err)
		}
		userInterest.Interest = &interest
		userInterests = append(userInterests, userInterest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user interests: %w", err)
	}

	return userInterests, nil
}

// GetByUserIDAndInterestID retrieves a specific user interest.
func (r *UserInterestRepository) GetByUserIDAndInterestID(ctx context.Context, userID, interestID int64) (*models.UserInterest, error) {
	query := `
		SELECT ui.id, ui.user_id, ui.interest_id, ui.created_at,
			   i.id, i.name, i.category, i.description, i.is_active, i.created_at
		FROM user_interests ui
		JOIN interests i ON ui.interest_id = i.id
		WHERE ui.user_id = $1 AND ui.interest_id = $2`

	var userInterest models.UserInterest
	var interest models.Interest
	err := r.db.QueryRow(ctx, query, userID, interestID).Scan(
		&userInterest.ID, &userInterest.UserID, &userInterest.InterestID, &userInterest.CreatedAt,
		&interest.ID, &interest.Name, &interest.Category, &interest.Description, &interest.IsActive, &interest.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user interest not found")
		}
		return nil, fmt.Errorf("failed to get user interest: %w", err)
	}

	userInterest.Interest = &interest
	return &userInterest, nil
}

// Delete deletes a user interest.
func (r *UserInterestRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM user_interests WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user interest: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user interest not found")
	}

	return nil
}

// DeleteByUserIDAndInterestID deletes a user interest by user ID and interest ID.
func (r *UserInterestRepository) DeleteByUserIDAndInterestID(ctx context.Context, userID, interestID int64) error {
	query := `DELETE FROM user_interests WHERE user_id = $1 AND interest_id = $2`
	result, err := r.db.Exec(ctx, query, userID, interestID)
	if err != nil {
		return fmt.Errorf("failed to delete user interest: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user interest not found")
	}

	return nil
}

// List retrieves a list of user interests with pagination.
func (r *UserInterestRepository) List(ctx context.Context, userID int64, page, perPage int) (*models.UserInterestListResponse, error) {
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
	countQuery := `SELECT COUNT(*) FROM user_interests WHERE user_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count user interests: %w", err)
	}

	// Get user interests
	query := `
		SELECT ui.id, ui.user_id, ui.interest_id, ui.created_at,
			   i.id, i.name, i.category, i.description, i.is_active, i.created_at
		FROM user_interests ui
		JOIN interests i ON ui.interest_id = i.id
		WHERE ui.user_id = $1
		ORDER BY ui.created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user interests: %w", err)
	}
	defer rows.Close()

	var userInterests []models.UserInterest
	for rows.Next() {
		var userInterest models.UserInterest
		var interest models.Interest
		err := rows.Scan(
			&userInterest.ID, &userInterest.UserID, &userInterest.InterestID, &userInterest.CreatedAt,
			&interest.ID, &interest.Name, &interest.Category, &interest.Description, &interest.IsActive, &interest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user interest: %w", err)
		}
		userInterest.Interest = &interest
		userInterests = append(userInterests, userInterest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user interests: %w", err)
	}

	// Calculate total pages
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	// Convert to responses
	userInterestResponses := make([]models.UserInterestResponse, len(userInterests))
	for i, userInterest := range userInterests {
		userInterestResponses[i] = userInterest.ToResponse()
	}

	return &models.UserInterestListResponse{
		Interests:  userInterestResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// Exists checks if a user interest exists.
func (r *UserInterestRepository) Exists(ctx context.Context, userID, interestID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_interests WHERE user_id = $1 AND interest_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, userID, interestID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user interest existence: %w", err)
	}
	return exists, nil
}
