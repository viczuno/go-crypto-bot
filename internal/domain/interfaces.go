package domain

import "context"

// PriceFetcher defines the interface for fetching cryptocurrency prices
type PriceFetcher interface {
	FetchPrices(ctx context.Context, coinIDs []string) (map[string]CryptoPrice, error)
}

// PriceRepository defines the interface for storing and retrieving price data
type PriceRepository interface {
	SavePrices(prices map[string]CryptoPrice) error
	GetHistoricalPrice(coinID string, daysAgo int) (float64, bool, error)
	GetPriceHistory(coinID string, days int) ([]CryptoPrice, error)
	Close() error
}

// ReadmeGenerator defines the interface for generating README content
type ReadmeGenerator interface {
	Generate(stats []CoinStats, coins []CoinMetadata) string
}
