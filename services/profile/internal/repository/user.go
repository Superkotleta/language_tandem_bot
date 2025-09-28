package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"profile/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user.
func (r *UserRepository) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	query := `
		INSERT INTO users (
			telegram_id, discord_id, username, first_name, last_name,
			email, phone, bio, age, gender, country, city, timezone, profile_picture_url
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, created_at, updated_at`

	var user models.User
	err := r.db.QueryRow(ctx, query,
		req.TelegramID, req.DiscordID, req.Username, req.FirstName, req.LastName,
		req.Email, req.Phone, req.Bio, req.Age, req.Gender, req.Country, req.City, req.Timezone, req.ProfilePictureURL,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Set the request fields
	user.TelegramID = req.TelegramID
	user.DiscordID = req.DiscordID
	user.Username = req.Username
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email
	user.Phone = req.Phone
	user.Bio = req.Bio
	user.Age = req.Age
	user.Gender = req.Gender
	user.Country = req.Country
	user.City = req.City
	user.Timezone = req.Timezone
	user.ProfilePictureURL = req.ProfilePictureURL
	user.IsActive = true
	user.IsVerified = false
	user.LastSeen = time.Now()

	return &user, nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, discord_id, username, first_name, last_name,
			   email, phone, bio, age, gender, country, city, timezone, profile_picture_url,
			   is_active, is_verified, last_seen, created_at, updated_at
		FROM users WHERE id = $1`

	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.TelegramID, &user.DiscordID, &user.Username, &user.FirstName, &user.LastName,
		&user.Email, &user.Phone, &user.Bio, &user.Age, &user.Gender, &user.Country, &user.City, &user.Timezone, &user.ProfilePictureURL,
		&user.IsActive, &user.IsVerified, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByTelegramID retrieves a user by Telegram ID.
func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, discord_id, username, first_name, last_name,
			   email, phone, bio, age, gender, country, city, timezone, profile_picture_url,
			   is_active, is_verified, last_seen, created_at, updated_at
		FROM users WHERE telegram_id = $1`

	var user models.User
	err := r.db.QueryRow(ctx, query, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.DiscordID, &user.Username, &user.FirstName, &user.LastName,
		&user.Email, &user.Phone, &user.Bio, &user.Age, &user.Gender, &user.Country, &user.City, &user.Timezone, &user.ProfilePictureURL,
		&user.IsActive, &user.IsVerified, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}

	return &user, nil
}

// GetByDiscordID retrieves a user by Discord ID.
func (r *UserRepository) GetByDiscordID(ctx context.Context, discordID int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, discord_id, username, first_name, last_name,
			   email, phone, bio, age, gender, country, city, timezone, profile_picture_url,
			   is_active, is_verified, last_seen, created_at, updated_at
		FROM users WHERE discord_id = $1`

	var user models.User
	err := r.db.QueryRow(ctx, query, discordID).Scan(
		&user.ID, &user.TelegramID, &user.DiscordID, &user.Username, &user.FirstName, &user.LastName,
		&user.Email, &user.Phone, &user.Bio, &user.Age, &user.Gender, &user.Country, &user.City, &user.Timezone, &user.ProfilePictureURL,
		&user.IsActive, &user.IsVerified, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by discord ID: %w", err)
	}

	return &user, nil
}

// Update updates a user.
func (r *UserRepository) Update(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	// Build dynamic query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *req.Username)
		argIndex++
	}
	if req.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *req.FirstName)
		argIndex++
	}
	if req.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *req.LastName)
		argIndex++
	}
	if req.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *req.Email)
		argIndex++
	}
	if req.Phone != nil {
		setParts = append(setParts, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, *req.Phone)
		argIndex++
	}
	if req.Bio != nil {
		setParts = append(setParts, fmt.Sprintf("bio = $%d", argIndex))
		args = append(args, *req.Bio)
		argIndex++
	}
	if req.Age != nil {
		setParts = append(setParts, fmt.Sprintf("age = $%d", argIndex))
		args = append(args, *req.Age)
		argIndex++
	}
	if req.Gender != nil {
		setParts = append(setParts, fmt.Sprintf("gender = $%d", argIndex))
		args = append(args, *req.Gender)
		argIndex++
	}
	if req.Country != nil {
		setParts = append(setParts, fmt.Sprintf("country = $%d", argIndex))
		args = append(args, *req.Country)
		argIndex++
	}
	if req.City != nil {
		setParts = append(setParts, fmt.Sprintf("city = $%d", argIndex))
		args = append(args, *req.City)
		argIndex++
	}
	if req.Timezone != nil {
		setParts = append(setParts, fmt.Sprintf("timezone = $%d", argIndex))
		args = append(args, *req.Timezone)
		argIndex++
	}
	if req.ProfilePictureURL != nil {
		setParts = append(setParts, fmt.Sprintf("profile_picture_url = $%d", argIndex))
		args = append(args, *req.ProfilePictureURL)
		argIndex++
	}
	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id)
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING id, telegram_id, discord_id, username, first_name, last_name,
				  email, phone, bio, age, gender, country, city, timezone, profile_picture_url,
				  is_active, is_verified, last_seen, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	var user models.User
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.TelegramID, &user.DiscordID, &user.Username, &user.FirstName, &user.LastName,
		&user.Email, &user.Phone, &user.Bio, &user.Age, &user.Gender, &user.Country, &user.City, &user.Timezone, &user.ProfilePictureURL,
		&user.IsActive, &user.IsVerified, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// Delete deletes a user.
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves a list of users with pagination.
func (r *UserRepository) List(ctx context.Context, req *models.UserSearchRequest) (*models.UserListResponse, error) {
	// Build WHERE clause
	whereParts := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if req.Query != "" {
		whereParts = append(whereParts, fmt.Sprintf("(username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d OR bio ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+req.Query+"%")
		argIndex++
	}
	if req.Country != "" {
		whereParts = append(whereParts, fmt.Sprintf("country = $%d", argIndex))
		args = append(args, req.Country)
		argIndex++
	}
	if req.City != "" {
		whereParts = append(whereParts, fmt.Sprintf("city = $%d", argIndex))
		args = append(args, req.City)
		argIndex++
	}
	if req.MinAge != nil {
		whereParts = append(whereParts, fmt.Sprintf("age >= $%d", argIndex))
		args = append(args, *req.MinAge)
		argIndex++
	}
	if req.MaxAge != nil {
		whereParts = append(whereParts, fmt.Sprintf("age <= $%d", argIndex))
		args = append(args, *req.MaxAge)
		argIndex++
	}
	if req.Gender != "" {
		whereParts = append(whereParts, fmt.Sprintf("gender = $%d", argIndex))
		args = append(args, req.Gender)
		argIndex++
	}

	whereClause := strings.Join(whereParts, " AND ")

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

	offset := (req.Page - 1) * req.PerPage

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	query := fmt.Sprintf(`
		SELECT id, telegram_id, discord_id, username, first_name, last_name,
			   email, phone, bio, age, gender, country, city, timezone, profile_picture_url,
			   is_active, is_verified, last_seen, created_at, updated_at
		FROM users WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, req.PerPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.TelegramID, &user.DiscordID, &user.Username, &user.FirstName, &user.LastName,
			&user.Email, &user.Phone, &user.Bio, &user.Age, &user.Gender, &user.Country, &user.City, &user.Timezone, &user.ProfilePictureURL,
			&user.IsActive, &user.IsVerified, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	// Convert to responses
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	return &models.UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
	}, nil
}

// UpdateLastSeen updates the last seen timestamp for a user.
func (r *UserRepository) UpdateLastSeen(ctx context.Context, id int64) error {
	query := `UPDATE users SET last_seen = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}
	return nil
}

// Exists checks if a user exists by ID.
func (r *UserRepository) Exists(ctx context.Context, id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}
