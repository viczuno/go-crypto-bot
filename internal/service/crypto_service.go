// Package service provides business logic orchestration for the crypto bot.
package service

import (
	"context"
	"fmt"
	"log"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// CryptoService coordinates fetching, storing, and reporting crypto prices.
type CryptoService struct {
	fetcher   domain.PriceFetcher
	repo      domain.PriceRepository
	generator domain.ReadmeGenerator
}

// NewCryptoService creates a new crypto service.
func NewCryptoService(
	fetcher domain.PriceFetcher,
	repo domain.PriceRepository,
	generator domain.ReadmeGenerator,
) *CryptoService {
	return &CryptoService{
		fetcher:   fetcher,
		repo:      repo,
		generator: generator,
	}
}

// UpdateAndGenerateReport fetches latest prices, stores them, and generates a report.
func (s *CryptoService) UpdateAndGenerateReport(ctx context.Context, coins []domain.CoinMetadata) (string, []domain.CoinStats, error) {
	coinIDs := make([]string, len(coins))
	for i, c := range coins {
		coinIDs[i] = c.ID
	}

	log.Println("Fetching latest prices from API...")
	prices, err := s.fetcher.FetchPrices(ctx, coinIDs)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	log.Printf("Successfully fetched prices for %d coins", len(prices))

	log.Println("Saving prices to database...")
	if err := s.repo.SavePrices(ctx, prices); err != nil {
		return "", nil, fmt.Errorf("failed to save prices: %w", err)
	}
	log.Println("Prices saved successfully")

	stats := s.buildStats(ctx, coins, prices)

	log.Println("Generating README...")
	content := s.generator.Generate(stats, coins)

	return content, stats, nil
}

func (s *CryptoService) buildStats(ctx context.Context, coins []domain.CoinMetadata, prices map[string]domain.CryptoPrice) []domain.CoinStats {
	stats := make([]domain.CoinStats, 0, len(coins))

	for _, coin := range coins {
		price, ok := prices[coin.ID]
		if !ok {
			log.Printf("No price data for %s", coin.ID)
			continue
		}

		stat := domain.CoinStats{
			ID:        coin.ID,
			Symbol:    coin.Symbol,
			Price:     price.PriceUSD,
			Change24h: price.Change24h,
			Change7d:  s.getHistoricalChange(ctx, coin.ID, price.PriceUSD, domain.Days7),
			Change30d: s.getHistoricalChange(ctx, coin.ID, price.PriceUSD, domain.Days30),
		}

		stats = append(stats, stat)
	}

	return stats
}

func (s *CryptoService) getHistoricalChange(ctx context.Context, coinID string, currentPrice float64, days int) domain.PriceChange {
	result, err := s.repo.GetHistoricalPrice(ctx, coinID, days)
	if err != nil {
		log.Printf("Error getting %d-day history for %s: %v", days, coinID, err)
		return domain.PriceChange{HasData: false, Days: days}
	}

	if !result.Found {
		return domain.PriceChange{HasData: false, Days: days}
	}

	// Use the shared calculation function from domain
	return domain.CalculatePriceChange(currentPrice, result.Price, days)
}
