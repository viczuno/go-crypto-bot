// Package indicators provides technical indicator calculations for crypto analysis.
package indicators

import (
	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// RSICalculator calculates the Relative Strength Index.
type RSICalculator struct {
	period int
}

// NewRSICalculator creates a new RSI calculator with the standard 14-period.
func NewRSICalculator() *RSICalculator {
	return &RSICalculator{period: domain.RSIPeriod}
}

var _ domain.IndicatorCalculator = (*RSICalculator)(nil)

// Name returns the indicator type.
func (r *RSICalculator) Name() domain.Indicator {
	return domain.IndicatorRSI
}

// MinDataPoints returns the minimum required data points for RSI calculation.
func (r *RSICalculator) MinDataPoints() int {
	return domain.MinDataForRSI
}

// Calculate computes the RSI value from price history.
// RSI = 100 - (100 / (1 + RS))
// RS = Average Gain / Average Loss over the period.
func (r *RSICalculator) Calculate(coinID string, prices []domain.CryptoPrice) domain.IndicatorResult {
	result := domain.IndicatorResult{
		Indicator: domain.IndicatorRSI,
		CoinID:    coinID,
		Metadata:  make(map[string]float64),
	}

	if len(prices) < r.MinDataPoints() {
		result.Error = ErrInsufficientData
		result.Signal = domain.SignalHold
		return result
	}

	// Extract price values (newest first, so reverse for calculation)
	priceValues := extractPrices(prices)
	reverse(priceValues)

	rsi := calculateRSI(priceValues, r.period)
	result.Value = rsi
	result.Metadata["period"] = float64(r.period)

	// Determine signal based on RSI value
	result.Signal, result.Confidence = r.determineSignal(rsi)

	return result
}

func (r *RSICalculator) determineSignal(rsi float64) (domain.Signal, float64) {
	switch {
	case rsi <= domain.RSIStrongOversold:
		// RSI <= 20: Strong buy signal (deeply oversold)
		confidence := 1.0 - (rsi / domain.RSIStrongOversold)
		return domain.SignalStrongBuy, 0.7 + (0.3 * confidence)
	case rsi <= domain.RSIOversold:
		// RSI 20-30: Buy signal (oversold)
		confidence := 1.0 - ((rsi - domain.RSIStrongOversold) / (domain.RSIOversold - domain.RSIStrongOversold))
		return domain.SignalBuy, 0.5 + (0.2 * confidence)
	case rsi >= domain.RSIStrongOverbought:
		// RSI >= 80: Strong sell signal (deeply overbought)
		confidence := (rsi - domain.RSIStrongOverbought) / (100.0 - domain.RSIStrongOverbought)
		return domain.SignalStrongSell, 0.7 + (0.3 * confidence)
	case rsi >= domain.RSIOverbought:
		// RSI 70-80: Sell signal (overbought)
		confidence := (rsi - domain.RSIOverbought) / (domain.RSIStrongOverbought - domain.RSIOverbought)
		return domain.SignalSell, 0.5 + (0.2 * confidence)
	default:
		// RSI 30-70: Hold (neutral zone)
		// Confidence is lower the closer to 50 (most neutral)
		distanceFrom50 := abs(rsi - 50.0)
		confidence := distanceFrom50 / 20.0 // Max distance in neutral zone is 20
		return domain.SignalHold, 0.3 + (0.2 * confidence)
	}
}

// calculateRSI computes the RSI using the standard Wilder smoothing method.
func calculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50.0 // Default neutral value
	}

	// Calculate initial average gain and loss
	var gainSum, lossSum float64
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gainSum += change
		} else {
			lossSum += -change
		}
	}

	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)

	// Apply Wilder's smoothing for remaining prices
	for i := period + 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		var gain, loss float64
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
	}

	// Calculate RSI
	if avgLoss == 0 {
		return 100.0 // All gains, no losses
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))

	return rsi
}
