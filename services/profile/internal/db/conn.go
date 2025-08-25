package db

import (
	"context"
	"fmt"

	"profile/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, err
	}
	// search_path = <schema>,public
	if _, err := pool.Exec(ctx, fmt.Sprintf("SET search_path = %s, public", cfg.DBSchema)); err != nil {
		pool.Close()
		return nil, err
	}
	// ping
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
