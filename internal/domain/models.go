package domain

import "time"

// CryptoPrice represents the current price data for a cryptocurrency
type CryptoPrice struct {
	Coin      string
	PriceUSD  float64
	Change24h float64
	FetchedAt time.Time
}

// PriceChange represents historical price change data
type PriceChange struct {
	PastPrice    float64
	CurrentPrice float64
	AbsChange    float64
	PctChange    float64
	HasData      bool
	Days         int
}

// CoinStats aggregates all statistics for a single coin
type CoinStats struct {
	Name      string
	Symbol    string
	Price     float64
	Change24h float64
	Change7d  PriceChange
	Change30d PriceChange
}

// CoinMetadata contains display information for coins
type CoinMetadata struct {
	ID     string
	Name   string
	Symbol string
}

// DefaultCoins returns the list of tracked cryptocurrencies with metadata
func DefaultCoins() []CoinMetadata {
	return []CoinMetadata{
		{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC"},
		{ID: "ethereum", Name: "Ethereum", Symbol: "ETH"},
		{ID: "solana", Name: "Solana", Symbol: "SOL"},
		{ID: "cardano", Name: "Cardano", Symbol: "ADA"},
		{ID: "polkadot", Name: "Polkadot", Symbol: "DOT"},
	}
}
