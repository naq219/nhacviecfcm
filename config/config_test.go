package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Success(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_ADDR", "localhost:9000")
	os.Setenv("WORKER_INTERVAL", "30")
	os.Setenv("FCM_CREDENTIALS", "./test-credentials.json")
	os.Setenv("ENVIRONMENT", "production")
	defer func() {
		os.Unsetenv("SERVER_ADDR")
		os.Unsetenv("WORKER_INTERVAL")
		os.Unsetenv("FCM_CREDENTIALS")
		os.Unsetenv("ENVIRONMENT")
	}()

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost:9000", cfg.ServerAddr)
	assert.Equal(t, 30, cfg.WorkerInterval)
	assert.Equal(t, "./test-credentials.json", cfg.FCMCredentials)
	assert.Equal(t, "production", cfg.Environment)
}

func TestLoad_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("SERVER_ADDR")
	os.Unsetenv("WORKER_INTERVAL")
	os.Unsetenv("FCM_CREDENTIALS")
	os.Unsetenv("ENVIRONMENT")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1:8888", cfg.ServerAddr)
	assert.Equal(t, 10, cfg.WorkerInterval)
	assert.Equal(t, "./firebase-credentials.json", cfg.FCMCredentials)
	assert.Equal(t, "development", cfg.Environment)
}

func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		ServerAddr:     "localhost:8080",
		WorkerInterval: 60,
		FCMCredentials: "./credentials.json",
		Environment:    "development",
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_EmptyServerAddr(t *testing.T) {
	cfg := &Config{
		ServerAddr:     "",
		WorkerInterval: 60,
		FCMCredentials: "./credentials.json",
		Environment:    "development",
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ServerAddr")
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestValidate_InvalidServerAddr(t *testing.T) {
	cfg := &Config{
		ServerAddr:     "localhost",
		WorkerInterval: 60,
		FCMCredentials: "./credentials.json",
		Environment:    "development",
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ServerAddr")
	assert.Contains(t, err.Error(), "must contain port")
}

func TestValidate_InvalidWorkerInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval int
		expected string
	}{
		{"zero interval", 0, "must be positive"},
		{"negative interval", -5, "must be positive"},
		{"too large interval", 4000, "cannot exceed 3600 seconds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ServerAddr:     "localhost:8080",
				WorkerInterval: tt.interval,
				FCMCredentials: "./credentials.json",
				Environment:    "development",
			}

			err := cfg.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "WorkerInterval")
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

func TestValidate_EmptyFCMCredentials(t *testing.T) {
	cfg := &Config{
		ServerAddr:     "localhost:8080",
		WorkerInterval: 60,
		FCMCredentials: "",
		Environment:    "development",
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FCMCredentials")
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestValidate_InvalidEnvironment(t *testing.T) {
	cfg := &Config{
		ServerAddr:     "localhost:8080",
		WorkerInterval: 60,
		FCMCredentials: "./credentials.json",
		Environment:    "invalid",
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Environment")
	assert.Contains(t, err.Error(), "must be one of")
}

func TestEnvironmentCheckers(t *testing.T) {
	tests := []struct {
		env           string
		isDevelopment bool
		isProduction  bool
		isTesting     bool
	}{
		{"development", true, false, false},
		{"production", false, true, false},
		{"testing", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Environment: tt.env}
			assert.Equal(t, tt.isDevelopment, cfg.IsDevelopment())
			assert.Equal(t, tt.isProduction, cfg.IsProduction())
			assert.Equal(t, tt.isTesting, cfg.IsTesting())
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		fallback int
		expected int
	}{
		{"valid integer", "42", 10, 42},
		{"invalid integer", "abc", 10, 10},
		{"empty value", "", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_INT_VAR"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}

			result := getEnvInt(key, tt.fallback)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	assert.True(t, contains(slice, "apple"))
	assert.True(t, contains(slice, "banana"))
	assert.False(t, contains(slice, "grape"))
	assert.False(t, contains(slice, ""))
}