package coins

import (
	"context"
	"testing"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

func TestStaticSource(t *testing.T) {
	tests := []struct {
		name      string
		coins     []domain.CoinMetadata
		wantCount int
	}{
		{
			name: "returns predefined coins",
			coins: []domain.CoinMetadata{
				{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC", Tags: []string{"top-5"}},
				{ID: "ethereum", Name: "Ethereum", Symbol: "ETH", Tags: []string{"top-5"}},
			},
			wantCount: 2,
		},
		{
			name:      "returns empty list when no coins",
			coins:     []domain.CoinMetadata{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewStaticSource(tt.coins)

			if got := source.Type(); got != "static" {
				t.Errorf("Type() = %v, want static", got)
			}

			ctx := context.Background()
			coins, err := source.FetchCoins(ctx)
			if err != nil {
				t.Fatalf("FetchCoins() error = %v", err)
			}

			if len(coins) != tt.wantCount {
				t.Errorf("FetchCoins() returned %d coins, want %d", len(coins), tt.wantCount)
			}

			for i, coin := range coins {
				if coin.ID != tt.coins[i].ID {
					t.Errorf("coin[%d].ID = %v, want %v", i, coin.ID, tt.coins[i].ID)
				}
			}
		})
	}
}
