package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

const (
	baseURL        = "https://api.coingecko.com/api/v3"
	defaultTimeout = 30 * time.Second
)

// CoinGeckoClient implements domain.PriceFetcher for the CoinGecko API
type CoinGeckoClient struct {
	httpClient *http.Client
	baseURL    string
}

// coinGeckoResponse represents the API response structure
type coinGeckoResponse map[string]struct {
	USD          float64 `json:"usd"`
	USD24hChange float64 `json:"usd_24h_change"`
}

// NewCoinGeckoClient creates a new CoinGecko API client
func NewCoinGeckoClient() *CoinGeckoClient {
	return &CoinGeckoClient{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: baseURL,
	}
}

// FetchPrices retrieves current prices for the specified coins
func (c *CoinGeckoClient) FetchPrices(ctx context.Context, coinIDs []string) (map[string]domain.CryptoPrice, error) {
	if len(coinIDs) == 0 {
		return nil, fmt.Errorf("no coin IDs provided")
	}

	ids := strings.Join(coinIDs, ",")
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd&include_24hr_change=true", c.baseURL, ids)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var apiResponse coinGeckoResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := make(map[string]domain.CryptoPrice, len(apiResponse))
	now := time.Now().UTC()

	for coinID, data := range apiResponse {
		result[coinID] = domain.CryptoPrice{
			Coin:      coinID,
			PriceUSD:  data.USD,
			Change24h: data.USD24hChange,
			FetchedAt: now,
		}
	}

	return result, nil
}

// Ensure CoinGeckoClient implements PriceFetcher
var _ domain.PriceFetcher = (*CoinGeckoClient)(nil)
