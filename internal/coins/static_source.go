// Package coins provides coin discovery and configuration functionality.
package coins

import (
	"context"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// StaticSource returns a predefined list of coins.
type StaticSource struct {
	coins []domain.CoinMetadata
}

// NewStaticSource creates a new static coin source.
func NewStaticSource(coins []domain.CoinMetadata) *StaticSource {
	return &StaticSource{coins: coins}
}

// FetchCoins returns the predefined list of coins.
func (s *StaticSource) FetchCoins(ctx context.Context) ([]domain.CoinMetadata, error) {
	return s.coins, nil
}

// Type returns the source type identifier.
func (s *StaticSource) Type() string {
	return "static"
}

var _ domain.CoinSource = (*StaticSource)(nil)
