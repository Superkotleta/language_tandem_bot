package config

import (
	"os"
)

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

func LoadProfile() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://profile_rw:profile_pwd@postgres:5432/languagebot?sslmode=disable"),
		DBSchema:      getEnv("DB_SCHEMA", "profile"),
		MigrationsDir: getEnv("MIGRATIONS_DIR", "/migrations/profile"),
		HTTPPort:      getEnv("HTTP_PORT", "8081"),
		Debug:         getEnv("DEBUG", "false") == "true",
	}
}
