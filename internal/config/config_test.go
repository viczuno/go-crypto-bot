package config_test

import (
	"testing"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"DatabasePath", cfg.DatabasePath, "./crypto_history.db"},
		{"ReadmePath", cfg.ReadmePath, "./README.md"},
		{"HugoDataPath", cfg.HugoDataPath, "./data/crypto.json"},
		{"HugoHistoryPath", cfg.HugoHistoryPath, "./data/history"},
		{"APIBaseURL", cfg.APIBaseURL, "https://api.coingecko.com/api/v3"},
		{"APITimeout", cfg.APITimeout, 30 * time.Second},
		{"Timeout", cfg.Timeout, 5 * time.Minute},
		{"HistoryDays", cfg.HistoryDays, 30},
		{"MaxRetries", cfg.MaxRetries, 3},
		{"RetryBackoff", cfg.RetryBackoff, time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	t.Setenv("CRYPTO_BOT_DB_PATH", "/custom/path.db")
	t.Setenv("CRYPTO_BOT_TIMEOUT", "10m")
	t.Setenv("CRYPTO_BOT_HISTORY_DAYS", "60")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DatabasePath != "/custom/path.db" {
		t.Errorf("DatabasePath = %v, want /custom/path.db", cfg.DatabasePath)
	}

	if cfg.Timeout != 10*time.Minute {
		t.Errorf("Timeout = %v, want 10m", cfg.Timeout)
	}

	if cfg.HistoryDays != 60 {
		t.Errorf("HistoryDays = %v, want 60", cfg.HistoryDays)
	}
}

func TestLoad_InvalidDuration(t *testing.T) {
	t.Setenv("CRYPTO_BOT_TIMEOUT", "invalid")

	_, err := config.Load()
	if err == nil {
		t.Error("Load() expected error for invalid duration, got nil")
	}
}

func TestLoad_InvalidInt(t *testing.T) {
	t.Setenv("CRYPTO_BOT_HISTORY_DAYS", "not-a-number")

	_, err := config.Load()
	if err == nil {
		t.Error("Load() expected error for invalid int, got nil")
	}
}

func TestLoad_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		envKey  string
		envVal  string
		wantErr bool
	}{
		{"empty database path", "CRYPTO_BOT_DB_PATH", "", false}, // empty string uses default
		{"negative history days", "CRYPTO_BOT_HISTORY_DAYS", "-5", true},
		{"zero timeout", "CRYPTO_BOT_TIMEOUT", "0s", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envVal)
			}

			_, err := config.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
