// Package config provides configuration management for the matcher service.
package config

import (
	"os"
)

// Config holds the configuration for the matcher service.
type Config struct {
	DatabaseURL   string
	DBSchema      string
	MigrationsDir string
	HTTPPort      string
	Debug         bool
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// LoadMatcher loads configuration from environment variables for the matcher service.
func LoadMatcher() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://matching_rw:matching_pwd@postgres:5432/languagebot?sslmode=disable"),
		DBSchema:      getEnv("DB_SCHEMA", "matching"),
		MigrationsDir: getEnv("MIGRATIONS_DIR", "/migrations/matching"),
		HTTPPort:      getEnv("HTTP_PORT", "8082"),
		Debug:         getEnv("DEBUG", "false") == "true",
	}
}
