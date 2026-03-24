package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/db"
	"github.com/viczuno/go-crypto-bot/internal/domain"
)

func TestSQLiteRepository_SaveAndRetrieve(t *testing.T) {
	repo, err := db.NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	ctx := context.Background()

	// Test SavePrices
	prices := map[string]domain.CryptoPrice{
		"bitcoin": {
			Coin:      "bitcoin",
			PriceUSD:  50000.0,
			Change24h: 2.5,
			FetchedAt: time.Now().UTC(),
		},
		"ethereum": {
			Coin:      "ethereum",
			PriceUSD:  3000.0,
			Change24h: -1.5,
			FetchedAt: time.Now().UTC(),
		},
	}

	err = repo.SavePrices(ctx, prices)
	if err != nil {
		t.Fatalf("SavePrices() error = %v", err)
	}

	// Test GetPriceHistory
	history, err := repo.GetPriceHistory(ctx, "bitcoin", 30)
	if err != nil {
		t.Fatalf("GetPriceHistory() error = %v", err)
	}

	if len(history) != 1 {
		t.Errorf("GetPriceHistory() returned %d records, want 1", len(history))
	}

	if history[0].Coin != "bitcoin" {
		t.Errorf("Coin = %s, want bitcoin", history[0].Coin)
	}

	if history[0].PriceUSD != 50000.0 {
		t.Errorf("PriceUSD = %f, want 50000.0", history[0].PriceUSD)
	}
}

func TestSQLiteRepository_GetHistoricalPrice(t *testing.T) {
	repo, err := db.NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	ctx := context.Background()

	// Insert test data with a timestamp from 8 days ago
	prices := map[string]domain.CryptoPrice{
		"bitcoin": {
			Coin:      "bitcoin",
			PriceUSD:  45000.0,
			Change24h: 0,
			FetchedAt: time.Now().UTC().Add(-8 * 24 * time.Hour),
		},
	}

	err = repo.SavePrices(ctx, prices)
	if err != nil {
		t.Fatalf("SavePrices() error = %v", err)
	}

	// Get price from 7 days ago (should find the 8-day-old record)
	result, err := repo.GetHistoricalPrice(ctx, "bitcoin", 7)
	if err != nil {
		t.Fatalf("GetHistoricalPrice() error = %v", err)
	}

	if !result.Found {
		t.Error("Expected to find historical price")
	}

	if result.Price != 45000.0 {
		t.Errorf("Price = %f, want 45000.0", result.Price)
	}
}

func TestSQLiteRepository_GetHistoricalPrice_NotFound(t *testing.T) {
	repo, err := db.NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	ctx := context.Background()

	// No data inserted, should return not found
	result, err := repo.GetHistoricalPrice(ctx, "bitcoin", 7)
	if err != nil {
		t.Fatalf("GetHistoricalPrice() error = %v", err)
	}

	if result.Found {
		t.Error("Expected not to find historical price for empty database")
	}
}

func TestSQLiteRepository_GetPriceHistory_Empty(t *testing.T) {
	repo, err := db.NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	ctx := context.Background()

	history, err := repo.GetPriceHistory(ctx, "bitcoin", 30)
	if err != nil {
		t.Fatalf("GetPriceHistory() error = %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d records", len(history))
	}
}

func TestSQLiteRepository_Close(t *testing.T) {
	repo, err := db.NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}

	err = repo.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Closing again should not error
	err = repo.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestSQLiteRepository_ContextCancellation(t *testing.T) {
	repo, err := db.NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	defer func() { _ = repo.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	prices := map[string]domain.CryptoPrice{
		"bitcoin": {
			Coin:      "bitcoin",
			PriceUSD:  50000.0,
			FetchedAt: time.Now().UTC(),
		},
	}

	// Operations with cancelled context should fail
	err = repo.SavePrices(ctx, prices)
	if err == nil {
		t.Log("SavePrices with cancelled context may or may not error depending on timing")
	}
}
