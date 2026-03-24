package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
	"github.com/viczuno/go-crypto-bot/internal/domain/mocks"
	"github.com/viczuno/go-crypto-bot/internal/service"
)

func TestCryptoService_UpdateAndGenerateReport(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		coins        []domain.CoinMetadata
		setupFetcher func() *mocks.MockPriceFetcher
		setupRepo    func() *mocks.MockPriceRepository
		setupGen     func() *mocks.MockReadmeGenerator
		wantErr      bool
		errContains  string
		wantStatsLen int
	}{
		{
			name: "successful update",
			coins: []domain.CoinMetadata{
				{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC"},
			},
			setupFetcher: func() *mocks.MockPriceFetcher {
				return &mocks.MockPriceFetcher{
					FetchPricesFunc: func(ctx context.Context, coinIDs []string) (map[string]domain.CryptoPrice, error) {
						return map[string]domain.CryptoPrice{
							"bitcoin": {
								Coin:      "bitcoin",
								PriceUSD:  50000,
								Change24h: 2.5,
								FetchedAt: time.Now(),
							},
						}, nil
					},
				}
			},
			setupRepo: func() *mocks.MockPriceRepository {
				return &mocks.MockPriceRepository{
					SavePricesFunc: func(ctx context.Context, prices map[string]domain.CryptoPrice) error {
						return nil
					},
					GetHistoricalPriceFunc: func(ctx context.Context, coinID string, daysAgo int) (domain.PriceResult, error) {
						return domain.PriceResult{Price: 48000, Found: true}, nil
					},
				}
			},
			setupGen: func() *mocks.MockReadmeGenerator {
				return &mocks.MockReadmeGenerator{
					GenerateFunc: func(stats []domain.CoinStats, coins []domain.CoinMetadata) string {
						return "# Generated README"
					},
				}
			},
			wantErr:      false,
			wantStatsLen: 1,
		},
		{
			name: "fetcher error",
			coins: []domain.CoinMetadata{
				{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC"},
			},
			setupFetcher: func() *mocks.MockPriceFetcher {
				return &mocks.MockPriceFetcher{
					FetchPricesFunc: func(ctx context.Context, coinIDs []string) (map[string]domain.CryptoPrice, error) {
						return nil, errors.New("network error")
					},
				}
			},
			setupRepo: func() *mocks.MockPriceRepository {
				return &mocks.MockPriceRepository{}
			},
			setupGen: func() *mocks.MockReadmeGenerator {
				return &mocks.MockReadmeGenerator{}
			},
			wantErr:     true,
			errContains: "fetch prices",
		},
		{
			name: "save error",
			coins: []domain.CoinMetadata{
				{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC"},
			},
			setupFetcher: func() *mocks.MockPriceFetcher {
				return &mocks.MockPriceFetcher{
					FetchPricesFunc: func(ctx context.Context, coinIDs []string) (map[string]domain.CryptoPrice, error) {
						return map[string]domain.CryptoPrice{
							"bitcoin": {Coin: "bitcoin", PriceUSD: 50000},
						}, nil
					},
				}
			},
			setupRepo: func() *mocks.MockPriceRepository {
				return &mocks.MockPriceRepository{
					SavePricesFunc: func(ctx context.Context, prices map[string]domain.CryptoPrice) error {
						return errors.New("database error")
					},
				}
			},
			setupGen: func() *mocks.MockReadmeGenerator {
				return &mocks.MockReadmeGenerator{}
			},
			wantErr:     true,
			errContains: "save prices",
		},
		{
			name: "no historical data",
			coins: []domain.CoinMetadata{
				{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC"},
			},
			setupFetcher: func() *mocks.MockPriceFetcher {
				return &mocks.MockPriceFetcher{
					FetchPricesFunc: func(ctx context.Context, coinIDs []string) (map[string]domain.CryptoPrice, error) {
						return map[string]domain.CryptoPrice{
							"bitcoin": {Coin: "bitcoin", PriceUSD: 50000, Change24h: 1.5},
						}, nil
					},
				}
			},
			setupRepo: func() *mocks.MockPriceRepository {
				return &mocks.MockPriceRepository{
					SavePricesFunc: func(ctx context.Context, prices map[string]domain.CryptoPrice) error {
						return nil
					},
					GetHistoricalPriceFunc: func(ctx context.Context, coinID string, daysAgo int) (domain.PriceResult, error) {
						return domain.PriceResult{Found: false}, nil
					},
				}
			},
			setupGen: func() *mocks.MockReadmeGenerator {
				return &mocks.MockReadmeGenerator{
					GenerateFunc: func(stats []domain.CoinStats, coins []domain.CoinMetadata) string {
						return "# README"
					},
				}
			},
			wantErr:      false,
			wantStatsLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := tt.setupFetcher()
			repo := tt.setupRepo()
			gen := tt.setupGen()

			svc := service.NewCryptoService(fetcher, repo, gen)

			content, stats, err := svc.UpdateAndGenerateReport(ctx, tt.coins)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !containsSubstring(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if content == "" {
				t.Error("expected non-empty content")
			}

			if len(stats) != tt.wantStatsLen {
				t.Errorf("stats length = %d, want %d", len(stats), tt.wantStatsLen)
			}
		})
	}
}

func TestCalculatePriceChange(t *testing.T) {
	tests := []struct {
		name        string
		current     float64
		past        float64
		days        int
		wantHasData bool
		wantPct     float64
	}{
		{
			name:        "positive change",
			current:     110,
			past:        100,
			days:        7,
			wantHasData: true,
			wantPct:     10.0,
		},
		{
			name:        "negative change",
			current:     90,
			past:        100,
			days:        7,
			wantHasData: true,
			wantPct:     -10.0,
		},
		{
			name:        "zero past price",
			current:     100,
			past:        0,
			days:        7,
			wantHasData: false,
			wantPct:     0,
		},
		{
			name:        "no change",
			current:     100,
			past:        100,
			days:        30,
			wantHasData: true,
			wantPct:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.CalculatePriceChange(tt.current, tt.past, tt.days)

			if result.HasData != tt.wantHasData {
				t.Errorf("HasData = %v, want %v", result.HasData, tt.wantHasData)
			}

			if result.HasData && result.PctChange != tt.wantPct {
				t.Errorf("PctChange = %v, want %v", result.PctChange, tt.wantPct)
			}

			if result.Days != tt.days {
				t.Errorf("Days = %v, want %v", result.Days, tt.days)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
