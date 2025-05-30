package config

import (
	"linked-clone/internal/config"
	"os"
)

func LoadTestConfig() *config.Config {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("DB_NAME", "linkedin_clone_test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "tyokeren")
	os.Setenv("DB_PASSWORD", "14Oktober04.")
	os.Setenv("DB_SSLMODE", "disable")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key")
	os.Setenv("JWT_EXPIRY_HOURS", "24")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("LOG_LEVEL", "error")

	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load test config: " + err.Error())
	}

	return cfg
}
