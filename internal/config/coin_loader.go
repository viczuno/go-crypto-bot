package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/viczuno/go-crypto-bot/internal/coins"
	"github.com/viczuno/go-crypto-bot/internal/domain"
	"gopkg.in/yaml.v3"
)

// CoinConfig represents the coin configuration file structure.
type CoinConfig struct {
	CoinSources  []CoinSourceConfig `yaml:"coin_sources"`
	DisplayOpts  DisplayOptions     `yaml:"display_options,omitempty"`
}

// CoinSourceConfig represents a single coin source configuration.
type CoinSourceConfig struct {
	Type       string                   `yaml:"type"`
	Coins      []StaticCoinConfig       `yaml:"coins,omitempty"`
	Count      int                      `yaml:"count,omitempty"`
	Exclude    []string                 `yaml:"exclude,omitempty"`
	TimeWindow string                   `yaml:"time_window,omitempty"`
	Name       string                   `yaml:"name,omitempty"`
	Tags       []string                 `yaml:"tags,omitempty"`
}

// StaticCoinConfig represents a static coin entry.
type StaticCoinConfig struct {
	ID     string   `yaml:"id"`
	Name   string   `yaml:"name,omitempty"`
	Symbol string   `yaml:"symbol,omitempty"`
	Tags   []string `yaml:"tags,omitempty"`
}

// DisplayOptions controls how coins are displayed.
type DisplayOptions struct {
	GroupByTags bool   `yaml:"group_by_tags"`
	SortBy      string `yaml:"sort_by"`
	MaxDisplay  int    `yaml:"max_display"`
}

// LoadCoinsFromConfig loads coin configuration from a YAML file.
func LoadCoinsFromConfig(ctx context.Context, configPath, apiBaseURL string) ([]domain.CoinMetadata, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg CoinConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	if len(cfg.CoinSources) == 0 {
		return nil, fmt.Errorf("no coin sources defined in config")
	}

	sources := make([]domain.CoinSource, 0, len(cfg.CoinSources))

	for i, srcCfg := range cfg.CoinSources {
		source, err := createCoinSource(srcCfg, apiBaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create coin source %d (%s): %w", i, srcCfg.Type, err)
		}
		sources = append(sources, source)
	}

	aggregator := coins.NewAggregator(sources...)
	allCoins, err := aggregator.FetchCoins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch coins: %w", err)
	}

	if cfg.DisplayOpts.MaxDisplay > 0 && len(allCoins) > cfg.DisplayOpts.MaxDisplay {
		allCoins = allCoins[:cfg.DisplayOpts.MaxDisplay]
	}

	return allCoins, nil
}

func createCoinSource(cfg CoinSourceConfig, apiBaseURL string) (domain.CoinSource, error) {
	switch strings.ToLower(cfg.Type) {
	case "static":
		if len(cfg.Coins) == 0 {
			return nil, fmt.Errorf("static source requires at least one coin")
		}
		staticCoins := make([]domain.CoinMetadata, len(cfg.Coins))
		for i, c := range cfg.Coins {
			staticCoins[i] = domain.CoinMetadata{
				ID:     c.ID,
				Name:   c.Name,
				Symbol: c.Symbol,
				Tags:   c.Tags,
			}
		}
		return coins.NewStaticSource(staticCoins), nil

	case "top_by_market_cap":
		count := cfg.Count
		if count <= 0 {
			count = 10
		}
		return coins.NewMarketCapSource(apiBaseURL, count, cfg.Exclude), nil

	case "trending":
		count := cfg.Count
		if count <= 0 {
			count = 5
		}
		return coins.NewTrendingSource(apiBaseURL, count), nil

	default:
		return nil, fmt.Errorf("unknown source type: %s", cfg.Type)
	}
}

// ValidateCoins validates that all coins can be fetched successfully.
func ValidateCoins(ctx context.Context, configPath, apiBaseURL string) error {
	coins, err := LoadCoinsFromConfig(ctx, configPath, apiBaseURL)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Successfully loaded %d coins:\n", len(coins))
	for _, coin := range coins {
		tags := ""
		if len(coin.Tags) > 0 {
			tags = fmt.Sprintf(" [%s]", strings.Join(coin.Tags, ", "))
		}
		fmt.Printf("  - %s (%s)%s\n", coin.Name, coin.Symbol, tags)
	}

	return nil
}
