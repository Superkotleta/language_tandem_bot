package repository

import (
	"context"
	"fmt"
	"strings"

	"profile/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LanguageRepository struct {
	db *pgxpool.Pool
}

func NewLanguageRepository(db *pgxpool.Pool) *LanguageRepository {
	return &LanguageRepository{db: db}
}

// Create creates a new language.
func (r *LanguageRepository) Create(ctx context.Context, code, name, nativeName string) (*models.Language, error) {
	query := `
		INSERT INTO languages (code, name, native_name)
		VALUES ($1, $2, $3)
		RETURNING id, is_active, created_at`

	var language models.Language
	err := r.db.QueryRow(ctx, query, code, name, nativeName).Scan(
		&language.ID, &language.IsActive, &language.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create language: %w", err)
	}

	language.Code = code
	language.Name = name
	language.NativeName = nativeName

	return &language, nil
}

// GetByID retrieves a language by ID.
func (r *LanguageRepository) GetByID(ctx context.Context, id int64) (*models.Language, error) {
	query := `
		SELECT id, code, name, native_name, is_active, created_at
		FROM languages WHERE id = $1`

	var language models.Language
	err := r.db.QueryRow(ctx, query, id).Scan(
		&language.ID, &language.Code, &language.Name, &language.NativeName, &language.IsActive, &language.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("language not found")
		}
		return nil, fmt.Errorf("failed to get language: %w", err)
	}

	return &language, nil
}

// GetByCode retrieves a language by code.
func (r *LanguageRepository) GetByCode(ctx context.Context, code string) (*models.Language, error) {
	query := `
		SELECT id, code, name, native_name, is_active, created_at
		FROM languages WHERE code = $1`

	var language models.Language
	err := r.db.QueryRow(ctx, query, code).Scan(
		&language.ID, &language.Code, &language.Name, &language.NativeName, &language.IsActive, &language.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("language not found")
		}
		return nil, fmt.Errorf("failed to get language by code: %w", err)
	}

	return &language, nil
}

// List retrieves a list of languages with pagination.
func (r *LanguageRepository) List(ctx context.Context, req *models.LanguageSearchRequest) (*models.LanguageListResponse, error) {
	// Build WHERE clause
	whereParts := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if req.Query != "" {
		whereParts = append(whereParts, fmt.Sprintf("(name ILIKE $%d OR native_name ILIKE $%d OR code ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+req.Query+"%")
		argIndex++
	}
	if req.IsActive != nil {
		whereParts = append(whereParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM languages WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count languages: %w", err)
	}

	// Get languages
	query := fmt.Sprintf(`
		SELECT id, code, name, native_name, is_active, created_at
		FROM languages WHERE %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, req.PerPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list languages: %w", err)
	}
	defer rows.Close()

	var languages []models.Language
	for rows.Next() {
		var language models.Language
		err := rows.Scan(
			&language.ID, &language.Code, &language.Name, &language.NativeName, &language.IsActive, &language.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan language: %w", err)
		}
		languages = append(languages, language)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate languages: %w", err)
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	return &models.LanguageListResponse{
		Languages:  languages,
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
	}, nil
}

// Update updates a language.
func (r *LanguageRepository) Update(ctx context.Context, id int64, name, nativeName string, isActive bool) (*models.Language, error) {
	query := `
		UPDATE languages SET name = $1, native_name = $2, is_active = $3
		WHERE id = $4
		RETURNING id, code, name, native_name, is_active, created_at`

	var language models.Language
	err := r.db.QueryRow(ctx, query, name, nativeName, isActive, id).Scan(
		&language.ID, &language.Code, &language.Name, &language.NativeName, &language.IsActive, &language.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("language not found")
		}
		return nil, fmt.Errorf("failed to update language: %w", err)
	}

	return &language, nil
}

// Delete deletes a language.
func (r *LanguageRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM languages WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete language: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("language not found")
	}

	return nil
}

// Exists checks if a language exists by ID.
func (r *LanguageRepository) Exists(ctx context.Context, id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM languages WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check language existence: %w", err)
	}
	return exists, nil
}

// ExistsByCode checks if a language exists by code.
func (r *LanguageRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM languages WHERE code = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, code).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check language existence by code: %w", err)
	}
	return exists, nil
}
