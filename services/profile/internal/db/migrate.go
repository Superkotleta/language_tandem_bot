package db

import (
	"errors"
	"fmt"

	"profile/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // PostgreSQL driver for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"       // File source driver for migrations
)

func RunMigrations(cfg *config.Config) error {
	src := "file://" + cfg.MigrationsDir
	m, err := migrate.New(src, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("m.Up: %w", err)
	}
	return nil
}
