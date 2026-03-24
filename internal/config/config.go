// Package config provides configuration management for the crypto bot.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration values.
type Config struct {
	// Database configuration
	DatabasePath string

	// Output paths
	ReadmePath      string
	HugoDataPath    string
	HugoHistoryPath string

	// API configuration
	APIBaseURL   string
	APITimeout   time.Duration
	MaxRetries   int
	RetryBackoff time.Duration

	// Application settings
	Timeout     time.Duration
	HistoryDays int
}

// Default configuration values.
const (
	defaultDatabasePath    = "./crypto_history.db"
	defaultReadmePath      = "./README.md"
	defaultHugoDataPath    = "./data/crypto.json"
	defaultHugoHistoryPath = "./data/history"
	defaultAPIBaseURL      = "https://api.coingecko.com/api/v3"
	defaultAPITimeout      = 30 * time.Second
	defaultMaxRetries      = 3
	defaultRetryBackoff    = time.Second
	defaultTimeout         = 5 * time.Minute
	defaultHistoryDays     = 30
)

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		DatabasePath:    getEnvOrDefault("CRYPTO_BOT_DB_PATH", defaultDatabasePath),
		ReadmePath:      getEnvOrDefault("CRYPTO_BOT_README_PATH", defaultReadmePath),
		HugoDataPath:    getEnvOrDefault("CRYPTO_BOT_HUGO_DATA_PATH", defaultHugoDataPath),
		HugoHistoryPath: getEnvOrDefault("CRYPTO_BOT_HUGO_HISTORY_PATH", defaultHugoHistoryPath),
		APIBaseURL:      getEnvOrDefault("CRYPTO_BOT_API_URL", defaultAPIBaseURL),
	}

	var err error

	cfg.APITimeout, err = parseDuration("CRYPTO_BOT_API_TIMEOUT", defaultAPITimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid API timeout: %w", err)
	}

	cfg.Timeout, err = parseDuration("CRYPTO_BOT_TIMEOUT", defaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout: %w", err)
	}

	cfg.MaxRetries, err = parseInt("CRYPTO_BOT_MAX_RETRIES", defaultMaxRetries)
	if err != nil {
		return nil, fmt.Errorf("invalid max retries: %w", err)
	}

	cfg.RetryBackoff, err = parseDuration("CRYPTO_BOT_RETRY_BACKOFF", defaultRetryBackoff)
	if err != nil {
		return nil, fmt.Errorf("invalid retry backoff: %w", err)
	}

	cfg.HistoryDays, err = parseInt("CRYPTO_BOT_HISTORY_DAYS", defaultHistoryDays)
	if err != nil {
		return nil, fmt.Errorf("invalid history days: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabasePath == "" {
		return fmt.Errorf("database path cannot be empty")
	}
	if c.ReadmePath == "" {
		return fmt.Errorf("readme path cannot be empty")
	}
	if c.APIBaseURL == "" {
		return fmt.Errorf("API base URL cannot be empty")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if c.HistoryDays <= 0 {
		return fmt.Errorf("history days must be positive")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	return time.ParseDuration(value)
}

func parseInt(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}
