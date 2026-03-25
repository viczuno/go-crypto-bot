package coins

import (
	"context"
	"errors"
	"testing"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

type mockCoinSource struct {
	coins []domain.CoinMetadata
	err   error
	typ   string
}

func (m *mockCoinSource) FetchCoins(ctx context.Context) ([]domain.CoinMetadata, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.coins, nil
}

func (m *mockCoinSource) Type() string {
	return m.typ
}

func TestAggregator(t *testing.T) {
	tests := []struct {
		name      string
		sources   []domain.CoinSource
		wantCount int
		wantErr   bool
	}{
		{
			name: "merges coins from multiple sources",
			sources: []domain.CoinSource{
				&mockCoinSource{
					coins: []domain.CoinMetadata{
						{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC"},
						{ID: "ethereum", Name: "Ethereum", Symbol: "ETH"},
					},
					typ: "source1",
				},
				&mockCoinSource{
					coins: []domain.CoinMetadata{
						{ID: "solana", Name: "Solana", Symbol: "SOL"},
					},
					typ: "source2",
				},
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "deduplicates coins from multiple sources",
			sources: []domain.CoinSource{
				&mockCoinSource{
					coins: []domain.CoinMetadata{
						{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC", Tags: []string{"layer1"}},
					},
					typ: "source1",
				},
				&mockCoinSource{
					coins: []domain.CoinMetadata{
						{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC", Tags: []string{"top-5"}},
					},
					typ: "source2",
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "continues on source error",
			sources: []domain.CoinSource{
				&mockCoinSource{
					err: errors.New("api error"),
					typ: "failing-source",
				},
				&mockCoinSource{
					coins: []domain.CoinMetadata{
						{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC"},
					},
					typ: "working-source",
				},
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "errors when no sources configured",
			sources:   []domain.CoinSource{},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "errors when all sources fail",
			sources: []domain.CoinSource{
				&mockCoinSource{err: errors.New("error1"), typ: "source1"},
				&mockCoinSource{err: errors.New("error2"), typ: "source2"},
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := NewAggregator(tt.sources...)

			if got := agg.Type(); got != "aggregator" {
				t.Errorf("Type() = %v, want aggregator", got)
			}

			ctx := context.Background()
			coins, err := agg.FetchCoins(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("FetchCoins() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(coins) != tt.wantCount {
				t.Errorf("FetchCoins() returned %d coins, want %d", len(coins), tt.wantCount)
			}
		})
	}
}

func TestMergeCoinMetadata(t *testing.T) {
	tests := []struct {
		name     string
		existing domain.CoinMetadata
		new      domain.CoinMetadata
		want     domain.CoinMetadata
	}{
		{
			name: "merges tags from both coins",
			existing: domain.CoinMetadata{
				ID:   "bitcoin",
				Tags: []string{"layer1"},
			},
			new: domain.CoinMetadata{
				ID:   "bitcoin",
				Tags: []string{"top-5"},
			},
			want: domain.CoinMetadata{
				ID:   "bitcoin",
				Tags: []string{"layer1", "top-5"},
			},
		},
		{
			name: "prefers non-empty values",
			existing: domain.CoinMetadata{
				ID:     "bitcoin",
				Name:   "Bitcoin",
				Symbol: "",
			},
			new: domain.CoinMetadata{
				ID:     "bitcoin",
				Name:   "",
				Symbol: "BTC",
			},
			want: domain.CoinMetadata{
				ID:     "bitcoin",
				Name:   "Bitcoin",
				Symbol: "BTC",
			},
		},
		{
			name: "prefers positive market cap and rank",
			existing: domain.CoinMetadata{
				ID:        "bitcoin",
				MarketCap: 0,
				Rank:      0,
			},
			new: domain.CoinMetadata{
				ID:        "bitcoin",
				MarketCap: 1000000,
				Rank:      1,
			},
			want: domain.CoinMetadata{
				ID:        "bitcoin",
				MarketCap: 1000000,
				Rank:      1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeCoinMetadata(tt.existing, tt.new)

			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Symbol != tt.want.Symbol {
				t.Errorf("Symbol = %v, want %v", got.Symbol, tt.want.Symbol)
			}
			if got.MarketCap != tt.want.MarketCap {
				t.Errorf("MarketCap = %v, want %v", got.MarketCap, tt.want.MarketCap)
			}
			if got.Rank != tt.want.Rank {
				t.Errorf("Rank = %v, want %v", got.Rank, tt.want.Rank)
			}
			if len(got.Tags) != len(tt.want.Tags) {
				t.Errorf("Tags length = %v, want %v", len(got.Tags), len(tt.want.Tags))
			}
		})
	}
}
