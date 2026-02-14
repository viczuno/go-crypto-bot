package service

import (
	"context"
	"fmt"
	"log"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// CryptoService coordinates fetching, storing, and reporting crypto prices
type CryptoService struct {
	fetcher   domain.PriceFetcher
	repo      domain.PriceRepository
	generator domain.ReadmeGenerator
}

// NewCryptoService creates a new crypto service
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

// UpdateAndGenerateReport fetches latest prices, stores them, and generates a report
func (s *CryptoService) UpdateAndGenerateReport(ctx context.Context, coins []domain.CoinMetadata) (string, error) {
	coinIDs := make([]string, len(coins))
	for i, c := range coins {
		coinIDs[i] = c.ID
	}

	// Fetch latest prices
	log.Println("Fetching latest prices from API...")
	prices, err := s.fetcher.FetchPrices(ctx, coinIDs)
	if err != nil {
		return "", fmt.Errorf("failed to fetch prices: %w", err)
	}
	log.Printf("Successfully fetched prices for %d coins", len(prices))

	// Save to database
	log.Println("Saving prices to database...")
	if err := s.repo.SavePrices(prices); err != nil {
		return "", fmt.Errorf("failed to save prices: %w", err)
	}
	log.Println("Prices saved successfully")

	// Build stats with historical data
	stats := s.buildStats(coins, prices)

	// Generate README
	log.Println("Generating README...")
	content := s.generator.Generate(stats, coins)

	return content, nil
}

func (s *CryptoService) buildStats(coins []domain.CoinMetadata, prices map[string]domain.CryptoPrice) []domain.CoinStats {
	stats := make([]domain.CoinStats, 0, len(coins))

	for _, coin := range coins {
		price, ok := prices[coin.ID]
		if !ok {
			log.Printf("No price data for %s", coin.ID)
			continue
		}

		stat := domain.CoinStats{
			Name:      coin.ID,
			Symbol:    coin.Symbol,
			Price:     price.PriceUSD,
			Change24h: price.Change24h,
			Change7d:  s.getHistoricalChange(coin.ID, price.PriceUSD, 7),
			Change30d: s.getHistoricalChange(coin.ID, price.PriceUSD, 30),
		}

		stats = append(stats, stat)
	}

	return stats
}

func (s *CryptoService) getHistoricalChange(coinID string, currentPrice float64, days int) domain.PriceChange {
	pastPrice, hasData, err := s.repo.GetHistoricalPrice(coinID, days)
	if err != nil {
		log.Printf("Error getting %d-day history for %s: %v", days, coinID, err)
		return domain.PriceChange{HasData: false, Days: days}
	}

	if !hasData {
		return domain.PriceChange{HasData: false, Days: days}
	}

	absChange := currentPrice - pastPrice
	pctChange := (absChange / pastPrice) * 100.0

	return domain.PriceChange{
		PastPrice:    pastPrice,
		CurrentPrice: currentPrice,
		AbsChange:    absChange,
		PctChange:    pctChange,
		HasData:      true,
		Days:         days,
	}
}

// Close cleans up service resources
func (s *CryptoService) Close() error {
	return s.repo.Close()
}
