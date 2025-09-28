package repository

import (
	"context"
	"fmt"
	"strings"

	"profile/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InterestRepository struct {
	db *pgxpool.Pool
}

func NewInterestRepository(db *pgxpool.Pool) *InterestRepository {
	return &InterestRepository{db: db}
}

// Create creates a new interest.
func (r *InterestRepository) Create(ctx context.Context, name, category string, description *string) (*models.Interest, error) {
	query := `
		INSERT INTO interests (name, category, description)
		VALUES ($1, $2, $3)
		RETURNING id, is_active, created_at`

	var interest models.Interest
	err := r.db.QueryRow(ctx, query, name, category, description).Scan(
		&interest.ID, &interest.IsActive, &interest.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create interest: %w", err)
	}

	interest.Name = name
	interest.Category = category
	interest.Description = description

	return &interest, nil
}

// GetByID retrieves an interest by ID.
func (r *InterestRepository) GetByID(ctx context.Context, id int64) (*models.Interest, error) {
	query := `
		SELECT id, name, category, description, is_active, created_at
		FROM interests WHERE id = $1`

	var interest models.Interest
	err := r.db.QueryRow(ctx, query, id).Scan(
		&interest.ID, &interest.Name, &interest.Category, &interest.Description, &interest.IsActive, &interest.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("interest not found")
		}
		return nil, fmt.Errorf("failed to get interest: %w", err)
	}

	return &interest, nil
}

// GetByName retrieves an interest by name.
func (r *InterestRepository) GetByName(ctx context.Context, name string) (*models.Interest, error) {
	query := `
		SELECT id, name, category, description, is_active, created_at
		FROM interests WHERE name = $1`

	var interest models.Interest
	err := r.db.QueryRow(ctx, query, name).Scan(
		&interest.ID, &interest.Name, &interest.Category, &interest.Description, &interest.IsActive, &interest.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("interest not found")
		}
		return nil, fmt.Errorf("failed to get interest by name: %w", err)
	}

	return &interest, nil
}

// List retrieves a list of interests with pagination.
func (r *InterestRepository) List(ctx context.Context, req *models.InterestSearchRequest) (*models.InterestListResponse, error) {
	// Build WHERE clause
	whereParts := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if req.Query != "" {
		whereParts = append(whereParts, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+req.Query+"%")
		argIndex++
	}
	if req.Category != "" {
		whereParts = append(whereParts, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, req.Category)
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM interests WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count interests: %w", err)
	}

	// Get interests
	query := fmt.Sprintf(`
		SELECT id, name, category, description, is_active, created_at
		FROM interests WHERE %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, req.PerPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list interests: %w", err)
	}
	defer rows.Close()

	var interests []models.Interest
	for rows.Next() {
		var interest models.Interest
		err := rows.Scan(
			&interest.ID, &interest.Name, &interest.Category, &interest.Description, &interest.IsActive, &interest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan interest: %w", err)
		}
		interests = append(interests, interest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate interests: %w", err)
	}

	// Calculate total pages
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	return &models.InterestListResponse{
		Interests:  interests,
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
	}, nil
}

// Update updates an interest.
func (r *InterestRepository) Update(ctx context.Context, id int64, name, category string, description *string, isActive bool) (*models.Interest, error) {
	query := `
		UPDATE interests SET name = $1, category = $2, description = $3, is_active = $4
		WHERE id = $5
		RETURNING id, name, category, description, is_active, created_at`

	var interest models.Interest
	err := r.db.QueryRow(ctx, query, name, category, description, isActive, id).Scan(
		&interest.ID, &interest.Name, &interest.Category, &interest.Description, &interest.IsActive, &interest.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("interest not found")
		}
		return nil, fmt.Errorf("failed to update interest: %w", err)
	}

	return &interest, nil
}

// Delete deletes an interest.
func (r *InterestRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM interests WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete interest: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("interest not found")
	}

	return nil
}

// Exists checks if an interest exists by ID.
func (r *InterestRepository) Exists(ctx context.Context, id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM interests WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check interest existence: %w", err)
	}
	return exists, nil
}

// ExistsByName checks if an interest exists by name.
func (r *InterestRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM interests WHERE name = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check interest existence by name: %w", err)
	}
	return exists, nil
}
