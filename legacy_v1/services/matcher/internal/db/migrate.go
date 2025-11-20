package db

import (
	"fmt"
	"log"

	"matcher/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source for migrations
)

// RunMigrations applies database migrations for the matcher service.
func RunMigrations(cfg *config.Config) error {
	src := "file://" + cfg.MigrationsDir
	m, err := migrate.New(src, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("Error closing migration source: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("Error closing migration database: %v", dbErr)
		}
	}()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("m.Up: %w", err)
	}
	return nil
}
