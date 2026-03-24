// Package api provides external API clients for the crypto bot.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
	apperrors "github.com/viczuno/go-crypto-bot/internal/errors"
)

const (
	defaultBaseURL = "https://api.coingecko.com/api/v3"
	defaultTimeout = 30 * time.Second
	defaultRetries = 3
	defaultBackoff = time.Second
)

// CoinGeckoClient implements domain.PriceFetcher for the CoinGecko API.
type CoinGeckoClient struct {
	httpClient      *http.Client
	baseURL         string
	maxRetries      int
	backoffDuration time.Duration
}

// ClientOption configures a CoinGeckoClient.
type ClientOption func(*CoinGeckoClient)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *CoinGeckoClient) {
		c.httpClient.Timeout = d
	}
}

// WithBaseURL sets the API base URL.
func WithBaseURL(url string) ClientOption {
	return func(c *CoinGeckoClient) {
		c.baseURL = url
	}
}

// WithRetry configures retry behavior.
func WithRetry(maxRetries int, backoff time.Duration) ClientOption {
	return func(c *CoinGeckoClient) {
		c.maxRetries = maxRetries
		c.backoffDuration = backoff
	}
}

// NewCoinGeckoClient creates a new CoinGecko API client with optional configuration.
func NewCoinGeckoClient(opts ...ClientOption) *CoinGeckoClient {
	c := &CoinGeckoClient{
		httpClient:      &http.Client{Timeout: defaultTimeout},
		baseURL:         defaultBaseURL,
		maxRetries:      defaultRetries,
		backoffDuration: defaultBackoff,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type coinGeckoResponse map[string]struct {
	USD          float64 `json:"usd"`
	USD24hChange float64 `json:"usd_24h_change"`
}

// FetchPrices retrieves current prices for the specified coins.
func (c *CoinGeckoClient) FetchPrices(ctx context.Context, coinIDs []string) (map[string]domain.CryptoPrice, error) {
	if len(coinIDs) == 0 {
		return nil, apperrors.ErrEmptyInput
	}

	ids := strings.Join(coinIDs, ",")
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd&include_24hr_change=true", c.baseURL, ids)

	var apiResponse coinGeckoResponse
	if err := c.doWithRetry(ctx, url, &apiResponse); err != nil {
		return nil, err
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

type historicalResponse struct {
	Prices [][]float64 `json:"prices"`
}

// FetchHistoricalPrices retrieves historical prices for a coin.
func (c *CoinGeckoClient) FetchHistoricalPrices(ctx context.Context, coinID string, days int) ([]domain.CryptoPrice, error) {
	url := fmt.Sprintf("%s/coins/%s/market_chart?vs_currency=usd&days=%d&interval=daily", c.baseURL, coinID, days)

	var data historicalResponse
	if err := c.doWithRetry(ctx, url, &data); err != nil {
		return nil, err
	}

	prices := make([]domain.CryptoPrice, 0, len(data.Prices))
	for _, point := range data.Prices {
		if len(point) < 2 {
			continue
		}
		timestamp := time.UnixMilli(int64(point[0])).UTC()
		price := point[1]

		prices = append(prices, domain.CryptoPrice{
			Coin:      coinID,
			PriceUSD:  price,
			Change24h: 0,
			FetchedAt: timestamp,
		})
	}

	return prices, nil
}

func (c *CoinGeckoClient) doWithRetry(ctx context.Context, url string, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.backoffDuration * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := c.doRequest(ctx, url, result)
		if err == nil {
			return nil
		}

		lastErr = err

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if apperrors.IsAPIError(err) {
			var apiErr *apperrors.APIError
			if apperrors.IsAPIError(err) {
				_ = err.(*apperrors.APIError)
			}
			if apiErr != nil && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 429 {
				return err
			}
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *CoinGeckoClient) doRequest(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return apperrors.NewAPIError(resp.StatusCode, "rate limit exceeded", apperrors.ErrRateLimited)
	}

	if resp.StatusCode != http.StatusOK {
		return apperrors.NewAPIError(resp.StatusCode, resp.Status, nil)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

var (
	_ domain.PriceFetcher           = (*CoinGeckoClient)(nil)
	_ domain.HistoricalPriceFetcher = (*CoinGeckoClient)(nil)
)
