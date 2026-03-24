package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/api"
)

func TestCoinGeckoClient_FetchPrices(t *testing.T) {
	tests := []struct {
		name         string
		responseCode int
		responseBody string
		coinIDs      []string
		wantErr      bool
		wantLen      int
	}{
		{
			name:         "successful fetch",
			responseCode: http.StatusOK,
			responseBody: `{"bitcoin":{"usd":50000,"usd_24h_change":2.5},"ethereum":{"usd":3000,"usd_24h_change":-1.5}}`,
			coinIDs:      []string{"bitcoin", "ethereum"},
			wantErr:      false,
			wantLen:      2,
		},
		{
			name:         "empty coin list",
			responseCode: http.StatusOK,
			responseBody: `{}`,
			coinIDs:      []string{},
			wantErr:      true,
			wantLen:      0,
		},
		{
			name:         "server error",
			responseCode: http.StatusInternalServerError,
			responseBody: `{"error": "internal error"}`,
			coinIDs:      []string{"bitcoin"},
			wantErr:      true,
			wantLen:      0,
		},
		{
			name:         "rate limited",
			responseCode: http.StatusTooManyRequests,
			responseBody: `{"error": "rate limited"}`,
			coinIDs:      []string{"bitcoin"},
			wantErr:      true,
			wantLen:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := api.NewCoinGeckoClient(
				api.WithBaseURL(server.URL),
				api.WithRetry(0, time.Millisecond), // No retries for faster tests
			)

			ctx := context.Background()
			prices, err := client.FetchPrices(ctx, tt.coinIDs)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(prices) != tt.wantLen {
				t.Errorf("prices length = %d, want %d", len(prices), tt.wantLen)
			}

			if tt.wantLen > 0 {
				if btc, ok := prices["bitcoin"]; ok {
					if btc.PriceUSD != 50000 {
						t.Errorf("bitcoin price = %f, want 50000", btc.PriceUSD)
					}
					if btc.Change24h != 2.5 {
						t.Errorf("bitcoin change = %f, want 2.5", btc.Change24h)
					}
				}
			}
		})
	}
}

func TestCoinGeckoClient_FetchHistoricalPrices(t *testing.T) {
	responseBody := `{
		"prices": [
			[1609459200000, 29000.0],
			[1609545600000, 29500.0],
			[1609632000000, 30000.0]
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	client := api.NewCoinGeckoClient(api.WithBaseURL(server.URL))

	ctx := context.Background()
	prices, err := client.FetchHistoricalPrices(ctx, "bitcoin", 30)

	if err != nil {
		t.Fatalf("FetchHistoricalPrices() error = %v", err)
	}

	if len(prices) != 3 {
		t.Errorf("prices length = %d, want 3", len(prices))
	}

	if prices[0].PriceUSD != 29000.0 {
		t.Errorf("first price = %f, want 29000.0", prices[0].PriceUSD)
	}
}

func TestCoinGeckoClient_WithOptions(t *testing.T) {
	client := api.NewCoinGeckoClient(
		api.WithTimeout(10*time.Second),
		api.WithBaseURL("https://custom.api.com"),
		api.WithRetry(5, 2*time.Second),
	)

	// Just verify the client was created successfully
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestCoinGeckoClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"bitcoin":{"usd":50000}}`))
	}))
	defer server.Close()

	client := api.NewCoinGeckoClient(
		api.WithBaseURL(server.URL),
		api.WithRetry(0, time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.FetchPrices(ctx, []string{"bitcoin"})
	if err == nil {
		t.Error("expected error due to context cancellation")
	}
}

func TestCoinGeckoClient_RetryOnError(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"bitcoin":{"usd":50000,"usd_24h_change":2.5}}`))
	}))
	defer server.Close()

	client := api.NewCoinGeckoClient(
		api.WithBaseURL(server.URL),
		api.WithRetry(3, time.Millisecond),
	)

	ctx := context.Background()
	prices, err := client.FetchPrices(ctx, []string{"bitcoin"})

	if err != nil {
		t.Errorf("expected success after retries, got error: %v", err)
	}

	if len(prices) != 1 {
		t.Errorf("expected 1 price, got %d", len(prices))
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls (2 failures + 1 success), got %d", callCount)
	}
}
