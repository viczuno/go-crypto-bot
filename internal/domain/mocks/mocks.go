// Package mocks provides mock implementations of domain interfaces for testing.
package mocks

import (
	"context"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// MockPriceFetcher is a mock implementation of domain.PriceFetcher.
type MockPriceFetcher struct {
	FetchPricesFunc func(ctx context.Context, coinIDs []string) (map[string]domain.CryptoPrice, error)
}

// FetchPrices calls the mock function.
func (m *MockPriceFetcher) FetchPrices(ctx context.Context, coinIDs []string) (map[string]domain.CryptoPrice, error) {
	if m.FetchPricesFunc != nil {
		return m.FetchPricesFunc(ctx, coinIDs)
	}
	return nil, nil
}

// MockHistoricalPriceFetcher is a mock implementation of domain.HistoricalPriceFetcher.
type MockHistoricalPriceFetcher struct {
	FetchHistoricalPricesFunc func(ctx context.Context, coinID string, days int) ([]domain.CryptoPrice, error)
}

// FetchHistoricalPrices calls the mock function.
func (m *MockHistoricalPriceFetcher) FetchHistoricalPrices(ctx context.Context, coinID string, days int) ([]domain.CryptoPrice, error) {
	if m.FetchHistoricalPricesFunc != nil {
		return m.FetchHistoricalPricesFunc(ctx, coinID, days)
	}
	return nil, nil
}

// MockPriceRepository is a mock implementation of domain.PriceRepository.
type MockPriceRepository struct {
	SavePricesFunc         func(ctx context.Context, prices map[string]domain.CryptoPrice) error
	GetHistoricalPriceFunc func(ctx context.Context, coinID string, daysAgo int) (domain.PriceResult, error)
	GetPriceHistoryFunc    func(ctx context.Context, coinID string, days int) ([]domain.CryptoPrice, error)
}

// SavePrices calls the mock function.
func (m *MockPriceRepository) SavePrices(ctx context.Context, prices map[string]domain.CryptoPrice) error {
	if m.SavePricesFunc != nil {
		return m.SavePricesFunc(ctx, prices)
	}
	return nil
}

// GetHistoricalPrice calls the mock function.
func (m *MockPriceRepository) GetHistoricalPrice(ctx context.Context, coinID string, daysAgo int) (domain.PriceResult, error) {
	if m.GetHistoricalPriceFunc != nil {
		return m.GetHistoricalPriceFunc(ctx, coinID, daysAgo)
	}
	return domain.PriceResult{}, nil
}

// GetPriceHistory calls the mock function.
func (m *MockPriceRepository) GetPriceHistory(ctx context.Context, coinID string, days int) ([]domain.CryptoPrice, error) {
	if m.GetPriceHistoryFunc != nil {
		return m.GetPriceHistoryFunc(ctx, coinID, days)
	}
	return nil, nil
}

// MockReadmeGenerator is a mock implementation of domain.ReadmeGenerator.
type MockReadmeGenerator struct {
	GenerateFunc func(stats []domain.CoinStats, coins []domain.CoinMetadata) string
}

// Generate calls the mock function.
func (m *MockReadmeGenerator) Generate(stats []domain.CoinStats, coins []domain.CoinMetadata) string {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(stats, coins)
	}
	return ""
}

// Ensure mocks implement interfaces.
var (
	_ domain.PriceFetcher           = (*MockPriceFetcher)(nil)
	_ domain.HistoricalPriceFetcher = (*MockHistoricalPriceFetcher)(nil)
	_ domain.PriceRepository        = (*MockPriceRepository)(nil)
	_ domain.ReadmeGenerator        = (*MockReadmeGenerator)(nil)
)
