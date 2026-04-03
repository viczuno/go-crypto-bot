package coins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// TrendingSource fetches trending coins from CoinGecko.
type TrendingSource struct {
	baseURL    string
	httpClient *http.Client
	count      int
}

type trendingResponse struct {
	Coins []struct {
		Item struct {
			ID        string `json:"id"`
			Symbol    string `json:"symbol"`
			Name      string `json:"name"`
			MarketCap int    `json:"market_cap_rank"`
		} `json:"item"`
	} `json:"coins"`
}

// NewTrendingSource creates a new trending coin source.
func NewTrendingSource(baseURL string, count int) *TrendingSource {
	return &TrendingSource{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		count:      count,
	}
}

// FetchCoins fetches trending coins from CoinGecko.
func (t *TrendingSource) FetchCoins(ctx context.Context) ([]domain.CoinMetadata, error) {
	url := fmt.Sprintf("%s/search/trending", t.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trending data: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var trendingData trendingResponse
	if err := json.NewDecoder(resp.Body).Decode(&trendingData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	maxCoins := t.count
	if len(trendingData.Coins) < maxCoins {
		maxCoins = len(trendingData.Coins)
	}

	coins := make([]domain.CoinMetadata, 0, maxCoins)
	for i := 0; i < maxCoins; i++ {
		item := trendingData.Coins[i].Item
		coins = append(coins, domain.CoinMetadata{
			ID:     item.ID,
			Name:   item.Name,
			Symbol: strings.ToUpper(item.Symbol),
			Rank:   item.MarketCap,
		})
	}

	return coins, nil
}

// Type returns the source type identifier.
func (t *TrendingSource) Type() string {
	return "trending"
}

var _ domain.CoinSource = (*TrendingSource)(nil)
