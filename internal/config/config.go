// Package config provides application-wide configuration.
package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Environment string

const (
	Dev  Environment = "dev"
	Prod Environment = "prod"
)

// Config holds the application configuration.
type Config struct {
	// Env is the environment the application is running in (e.g., "prod", "dev").
	Env Environment
	// ServerAddr is the address the HTTP server listens on.
	ServerAddr string
	// LogLevel is the logging level (e.g., "debug", "info", "warn", "error").
	LogLevel string
}

var (
	Global *Config
	once   sync.Once
)

func init() {
	once.Do(func() {
		Global = Load()
	})
}

// Load reads configuration from environment variables and returns a Config.
// Defaults are provided for all values.
func Load() *Config {
	if err := godotenv.Load(".env.local"); err != nil {
		fmt.Printf("No .env.local file found or error loading it: %v\n", err)
	}

	return &Config{
		Env:        Environment(getEnv("ENV", string(Dev))),
		ServerAddr: fmt.Sprintf(":%s", getEnv("PORT", "8080")),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv retrieves an environment variable or returns a default value if not set.
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
