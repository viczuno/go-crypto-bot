// Package indicators provides technical indicator calculations for crypto analysis.
package indicators

import (
	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// MACDCalculator calculates the Moving Average Convergence Divergence.
type MACDCalculator struct {
	fastPeriod   int
	slowPeriod   int
	signalPeriod int
}

// NewMACDCalculator creates a new MACD calculator with standard parameters (12, 26, 9).
func NewMACDCalculator() *MACDCalculator {
	return &MACDCalculator{
		fastPeriod:   domain.MACDFastPeriod,
		slowPeriod:   domain.MACDSlowPeriod,
		signalPeriod: domain.MACDSignalPeriod,
	}
}

var _ domain.IndicatorCalculator = (*MACDCalculator)(nil)

// Name returns the indicator type.
func (m *MACDCalculator) Name() domain.Indicator {
	return domain.IndicatorMACD
}

// MinDataPoints returns the minimum required data points for MACD calculation.
func (m *MACDCalculator) MinDataPoints() int {
	return domain.MinDataForMACD
}

// Calculate computes the MACD from price history.
// MACD Line = 12-EMA - 26-EMA
// Signal Line = 9-EMA of MACD Line
// Histogram = MACD Line - Signal Line
func (m *MACDCalculator) Calculate(coinID string, prices []domain.CryptoPrice) domain.IndicatorResult {
	result := domain.IndicatorResult{
		Indicator: domain.IndicatorMACD,
		CoinID:    coinID,
		Metadata:  make(map[string]float64),
	}

	if len(prices) < m.MinDataPoints() {
		result.Error = ErrInsufficientData
		result.Signal = domain.SignalHold
		return result
	}

	// Extract price values (newest first, so reverse for calculation)
	priceValues := extractPrices(prices)
	reverse(priceValues)

	// Calculate MACD components
	macdLine, signalLine, histogram := m.calculateMACD(priceValues)

	result.Value = macdLine
	result.Metadata["macd_line"] = macdLine
	result.Metadata["signal_line"] = signalLine
	result.Metadata["histogram"] = histogram
	result.Metadata["fast_period"] = float64(m.fastPeriod)
	result.Metadata["slow_period"] = float64(m.slowPeriod)
	result.Metadata["signal_period"] = float64(m.signalPeriod)

	// Determine signal based on MACD crossover and histogram
	result.Signal, result.Confidence = m.determineSignal(macdLine, signalLine, histogram, priceValues)

	return result
}

func (m *MACDCalculator) calculateMACD(prices []float64) (macdLine, signalLine, histogram float64) {
	// Calculate fast and slow EMAs
	fastEMA := calculateEMA(prices, m.fastPeriod)
	slowEMA := calculateEMA(prices, m.slowPeriod)

	// MACD Line = Fast EMA - Slow EMA
	macdLine = fastEMA - slowEMA

	// Calculate MACD history for signal line
	macdHistory := m.calculateMACDHistory(prices)

	// Signal Line = 9-EMA of MACD history
	if len(macdHistory) >= m.signalPeriod {
		signalLine = calculateEMA(macdHistory, m.signalPeriod)
	}

	// Histogram = MACD Line - Signal Line
	histogram = macdLine - signalLine

	return macdLine, signalLine, histogram
}

func (m *MACDCalculator) calculateMACDHistory(prices []float64) []float64 {
	if len(prices) < m.slowPeriod {
		return nil
	}

	// Generate MACD values for each point starting from slowPeriod
	macdHistory := make([]float64, 0, len(prices)-m.slowPeriod+1)

	for i := m.slowPeriod; i <= len(prices); i++ {
		subset := prices[:i]
		fastEMA := calculateEMA(subset, m.fastPeriod)
		slowEMA := calculateEMA(subset, m.slowPeriod)
		macdHistory = append(macdHistory, fastEMA-slowEMA)
	}

	return macdHistory
}

func (m *MACDCalculator) determineSignal(macdLine, signalLine, histogram float64, prices []float64) (domain.Signal, float64) {
	// Get previous histogram to detect crossover
	prevHistogram := m.getPreviousHistogram(prices)

	// Detect crossover direction
	crossoverUp := prevHistogram < 0 && histogram > 0
	crossoverDown := prevHistogram > 0 && histogram < 0

	// Calculate signal strength based on histogram magnitude
	histogramStrength := abs(histogram) / abs(signalLine)
	if histogramStrength > 1 {
		histogramStrength = 1
	}

	switch {
	case crossoverUp && histogram > 0:
		// Bullish crossover: MACD crossed above signal line
		if histogramStrength > 0.5 {
			return domain.SignalStrongBuy, 0.7 + (0.3 * histogramStrength)
		}
		return domain.SignalBuy, 0.5 + (0.2 * histogramStrength)

	case crossoverDown && histogram < 0:
		// Bearish crossover: MACD crossed below signal line
		if histogramStrength > 0.5 {
			return domain.SignalStrongSell, 0.7 + (0.3 * histogramStrength)
		}
		return domain.SignalSell, 0.5 + (0.2 * histogramStrength)

	case histogram > 0 && macdLine > 0:
		// Bullish momentum: MACD and histogram both positive
		return domain.SignalBuy, 0.4 + (0.2 * histogramStrength)

	case histogram < 0 && macdLine < 0:
		// Bearish momentum: MACD and histogram both negative
		return domain.SignalSell, 0.4 + (0.2 * histogramStrength)

	default:
		// Mixed signals or near zero
		return domain.SignalHold, 0.3 + (0.2 * (1 - histogramStrength))
	}
}

func (m *MACDCalculator) getPreviousHistogram(prices []float64) float64 {
	if len(prices) < m.MinDataPoints()+1 {
		return 0
	}

	// Calculate MACD components for previous day
	prevPrices := prices[:len(prices)-1]
	prevMACDLine, prevSignalLine, _ := m.calculateMACD(prevPrices)
	return prevMACDLine - prevSignalLine
}
