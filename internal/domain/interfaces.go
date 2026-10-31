// Package domain contains contracts shared by the Go analytics.
package domain

// IndicatorCalculator defines the interface for calculating a single technical indicator.
type IndicatorCalculator interface {
	// Calculate computes the indicator value from price history.
	// Returns an IndicatorResult with the calculated value and signal.
	Calculate(coinID string, prices []CryptoPrice) IndicatorResult

	// Name returns the indicator type.
	Name() Indicator

	// MinDataPoints returns the minimum number of price points required.
	MinDataPoints() int
}

// SignalGenerator defines the interface for generating aggregated trading signals.
type SignalGenerator interface {
	// GenerateSignals calculates all indicators and returns an aggregated summary.
	GenerateSignals(coinID string, prices []CryptoPrice) IndicatorSummary
}
