// Package indicators provides technical indicator calculations for crypto analysis.
package indicators

import (
	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// MovingAverageCalculator calculates Moving Averages and detects golden/death crosses.
type MovingAverageCalculator struct {
	shortPeriod int
	longPeriod  int
}

// NewMovingAverageCalculator creates a new MA calculator with standard parameters (50, 200).
func NewMovingAverageCalculator() *MovingAverageCalculator {
	return &MovingAverageCalculator{
		shortPeriod: domain.MAShortPeriod,
		longPeriod:  domain.MALongPeriod,
	}
}

var _ domain.IndicatorCalculator = (*MovingAverageCalculator)(nil)

// Name returns the indicator type.
func (ma *MovingAverageCalculator) Name() domain.Indicator {
	return domain.IndicatorMovingAverage
}

// MinDataPoints returns the minimum required data points for MA calculation.
func (ma *MovingAverageCalculator) MinDataPoints() int {
	return domain.MinDataForMA
}

// Calculate computes the Moving Averages and detects cross patterns.
// Golden Cross: Short MA crosses above Long MA (bullish)
// Death Cross: Short MA crosses below Long MA (bearish)
func (ma *MovingAverageCalculator) Calculate(coinID string, prices []domain.CryptoPrice) domain.IndicatorResult {
	result := domain.IndicatorResult{
		Indicator: domain.IndicatorMovingAverage,
		CoinID:    coinID,
		Metadata:  make(map[string]float64),
	}

	if len(prices) < ma.MinDataPoints() {
		result.Error = ErrInsufficientData
		result.Signal = domain.SignalHold
		return result
	}

	// Extract price values (newest first, so reverse for calculation)
	priceValues := extractPrices(prices)
	reverse(priceValues)

	currentPrice := priceValues[len(priceValues)-1]

	// Calculate current SMAs
	shortSMA := calculateSMA(priceValues, ma.shortPeriod)
	longSMA := calculateSMA(priceValues, ma.longPeriod)

	// Calculate previous SMAs (for crossover detection)
	prevPrices := priceValues[:len(priceValues)-1]
	prevShortSMA := calculateSMA(prevPrices, ma.shortPeriod)
	prevLongSMA := calculateSMA(prevPrices, ma.longPeriod)

	// Store metadata
	result.Value = shortSMA // Primary value is the short-term MA
	result.Metadata["sma_50"] = shortSMA
	result.Metadata["sma_200"] = longSMA
	result.Metadata["price"] = currentPrice
	result.Metadata["short_period"] = float64(ma.shortPeriod)
	result.Metadata["long_period"] = float64(ma.longPeriod)

	// Calculate price position relative to MAs
	priceAboveShort := currentPrice > shortSMA
	priceAboveLong := currentPrice > longSMA

	// Detect crossovers
	goldenCross := prevShortSMA <= prevLongSMA && shortSMA > longSMA
	deathCross := prevShortSMA >= prevLongSMA && shortSMA < longSMA

	result.Metadata["golden_cross"] = boolToFloat(goldenCross)
	result.Metadata["death_cross"] = boolToFloat(deathCross)
	result.Metadata["price_above_short"] = boolToFloat(priceAboveShort)
	result.Metadata["price_above_long"] = boolToFloat(priceAboveLong)

	// Determine signal
	result.Signal, result.Confidence = ma.determineSignal(
		shortSMA, longSMA, currentPrice,
		goldenCross, deathCross,
		priceAboveShort, priceAboveLong,
	)

	return result
}

func (ma *MovingAverageCalculator) determineSignal(
	shortSMA, longSMA, currentPrice float64,
	goldenCross, deathCross bool,
	priceAboveShort, priceAboveLong bool,
) (domain.Signal, float64) {
	// Calculate trend strength
	trendStrength := abs(shortSMA-longSMA) / longSMA

	switch {
	case goldenCross:
		// Golden Cross: Strong bullish signal
		return domain.SignalStrongBuy, 0.8 + (0.2 * min(trendStrength*10, 1))

	case deathCross:
		// Death Cross: Strong bearish signal
		return domain.SignalStrongSell, 0.8 + (0.2 * min(trendStrength*10, 1))

	case shortSMA > longSMA && priceAboveShort && priceAboveLong:
		// Uptrend: Price above both MAs, short MA above long MA
		return domain.SignalBuy, 0.5 + (0.3 * min(trendStrength*10, 1))

	case shortSMA < longSMA && !priceAboveShort && !priceAboveLong:
		// Downtrend: Price below both MAs, short MA below long MA
		return domain.SignalSell, 0.5 + (0.3 * min(trendStrength*10, 1))

	case shortSMA > longSMA:
		// Weak bullish: MAs bullish but price mixed
		return domain.SignalBuy, 0.4

	case shortSMA < longSMA:
		// Weak bearish: MAs bearish but price mixed
		return domain.SignalSell, 0.4

	default:
		// MAs converging or price between MAs
		return domain.SignalHold, 0.3
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
