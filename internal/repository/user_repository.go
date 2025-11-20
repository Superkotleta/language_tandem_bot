package repository

import (
	"context"
	"errors"
	"time"

	"language-exchange-bot/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetBySocialID(ctx context.Context, socialID, platform string) (*domain.User, error) {
	query := `
		SELECT id, social_id, platform, first_name, username, language, created_at, updated_at
		FROM users
		WHERE social_id = $1 AND platform = $2
	`

	row := r.db.QueryRow(ctx, query, socialID, platform)
	user := &domain.User{}

	err := row.Scan(
		&user.ID,
		&user.SocialID,
		&user.Platform,
		&user.FirstName,
		&user.Username,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}

	return user, nil
}

// CreateOrUpdate saves the user. If user exists (by social_id+platform), it updates the info.
func (r *UserRepository) CreateOrUpdate(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (social_id, platform, first_name, username, language, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (social_id, platform) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			username = EXCLUDED.username,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	err := r.db.QueryRow(ctx, query,
		user.SocialID,
		user.Platform,
		user.FirstName,
		user.Username,
		user.Language,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	return err
}


