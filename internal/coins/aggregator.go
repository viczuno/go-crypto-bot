package coins

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// Aggregator merges coins from multiple sources and handles deduplication.
type Aggregator struct {
	sources []domain.CoinSource
}

// NewAggregator creates a new coin aggregator.
func NewAggregator(sources ...domain.CoinSource) *Aggregator {
	return &Aggregator{sources: sources}
}

// FetchCoins fetches coins from all sources, merges them, and deduplicates.
func (a *Aggregator) FetchCoins(ctx context.Context) ([]domain.CoinMetadata, error) {
	if len(a.sources) == 0 {
		return nil, fmt.Errorf("no coin sources configured")
	}

	coinMap := make(map[string]domain.CoinMetadata)
	var allCoins []domain.CoinMetadata

	for _, source := range a.sources {
		log.Printf("Fetching coins from %s source...", source.Type())

		coins, err := source.FetchCoins(ctx)
		if err != nil {
			log.Printf("Warning: failed to fetch coins from %s source: %v", source.Type(), err)
			continue
		}

		log.Printf("Fetched %d coins from %s source", len(coins), source.Type())

		for _, coin := range coins {
			id := strings.ToLower(coin.ID)

			if existing, found := coinMap[id]; found {
				coinMap[id] = mergeCoinMetadata(existing, coin)
			} else {
				coinMap[id] = coin
			}
		}
	}

	for _, coin := range coinMap {
		allCoins = append(allCoins, coin)
	}

	if len(allCoins) == 0 {
		return nil, fmt.Errorf("no coins fetched from any source")
	}

	log.Printf("Total unique coins after aggregation: %d", len(allCoins))
	return allCoins, nil
}

// Type returns the source type identifier.
func (a *Aggregator) Type() string {
	return "aggregator"
}

// mergeCoinMetadata merges two coin metadata entries, preferring non-empty values.
func mergeCoinMetadata(existing, new domain.CoinMetadata) domain.CoinMetadata {
	merged := existing

	if new.Name != "" {
		merged.Name = new.Name
	}
	if new.Symbol != "" {
		merged.Symbol = new.Symbol
	}
	if new.MarketCap > 0 {
		merged.MarketCap = new.MarketCap
	}
	if new.Rank > 0 {
		merged.Rank = new.Rank
	}

	return merged
}

var _ domain.CoinSource = (*Aggregator)(nil)
