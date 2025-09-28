package repository

import (
	"context"
	"fmt"

	"profile/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserLanguageRepository struct {
	db *pgxpool.Pool
}

func NewUserLanguageRepository(db *pgxpool.Pool) *UserLanguageRepository {
	return &UserLanguageRepository{db: db}
}

// Create creates a new user language.
func (r *UserLanguageRepository) Create(ctx context.Context, req *models.CreateUserLanguageRequest, userID int64) (*models.UserLanguage, error) {
	query := `
		INSERT INTO user_languages (user_id, language_id, level, is_learning)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	var userLanguage models.UserLanguage
	err := r.db.QueryRow(ctx, query, userID, req.LanguageID, req.Level, req.IsLearning).Scan(
		&userLanguage.ID, &userLanguage.CreatedAt, &userLanguage.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user language: %w", err)
	}

	userLanguage.UserID = userID
	userLanguage.LanguageID = req.LanguageID
	userLanguage.Level = req.Level
	userLanguage.IsLearning = req.IsLearning

	return &userLanguage, nil
}

// GetByID retrieves a user language by ID.
func (r *UserLanguageRepository) GetByID(ctx context.Context, id int64) (*models.UserLanguage, error) {
	query := `
		SELECT ul.id, ul.user_id, ul.language_id, ul.level, ul.is_learning, ul.created_at, ul.updated_at,
			   l.id, l.code, l.name, l.native_name, l.is_active, l.created_at
		FROM user_languages ul
		JOIN languages l ON ul.language_id = l.id
		WHERE ul.id = $1`

	var userLanguage models.UserLanguage
	var language models.Language
	err := r.db.QueryRow(ctx, query, id).Scan(
		&userLanguage.ID, &userLanguage.UserID, &userLanguage.LanguageID, &userLanguage.Level, &userLanguage.IsLearning, &userLanguage.CreatedAt, &userLanguage.UpdatedAt,
		&language.ID, &language.Code, &language.Name, &language.NativeName, &language.IsActive, &language.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user language not found")
		}
		return nil, fmt.Errorf("failed to get user language: %w", err)
	}

	userLanguage.Language = &language
	return &userLanguage, nil
}

// GetByUserID retrieves all languages for a user.
func (r *UserLanguageRepository) GetByUserID(ctx context.Context, userID int64) ([]models.UserLanguage, error) {
	query := `
		SELECT ul.id, ul.user_id, ul.language_id, ul.level, ul.is_learning, ul.created_at, ul.updated_at,
			   l.id, l.code, l.name, l.native_name, l.is_active, l.created_at
		FROM user_languages ul
		JOIN languages l ON ul.language_id = l.id
		WHERE ul.user_id = $1
		ORDER BY ul.created_at ASC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user languages: %w", err)
	}
	defer rows.Close()

	var userLanguages []models.UserLanguage
	for rows.Next() {
		var userLanguage models.UserLanguage
		var language models.Language
		err := rows.Scan(
			&userLanguage.ID, &userLanguage.UserID, &userLanguage.LanguageID, &userLanguage.Level, &userLanguage.IsLearning, &userLanguage.CreatedAt, &userLanguage.UpdatedAt,
			&language.ID, &language.Code, &language.Name, &language.NativeName, &language.IsActive, &language.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user language: %w", err)
		}
		userLanguage.Language = &language
		userLanguages = append(userLanguages, userLanguage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user languages: %w", err)
	}

	return userLanguages, nil
}

// GetByUserIDAndLanguageID retrieves a specific user language.
func (r *UserLanguageRepository) GetByUserIDAndLanguageID(ctx context.Context, userID, languageID int64) (*models.UserLanguage, error) {
	query := `
		SELECT ul.id, ul.user_id, ul.language_id, ul.level, ul.is_learning, ul.created_at, ul.updated_at,
			   l.id, l.code, l.name, l.native_name, l.is_active, l.created_at
		FROM user_languages ul
		JOIN languages l ON ul.language_id = l.id
		WHERE ul.user_id = $1 AND ul.language_id = $2`

	var userLanguage models.UserLanguage
	var language models.Language
	err := r.db.QueryRow(ctx, query, userID, languageID).Scan(
		&userLanguage.ID, &userLanguage.UserID, &userLanguage.LanguageID, &userLanguage.Level, &userLanguage.IsLearning, &userLanguage.CreatedAt, &userLanguage.UpdatedAt,
		&language.ID, &language.Code, &language.Name, &language.NativeName, &language.IsActive, &language.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user language not found")
		}
		return nil, fmt.Errorf("failed to get user language: %w", err)
	}

	userLanguage.Language = &language
	return &userLanguage, nil
}

// Update updates a user language.
func (r *UserLanguageRepository) Update(ctx context.Context, id int64, req *models.UpdateUserLanguageRequest) (*models.UserLanguage, error) {
	query := `
		UPDATE user_languages SET level = $1, is_learning = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, user_id, language_id, level, is_learning, created_at, updated_at`

	var userLanguage models.UserLanguage
	err := r.db.QueryRow(ctx, query, req.Level, req.IsLearning, id).Scan(
		&userLanguage.ID, &userLanguage.UserID, &userLanguage.LanguageID, &userLanguage.Level, &userLanguage.IsLearning, &userLanguage.CreatedAt, &userLanguage.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user language not found")
		}
		return nil, fmt.Errorf("failed to update user language: %w", err)
	}

	// Get the language details
	languageRepo := NewLanguageRepository(r.db)
	language, err := languageRepo.GetByID(ctx, userLanguage.LanguageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get language details: %w", err)
	}
	userLanguage.Language = language

	return &userLanguage, nil
}

// Delete deletes a user language.
func (r *UserLanguageRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM user_languages WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user language: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user language not found")
	}

	return nil
}

// DeleteByUserIDAndLanguageID deletes a user language by user ID and language ID.
func (r *UserLanguageRepository) DeleteByUserIDAndLanguageID(ctx context.Context, userID, languageID int64) error {
	query := `DELETE FROM user_languages WHERE user_id = $1 AND language_id = $2`
	result, err := r.db.Exec(ctx, query, userID, languageID)
	if err != nil {
		return fmt.Errorf("failed to delete user language: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user language not found")
	}

	return nil
}

// List retrieves a list of user languages with pagination.
func (r *UserLanguageRepository) List(ctx context.Context, userID int64, page, perPage int) (*models.UserLanguageListResponse, error) {
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
	countQuery := `SELECT COUNT(*) FROM user_languages WHERE user_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count user languages: %w", err)
	}

	// Get user languages
	query := `
		SELECT ul.id, ul.user_id, ul.language_id, ul.level, ul.is_learning, ul.created_at, ul.updated_at,
			   l.id, l.code, l.name, l.native_name, l.is_active, l.created_at
		FROM user_languages ul
		JOIN languages l ON ul.language_id = l.id
		WHERE ul.user_id = $1
		ORDER BY ul.created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user languages: %w", err)
	}
	defer rows.Close()

	var userLanguages []models.UserLanguage
	for rows.Next() {
		var userLanguage models.UserLanguage
		var language models.Language
		err := rows.Scan(
			&userLanguage.ID, &userLanguage.UserID, &userLanguage.LanguageID, &userLanguage.Level, &userLanguage.IsLearning, &userLanguage.CreatedAt, &userLanguage.UpdatedAt,
			&language.ID, &language.Code, &language.Name, &language.NativeName, &language.IsActive, &language.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user language: %w", err)
		}
		userLanguage.Language = &language
		userLanguages = append(userLanguages, userLanguage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user languages: %w", err)
	}

	// Calculate total pages
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	// Convert to responses
	userLanguageResponses := make([]models.UserLanguageResponse, len(userLanguages))
	for i, userLanguage := range userLanguages {
		userLanguageResponses[i] = userLanguage.ToResponse()
	}

	return &models.UserLanguageListResponse{
		Languages:  userLanguageResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// Exists checks if a user language exists.
func (r *UserLanguageRepository) Exists(ctx context.Context, userID, languageID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_languages WHERE user_id = $1 AND language_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, userID, languageID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user language existence: %w", err)
	}
	return exists, nil
}
