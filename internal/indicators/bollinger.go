// Package indicators provides technical indicator calculations for crypto analysis.
package indicators

import (
	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// BollingerBandsCalculator calculates Bollinger Bands.
type BollingerBandsCalculator struct {
	period int
	stdDev float64
}

// NewBollingerBandsCalculator creates a new Bollinger Bands calculator with standard parameters (20, 2).
func NewBollingerBandsCalculator() *BollingerBandsCalculator {
	return &BollingerBandsCalculator{
		period: domain.BollingerPeriod,
		stdDev: domain.BollingerStdDev,
	}
}

var _ domain.IndicatorCalculator = (*BollingerBandsCalculator)(nil)

// Name returns the indicator type.
func (bb *BollingerBandsCalculator) Name() domain.Indicator {
	return domain.IndicatorBollingerBands
}

// MinDataPoints returns the minimum required data points for Bollinger Bands calculation.
func (bb *BollingerBandsCalculator) MinDataPoints() int {
	return domain.MinDataForBollinger
}

// Calculate computes the Bollinger Bands from price history.
// Middle Band = 20-period SMA
// Upper Band = Middle Band + (2 * Standard Deviation)
// Lower Band = Middle Band - (2 * Standard Deviation)
func (bb *BollingerBandsCalculator) Calculate(coinID string, prices []domain.CryptoPrice) domain.IndicatorResult {
	result := domain.IndicatorResult{
		Indicator: domain.IndicatorBollingerBands,
		CoinID:    coinID,
		Metadata:  make(map[string]float64),
	}

	if len(prices) < bb.MinDataPoints() {
		result.Error = ErrInsufficientData
		result.Signal = domain.SignalHold
		return result
	}

	// Extract price values (newest first, so reverse for calculation)
	priceValues := extractPrices(prices)
	reverse(priceValues)

	currentPrice := priceValues[len(priceValues)-1]

	// Calculate bands
	middleBand := calculateSMA(priceValues, bb.period)
	stdDeviation := calculateStdDev(priceValues, middleBand, bb.period)
	upperBand := middleBand + (bb.stdDev * stdDeviation)
	lowerBand := middleBand - (bb.stdDev * stdDeviation)

	// Calculate %B (position within bands)
	// %B = (Price - Lower Band) / (Upper Band - Lower Band)
	bandWidth := upperBand - lowerBand
	var percentB float64
	if bandWidth > 0 {
		percentB = (currentPrice - lowerBand) / bandWidth
	}

	// Calculate bandwidth (volatility indicator)
	// Bandwidth = (Upper Band - Lower Band) / Middle Band
	var bandwidth float64
	if middleBand > 0 {
		bandwidth = bandWidth / middleBand
	}

	// Store metadata
	result.Value = percentB // Primary value is %B position
	result.Metadata["upper_band"] = upperBand
	result.Metadata["middle_band"] = middleBand
	result.Metadata["lower_band"] = lowerBand
	result.Metadata["percent_b"] = percentB
	result.Metadata["bandwidth"] = bandwidth
	result.Metadata["price"] = currentPrice
	result.Metadata["period"] = float64(bb.period)
	result.Metadata["std_dev_multiplier"] = bb.stdDev

	// Determine signal based on price position within bands
	result.Signal, result.Confidence = bb.determineSignal(currentPrice, upperBand, middleBand, lowerBand, percentB, bandwidth)

	return result
}

func (bb *BollingerBandsCalculator) determineSignal(
	price, upperBand, middleBand, lowerBand, percentB, bandwidth float64,
) (domain.Signal, float64) {
	// Determine distance from bands (for confidence calculation)
	distanceFromUpper := (upperBand - price) / (upperBand - middleBand)
	distanceFromLower := (price - lowerBand) / (middleBand - lowerBand)

	switch {
	case percentB <= 0:
		// Price at or below lower band - Strong oversold
		confidence := 0.7 + (0.3 * min(abs(percentB)*2, 1))
		return domain.SignalStrongBuy, confidence

	case percentB < 0.2:
		// Price near lower band - Oversold
		confidence := 0.5 + (0.2 * (1 - percentB/0.2))
		return domain.SignalBuy, confidence

	case percentB >= 1.0:
		// Price at or above upper band - Strong overbought
		confidence := 0.7 + (0.3 * min((percentB-1.0)*2, 1))
		return domain.SignalStrongSell, confidence

	case percentB > 0.8:
		// Price near upper band - Overbought
		confidence := 0.5 + (0.2 * ((percentB - 0.8) / 0.2))
		return domain.SignalSell, confidence

	case percentB >= 0.4 && percentB <= 0.6:
		// Price in middle zone - Neutral
		// Higher confidence when exactly at middle
		distanceFromMiddle := abs(percentB - 0.5) / 0.1
		confidence := 0.5 - (0.2 * distanceFromMiddle)
		return domain.SignalHold, confidence

	case percentB > 0.5:
		// Price in upper half but not overbought
		// Slight bearish bias
		_ = distanceFromUpper // suppress unused warning
		return domain.SignalHold, 0.35

	default:
		// Price in lower half but not oversold
		// Slight bullish bias
		_ = distanceFromLower // suppress unused warning
		return domain.SignalHold, 0.35
	}
}
