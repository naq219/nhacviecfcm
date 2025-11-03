package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds application configuration
type Config struct {
	ServerAddr     string
	WorkerInterval int    // seconds
	FCMCredentials string // path to firebase credentials JSON
	Environment    string // development, production
}

// ValidationError represents configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation error for %s: %s", e.Field, e.Message)
}

// Load loads configuration from environment variables with validation
func Load() (*Config, error) {
	cfg := &Config{
		ServerAddr:     getEnv("SERVER_ADDR", "127.0.0.1:8888"),
		WorkerInterval: getEnvInt("WORKER_INTERVAL", 10),
		FCMCredentials: getEnv("FCM_CREDENTIALS", "./firebase-credentials.json"),
		Environment:    getEnv("ENVIRONMENT", "development"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates configuration values
func (c *Config) Validate() error {
	// Validate ServerAddr
	if c.ServerAddr == "" {
		return &ValidationError{Field: "ServerAddr", Message: "cannot be empty"}
	}
	if !strings.Contains(c.ServerAddr, ":") {
		return &ValidationError{Field: "ServerAddr", Message: "must contain port (format: host:port)"}
	}

	// Validate WorkerInterval
	if c.WorkerInterval <= 0 {
		return &ValidationError{Field: "WorkerInterval", Message: "must be positive"}
	}
	if c.WorkerInterval > 3600 {
		return &ValidationError{Field: "WorkerInterval", Message: "cannot exceed 3600 seconds (1 hour)"}
	}

	// Validate FCMCredentials
	if c.FCMCredentials == "" {
		return &ValidationError{Field: "FCMCredentials", Message: "cannot be empty"}
	}

	// Validate Environment
	validEnvs := []string{"development", "production", "testing"}
	if !contains(validEnvs, c.Environment) {
		return &ValidationError{Field: "Environment", Message: fmt.Sprintf("must be one of: %s", strings.Join(validEnvs, ", "))}
	}

	return nil
}

// IsDevelopment returns true if environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsTesting returns true if environment is testing
func (c *Config) IsTesting() bool {
	return c.Environment == "testing"
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt gets environment variable as integer with fallback
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
