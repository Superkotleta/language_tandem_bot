package repository

import (
	"context"
	"language-exchange-bot/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReferenceRepository struct {
	db *pgxpool.Pool
}

func NewReferenceRepository(db *pgxpool.Pool) *ReferenceRepository {
	return &ReferenceRepository{db: db}
}

func (r *ReferenceRepository) GetLanguages(ctx context.Context) ([]domain.Language, error) {
	query := `SELECT code, names, flag FROM languages`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var languages []domain.Language

	for rows.Next() {
		var l domain.Language
		if err := rows.Scan(&l.Code, &l.Names, &l.Flag); err != nil {
			return nil, err
		}

		languages = append(languages, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return languages, nil
}

func (r *ReferenceRepository) GetCategories(ctx context.Context) ([]domain.InterestCategory, error) {
	query := `SELECT id::text, slug, names, display_order FROM interest_categories ORDER BY display_order ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var categories []domain.InterestCategory

	for rows.Next() {
		var c domain.InterestCategory
		if err := rows.Scan(&c.ID, &c.Slug, &c.Names, &c.Order); err != nil {
			return nil, err
		}

		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *ReferenceRepository) GetInterestsByCategory(ctx context.Context, categoryID string) ([]domain.Interest, error) {
	query := `SELECT id::text, category_id::text, slug, names FROM interests WHERE category_id = $1`

	rows, err := r.db.Query(ctx, query, categoryID)
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
