package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	ServerAddr     string
	WorkerInterval int    // seconds
	FCMCredentials string // path to firebase credentials JSON
	Environment    string // development, production
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		ServerAddr:     getEnv("SERVER_ADDR", "127.0.0.1:8888"),
		WorkerInterval: 10, // 10 seconds
		FCMCredentials: getEnv("FCM_CREDENTIALS", "./firebase-credentials.json"),
		Environment:    getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
