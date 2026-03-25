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

// MarketCapSource fetches top coins by market cap from CoinGecko.
type MarketCapSource struct {
	baseURL    string
	httpClient *http.Client
	count      int
	exclude    []string
}

type coinGeckoMarketData struct {
	ID              string  `json:"id"`
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	MarketCap       float64 `json:"market_cap"`
	MarketCapRank   int     `json:"market_cap_rank"`
	CurrentPrice    float64 `json:"current_price"`
	PriceChange24h  float64 `json:"price_change_percentage_24h"`
}

// NewMarketCapSource creates a new market cap-based coin source.
func NewMarketCapSource(baseURL string, count int, exclude []string) *MarketCapSource {
	return &MarketCapSource{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		count:      count,
		exclude:    exclude,
	}
}

// FetchCoins fetches top N coins by market cap from CoinGecko.
func (m *MarketCapSource) FetchCoins(ctx context.Context) ([]domain.CoinMetadata, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1&sparkline=false",
		m.baseURL, m.count+len(m.exclude))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var marketData []coinGeckoMarketData
	if err := json.NewDecoder(resp.Body).Decode(&marketData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	excludeMap := make(map[string]bool)
	for _, id := range m.exclude {
		excludeMap[strings.ToLower(id)] = true
	}

	coins := make([]domain.CoinMetadata, 0, m.count)
	for _, data := range marketData {
		if excludeMap[strings.ToLower(data.ID)] {
			continue
		}

		if len(coins) >= m.count {
			break
		}

		coins = append(coins, domain.CoinMetadata{
			ID:        data.ID,
			Name:      data.Name,
			Symbol:    strings.ToUpper(data.Symbol),
			Tags:      []string{"top-market-cap"},
			MarketCap: data.MarketCap,
			Rank:      data.MarketCapRank,
		})
	}

	return coins, nil
}

// Type returns the source type identifier.
func (m *MarketCapSource) Type() string {
	return "top_by_market_cap"
}

var _ domain.CoinSource = (*MarketCapSource)(nil)
