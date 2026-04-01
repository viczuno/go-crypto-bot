# CLAUDE.md - Go Crypto Bot

## Project Overview

A Go cryptocurrency tracking bot that fetches prices from CoinGecko API, stores historical data in SQLite, calculates technical indicators (RSI, MACD, Moving Averages, Bollinger Bands), and generates a README with price tables and charts. Runs via GitHub Actions every 12 hours.

## Quick Reference

```bash
make build          # Build binary to bin/crypto-bot
make test           # Run tests with race detection
make test-coverage  # Generate coverage.html report
make lint           # Run golangci-lint
make fmt            # Format code with gofmt
make run            # Run the application
```

## Project Structure

```
cmd/main.go                     # Entry point - flag parsing, orchestration
internal/
├── api/client.go               # CoinGecko API client (PriceFetcher interface)
├── coins/                      # Coin discovery sources
│   ├── aggregator.go           # Multi-source aggregator with deduplication
│   ├── static_source.go        # Static coin list from YAML
│   ├── trending_source.go      # API-based trending coins
│   └── market_cap_source.go    # Top coins by market cap
├── config/config.go            # Environment-based configuration
├── db/sqlite.go                # SQLite repository (PriceRepository interface)
├── domain/
│   ├── interfaces.go           # Core contracts (PriceFetcher, PriceRepository, etc.)
│   ├── models.go               # Data models (CryptoPrice, CoinStats, etc.)
│   ├── indicator_models.go     # Indicator types (Signal, IndicatorResult)
│   ├── constants.go            # Indicator parameters and weights
│   └── mocks/mocks.go          # Function-based test mocks
├── errors/errors.go            # Custom error types (APIError, NotFoundError, etc.)
├── indicators/
│   ├── signal_generator.go     # Weighted consensus from 4 indicators
│   ├── rsi.go                  # RSI (period=14)
│   ├── macd.go                 # MACD (12/26/9)
│   ├── moving_average.go       # SMA 50/200 with cross detection
│   ├── bollinger.go            # Bollinger Bands (20, 2.0)
│   └── helpers.go              # SMA, EMA, StdDev calculations
├── markdown/builder.go         # README generator with HTML tables
├── exporter/hugo.go            # Hugo static site data exporter
└── service/crypto_service.go   # Business logic orchestration
```

## Development Workflow

### Starting a New Feature

```bash
# 1. Create feature branch from main
git checkout main
git pull origin main
git checkout -b feature/your-feature-name

# 2. Implement changes following project patterns

# 3. Run tests
make test

# 4. Format and lint
make fmt
make lint

# 5. Commit with descriptive message
git add .
git commit -m "Add feature description"

# 6. Push and create PR
git push -u origin feature/your-feature-name
```

### Code Patterns to Follow

**Interface-Based Design**
- All major components depend on interfaces from `internal/domain/interfaces.go`
- New components should implement existing interfaces or define new ones in domain

**Constructor Pattern**
```go
func NewComponent(dep1 Interface1, dep2 Interface2) *Component {
    return &Component{dep1: dep1, dep2: dep2}
}
```

**Options Pattern (for configurable components)**
```go
type Option func(*Component)

func WithTimeout(d time.Duration) Option {
    return func(c *Component) { c.timeout = d }
}

func NewComponent(opts ...Option) *Component {
    c := &Component{timeout: defaultTimeout}
    for _, opt := range opts {
        opt(c)
    }
    return c
}
```

**Error Handling**
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("operation description: %w", err)
}

// Use custom error types from internal/errors
return &errors.APIError{StatusCode: 500, Message: "server error"}
```

**Context Propagation**
- All I/O operations must accept `context.Context` as first parameter
- Respect context cancellation and deadlines

## Testing Standards

### Test File Location
- Test files in same package: `package_name_test.go`
- Use `_test` suffix for black-box testing

### Mock Pattern
Use function-based mocks from `internal/domain/mocks/`:
```go
mock := &mocks.MockPriceFetcher{
    FetchPricesFunc: func(ctx context.Context, coinIDs []string) (map[string]domain.CryptoPrice, error) {
        return map[string]domain.CryptoPrice{
            "bitcoin": {Price: 50000, Change24h: 2.5},
        }, nil
    },
}
```

### Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {"valid input", validInput, expectedOutput, false},
        {"invalid input", invalidInput, OutputType{}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### HTTP Testing
Use `httptest.Server` for API client tests:
```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"bitcoin": {"usd": 50000}}`))
}))
defer server.Close()

client := api.NewCoinGeckoClient(api.WithBaseURL(server.URL))
```

## Configuration

### Environment Variables (CRYPTO_BOT_ prefix)
| Variable | Default | Description |
|----------|---------|-------------|
| `CRYPTO_BOT_DB_PATH` | `./crypto_history.db` | SQLite database path |
| `CRYPTO_BOT_README_PATH` | `./README.md` | Output README path |
| `CRYPTO_BOT_API_URL` | `https://api.coingecko.com/api/v3` | CoinGecko API base URL |
| `CRYPTO_BOT_API_TIMEOUT` | `30s` | HTTP request timeout |
| `CRYPTO_BOT_TIMEOUT` | `5m` | Overall execution timeout |
| `CRYPTO_BOT_MAX_RETRIES` | `3` | API retry attempts |
| `CRYPTO_BOT_HISTORY_DAYS` | `30` | Days of history to track |

### Coin Configuration (coins.yaml)
```yaml
coin_sources:
  - type: static
    coins:
      - id: bitcoin
        name: Bitcoin
        symbol: BTC
  - type: top_by_market_cap
    count: 10
    exclude: [tether, usd-coin]
  - type: trending
    count: 5
```

## Key Interfaces

```go
// Fetch current prices from external API
type PriceFetcher interface {
    FetchPrices(ctx context.Context, coinIDs []string) (map[string]CryptoPrice, error)
}

// Fetch historical price data
type HistoricalPriceFetcher interface {
    FetchHistoricalPrices(ctx context.Context, coinID string, days int) ([]HistoricalPrice, error)
}

// Store and retrieve price data
type PriceRepository interface {
    SavePrice(ctx context.Context, coin string, price float64) error
    GetPrice(ctx context.Context, coin string, daysAgo int) (float64, error)
    GetPriceHistory(ctx context.Context, coin string, days int) ([]float64, error)
    Close() error
}

// Calculate technical indicator
type IndicatorCalculator interface {
    Calculate(prices []float64) (IndicatorResult, error)
    Name() string
}

// Aggregate multiple indicators
type SignalGenerator interface {
    GenerateSignal(prices []float64) (IndicatorSummary, error)
}

// Discover coins from various sources
type CoinSource interface {
    FetchCoins(ctx context.Context) ([]CoinMetadata, error)
    Type() string
}
```

## Technical Indicators

| Indicator | Parameters | Signals |
|-----------|------------|---------|
| RSI | period=14 | StrongBuy (<=20), Buy (20-30), Hold (30-70), Sell (70-80), StrongSell (>=80) |
| MACD | fast=12, slow=26, signal=9 | Based on MACD/signal line crossovers and histogram |
| Moving Average | SMA 50, SMA 200 | Golden cross (bullish), Death cross (bearish) |
| Bollinger Bands | period=20, stddev=2.0 | Based on %B position within bands |

**Consensus Weights**: MACD (30%), RSI (25%), MA (25%), Bollinger (20%)

## Adding New Features

### New Technical Indicator
1. Create `internal/indicators/your_indicator.go`
2. Implement `domain.IndicatorCalculator` interface
3. Add to signal generator in `signal_generator.go`
4. Add weight constant in `domain/constants.go`
5. Write table-driven tests

### New Coin Source
1. Create `internal/coins/your_source.go`
2. Implement `domain.CoinSource` interface
3. Register in coin loader (`config/coin_loader.go`)
4. Add YAML configuration support

### New Output Format
1. Implement `domain.ReadmeGenerator` interface
2. Wire up in `cmd/main.go` or add as option to service

## Common Gotchas

- **CoinGecko rate limits**: Free tier ~50 calls/min. Use retry with backoff.
- **SQLite concurrent writes**: Use transactions; only one writer at a time.
- **Context cancellation**: Always check `ctx.Done()` in loops.
- **Indicator data requirements**: RSI needs 15+ prices, MACD needs 35+ prices.
- **Time zones**: All timestamps stored as UTC in database.

## CI/CD (GitHub Actions)

Workflow runs every 12 hours (`.github/workflows/bot.yaml`):
1. Checkout → Setup Go 1.24 → Run tests
2. Setup Hugo → Run bot → Build site
3. Commit data files → Deploy to GitHub Pages

## Dependencies

- `modernc.org/sqlite` - Pure Go SQLite driver (no CGO)
- `gopkg.in/yaml.v3` - YAML parsing (indirect)
