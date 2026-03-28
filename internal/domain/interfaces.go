// Package domain contains the core business types and interfaces for the crypto bot.
package domain

import "context"

// PriceResult represents the result of a historical price lookup.
// Using a struct instead of tuple (float64, bool, error) for clarity.
type PriceResult struct {
	Price float64
	Found bool
}

// PriceFetcher defines the interface for fetching cryptocurrency prices.
type PriceFetcher interface {
	FetchPrices(ctx context.Context, coinIDs []string) (map[string]CryptoPrice, error)
}

// HistoricalPriceFetcher defines the interface for fetching historical price data from an API.
type HistoricalPriceFetcher interface {
	FetchHistoricalPrices(ctx context.Context, coinID string, days int) ([]CryptoPrice, error)
}

// PriceRepository defines the interface for storing and retrieving price data.
type PriceRepository interface {
	SavePrices(ctx context.Context, prices map[string]CryptoPrice) error
	GetHistoricalPrice(ctx context.Context, coinID string, daysAgo int) (PriceResult, error)
	GetPriceHistory(ctx context.Context, coinID string, days int) ([]CryptoPrice, error)
}

// Closer defines the interface for resources that need cleanup.
type Closer interface {
	Close() error
}

// ReadmeGenerator defines the interface for generating README content.
type ReadmeGenerator interface {
	Generate(stats []CoinStats, coins []CoinMetadata) string
}

// CoinSource defines the interface for fetching coin metadata from various sources.
type CoinSource interface {
	FetchCoins(ctx context.Context) ([]CoinMetadata, error)
	Type() string
}

// CalculatePriceChange calculates price changes between two prices.
func CalculatePriceChange(current, past float64, days int) PriceChange {
	if past == 0 {
		return PriceChange{HasData: false, Days: days}
	}

	absChange := current - past
	pctChange := (absChange / past) * 100.0

	return PriceChange{
		PastPrice:    past,
		CurrentPrice: current,
		AbsChange:    absChange,
		PctChange:    pctChange,
		HasData:      true,
		Days:         days,
	}
}
