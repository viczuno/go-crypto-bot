// Package indicators provides technical indicator calculations for crypto analysis.
package indicators

import (
	"errors"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// ErrInsufficientData is returned when there's not enough price data.
var ErrInsufficientData = errors.New("insufficient price data for indicator calculation")

// extractPrices extracts price values from CryptoPrice slice.
func extractPrices(prices []domain.CryptoPrice) []float64 {
	values := make([]float64, len(prices))
	for i, p := range prices {
		values[i] = p.PriceUSD
	}
	return values
}

// reverse reverses a slice in place.
func reverse(s []float64) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// calculateSMA computes the Simple Moving Average for the given period.
func calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

// calculateEMA computes the Exponential Moving Average for the given period.
// Uses the standard multiplier: 2 / (period + 1)
func calculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	// Start with SMA for first period values
	sma := 0.0
	for i := 0; i < period; i++ {
		sma += prices[i]
	}
	sma /= float64(period)

	if len(prices) == period {
		return sma
	}

	// Calculate EMA for remaining prices
	multiplier := 2.0 / float64(period+1)
	ema := sma

	for i := period; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}

	return ema
}

// calculateStdDev computes the standard deviation from the mean.
func calculateStdDev(prices []float64, mean float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sumSquaredDiff := 0.0
	start := len(prices) - period
	for i := start; i < len(prices); i++ {
		diff := prices[i] - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(period)
	return sqrt(variance)
}

// sqrt computes square root using Newton's method.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
