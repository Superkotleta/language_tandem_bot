package config

import (
	"os"
	"strconv"
)

// Config represents the API Gateway configuration.
type Config struct {
	HTTPPort        string
	Debug           bool
	ProfileService  ServiceConfig
	BotService      ServiceConfig
	RateLimitConfig RateLimitConfig
}

// ServiceConfig represents configuration for a backend service.
type ServiceConfig struct {
	URL     string
	Timeout int // in seconds
}

// RateLimitConfig represents rate limiting configuration.
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerMinute int
	BurstSize         int
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			return parsed
		}
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return def
}

// LoadAPIGateway loads the API Gateway configuration.
func LoadAPIGateway() *Config {
	return &Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		Debug:    getEnvBool("DEBUG", false),
		ProfileService: ServiceConfig{
			URL:     getEnv("PROFILE_SERVICE_URL", "http://localhost:8081"),
			Timeout: getEnvInt("PROFILE_SERVICE_TIMEOUT", 30),
		},
		BotService: ServiceConfig{
			URL:     getEnv("BOT_SERVICE_URL", "http://localhost:8082"),
			Timeout: getEnvInt("BOT_SERVICE_TIMEOUT", 30),
		},
		RateLimitConfig: RateLimitConfig{
			Enabled:           getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMinute: getEnvInt("RATE_LIMIT_RPM", 100),
			BurstSize:         getEnvInt("RATE_LIMIT_BURST", 20),
		},
	}
}
