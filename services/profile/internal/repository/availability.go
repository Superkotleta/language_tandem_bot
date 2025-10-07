package repository

import (
	"context"
	"errors"
	"fmt"

	"profile/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserTimeAvailabilityRepository struct {
	db *pgxpool.Pool
}

func NewUserTimeAvailabilityRepository(db *pgxpool.Pool) *UserTimeAvailabilityRepository {
	return &UserTimeAvailabilityRepository{db: db}
}

// Create creates a new user time availability.
func (r *UserTimeAvailabilityRepository) Create(ctx context.Context, req *models.CreateUserTimeAvailabilityRequest, userID int64) (*models.UserTimeAvailability, error) {
	query := `
		INSERT INTO user_time_availability (user_id, day_of_week, start_time, end_time, timezone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	var userTimeAvailability models.UserTimeAvailability
	err := r.db.QueryRow(ctx, query, userID, req.DayOfWeek, req.StartTime, req.EndTime, req.Timezone).Scan(
		&userTimeAvailability.ID, &userTimeAvailability.CreatedAt, &userTimeAvailability.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user time availability: %w", err)
	}

	userTimeAvailability.UserID = userID
	userTimeAvailability.DayOfWeek = req.DayOfWeek
	userTimeAvailability.StartTime = req.StartTime
	userTimeAvailability.EndTime = req.EndTime
	userTimeAvailability.Timezone = req.Timezone

	return &userTimeAvailability, nil
}

// GetByID retrieves a user time availability by ID.
func (r *UserTimeAvailabilityRepository) GetByID(ctx context.Context, id int64) (*models.UserTimeAvailability, error) {
	query := `
		SELECT id, user_id, day_of_week, start_time, end_time, timezone, created_at, updated_at
		FROM user_time_availability WHERE id = $1`

	var userTimeAvailability models.UserTimeAvailability
	err := r.db.QueryRow(ctx, query, id).Scan(
		&userTimeAvailability.ID, &userTimeAvailability.UserID, &userTimeAvailability.DayOfWeek,
		&userTimeAvailability.StartTime, &userTimeAvailability.EndTime, &userTimeAvailability.Timezone,
		&userTimeAvailability.CreatedAt, &userTimeAvailability.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user time availability not found")
		}
		return nil, fmt.Errorf("failed to get user time availability: %w", err)
	}

	return &userTimeAvailability, nil
}

// GetByUserID retrieves all time availability for a user.
func (r *UserTimeAvailabilityRepository) GetByUserID(ctx context.Context, userID int64) ([]models.UserTimeAvailability, error) {
	query := `
		SELECT id, user_id, day_of_week, start_time, end_time, timezone, created_at, updated_at
		FROM user_time_availability WHERE user_id = $1
		ORDER BY day_of_week ASC, start_time ASC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user time availability: %w", err)
	}
	defer rows.Close()

	var userTimeAvailabilities []models.UserTimeAvailability
	for rows.Next() {
		var userTimeAvailability models.UserTimeAvailability
		err := rows.Scan(
			&userTimeAvailability.ID, &userTimeAvailability.UserID, &userTimeAvailability.DayOfWeek,
			&userTimeAvailability.StartTime, &userTimeAvailability.EndTime, &userTimeAvailability.Timezone,
			&userTimeAvailability.CreatedAt, &userTimeAvailability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user time availability: %w", err)
		}
		userTimeAvailabilities = append(userTimeAvailabilities, userTimeAvailability)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user time availability: %w", err)
	}

	return userTimeAvailabilities, nil
}

// GetByUserIDAndDayOfWeek retrieves time availability for a specific day.
func (r *UserTimeAvailabilityRepository) GetByUserIDAndDayOfWeek(ctx context.Context, userID int64, dayOfWeek int) ([]models.UserTimeAvailability, error) {
	query := `
		SELECT id, user_id, day_of_week, start_time, end_time, timezone, created_at, updated_at
		FROM user_time_availability WHERE user_id = $1 AND day_of_week = $2
		ORDER BY start_time ASC`

	rows, err := r.db.Query(ctx, query, userID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get user time availability: %w", err)
	}
	defer rows.Close()

	var userTimeAvailabilities []models.UserTimeAvailability
	for rows.Next() {
		var userTimeAvailability models.UserTimeAvailability
		err := rows.Scan(
			&userTimeAvailability.ID, &userTimeAvailability.UserID, &userTimeAvailability.DayOfWeek,
			&userTimeAvailability.StartTime, &userTimeAvailability.EndTime, &userTimeAvailability.Timezone,
			&userTimeAvailability.CreatedAt, &userTimeAvailability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user time availability: %w", err)
		}
		userTimeAvailabilities = append(userTimeAvailabilities, userTimeAvailability)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user time availability: %w", err)
	}

	return userTimeAvailabilities, nil
}

// Update updates a user time availability.
func (r *UserTimeAvailabilityRepository) Update(ctx context.Context, id int64, req *models.UpdateUserTimeAvailabilityRequest) (*models.UserTimeAvailability, error) {
	query := `
		UPDATE user_time_availability SET start_time = $1, end_time = $2, timezone = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, user_id, day_of_week, start_time, end_time, timezone, created_at, updated_at`

	var userTimeAvailability models.UserTimeAvailability
	err := r.db.QueryRow(ctx, query, req.StartTime, req.EndTime, req.Timezone, id).Scan(
		&userTimeAvailability.ID, &userTimeAvailability.UserID, &userTimeAvailability.DayOfWeek,
		&userTimeAvailability.StartTime, &userTimeAvailability.EndTime, &userTimeAvailability.Timezone,
		&userTimeAvailability.CreatedAt, &userTimeAvailability.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user time availability not found")
		}
		return nil, fmt.Errorf("failed to update user time availability: %w", err)
	}

	return &userTimeAvailability, nil
}

// Delete deletes a user time availability.
func (r *UserTimeAvailabilityRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM user_time_availability WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user time availability: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user time availability not found")
	}

	return nil
}

// DeleteByUserID deletes all time availability for a user.
func (r *UserTimeAvailabilityRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM user_time_availability WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user time availability: %w", err)
	}

	return nil
}

// List retrieves a list of user time availability with pagination.
func (r *UserTimeAvailabilityRepository) List(ctx context.Context, userID int64, page, perPage int) (*models.UserTimeAvailabilityListResponse, error) {
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
	countQuery := `SELECT COUNT(*) FROM user_time_availability WHERE user_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count user time availability: %w", err)
	}

	// Get user time availability
	query := `
		SELECT id, user_id, day_of_week, start_time, end_time, timezone, created_at, updated_at
		FROM user_time_availability WHERE user_id = $1
		ORDER BY day_of_week ASC, start_time ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user time availability: %w", err)
	}
	defer rows.Close()

	var userTimeAvailabilities []models.UserTimeAvailability
	for rows.Next() {
		var userTimeAvailability models.UserTimeAvailability
		err := rows.Scan(
			&userTimeAvailability.ID, &userTimeAvailability.UserID, &userTimeAvailability.DayOfWeek,
			&userTimeAvailability.StartTime, &userTimeAvailability.EndTime, &userTimeAvailability.Timezone,
			&userTimeAvailability.CreatedAt, &userTimeAvailability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user time availability: %w", err)
		}
		userTimeAvailabilities = append(userTimeAvailabilities, userTimeAvailability)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user time availability: %w", err)
	}

	// Calculate total pages
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	// Convert to responses
	userTimeAvailabilityResponses := make([]models.UserTimeAvailabilityResponse, len(userTimeAvailabilities))
	for i, userTimeAvailability := range userTimeAvailabilities {
		userTimeAvailabilityResponses[i] = userTimeAvailability.ToResponse()
	}

	return &models.UserTimeAvailabilityListResponse{
		Availability: userTimeAvailabilityResponses,
		Total:        total,
		Page:         page,
		PerPage:      perPage,
		TotalPages:   totalPages,
	}, nil
}

// Exists checks if a user time availability exists.
func (r *UserTimeAvailabilityRepository) Exists(ctx context.Context, userID int64, dayOfWeek int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_time_availability WHERE user_id = $1 AND day_of_week = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, userID, dayOfWeek).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user time availability existence: %w", err)
	}
	return exists, nil
}
