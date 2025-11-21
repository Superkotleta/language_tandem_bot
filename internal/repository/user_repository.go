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
		SELECT id::text, social_id, platform, COALESCE(first_name, ''), COALESCE(username, ''), 
		       COALESCE(native_lang, ''), COALESCE(target_lang, ''), COALESCE(target_level, ''), 
		       interface_lang, status, created_at, updated_at
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
		&user.NativeLang,
		&user.TargetLang,
		&user.TargetLevel,
		&user.InterfaceLang,
		&user.Status,
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

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (social_id, platform, first_name, username, interface_lang, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text
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
		user.InterfaceLang,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	return err
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET first_name = $1, username = $2, native_lang = NULLIF($3, ''), target_lang = NULLIF($4, ''), 
		    target_level = NULLIF($5, ''), interface_lang = $6, status = $7, updated_at = $8
		WHERE id = $9
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		user.FirstName,
		user.Username,
		user.NativeLang,
		user.TargetLang,
		user.TargetLevel,
		user.InterfaceLang,
		user.Status,
		user.UpdatedAt,
		user.ID,
	)

	return err
}

func (r *UserRepository) AddInterest(ctx context.Context, userID, interestID string) error {
	query := `INSERT INTO user_interests (user_id, interest_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, userID, interestID)

	return err
}

func (r *UserRepository) RemoveInterest(ctx context.Context, userID, interestID string) error {
	query := `DELETE FROM user_interests WHERE user_id = $1 AND interest_id = $2`
	_, err := r.db.Exec(ctx, query, userID, interestID)

	return err
}

func (r *UserRepository) GetUserInterests(ctx context.Context, userID string) ([]domain.Interest, error) {
	query := `
		SELECT i.id::text, i.category_id::text, i.slug, i.names
		FROM interests i
		JOIN user_interests ui ON i.id = ui.interest_id
		WHERE ui.user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var interests []domain.Interest

	for rows.Next() {
		var i domain.Interest
		if err := rows.Scan(&i.ID, &i.CategoryID, &i.Slug, &i.Names); err != nil {
			return nil, err
		}

		interests = append(interests, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return interests, nil
}
