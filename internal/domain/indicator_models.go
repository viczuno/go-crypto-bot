// Package domain contains the core business types and interfaces for the crypto bot.
package domain

import "time"

// Indicator represents the type of technical indicator.
type Indicator string

const (
	IndicatorRSI            Indicator = "RSI"
	IndicatorMACD           Indicator = "MACD"
	IndicatorMovingAverage  Indicator = "MovingAverage"
	IndicatorBollingerBands Indicator = "BollingerBands"
)

// Signal represents trading signal strength.
type Signal int

const (
	SignalStrongSell Signal = -2
	SignalSell       Signal = -1
	SignalHold       Signal = 0
	SignalBuy        Signal = 1
	SignalStrongBuy  Signal = 2
)

// String returns the human-readable name of the signal.
func (s Signal) String() string {
	switch s {
	case SignalStrongBuy:
		return "Strong Buy"
	case SignalBuy:
		return "Buy"
	case SignalHold:
		return "Hold"
	case SignalSell:
		return "Sell"
	case SignalStrongSell:
		return "Strong Sell"
	default:
		return "Unknown"
	}
}

// Emoji returns the emoji representation of the signal.
func (s Signal) Emoji() string {
	switch s {
	case SignalStrongBuy:
		return "🟢"
	case SignalBuy:
		return "🔵"
	case SignalHold:
		return "⚪"
	case SignalSell:
		return "🟠"
	case SignalStrongSell:
		return "🔴"
	default:
		return "❓"
	}
}

// IndicatorResult represents the result of a single indicator calculation.
type IndicatorResult struct {
	Indicator  Indicator
	CoinID     string
	Value      float64
	Signal     Signal
	Confidence float64
	Metadata   map[string]float64
	Error      error
}

// IndicatorSummary aggregates multiple indicator results with a consensus signal.
type IndicatorSummary struct {
	CoinID           string
	Indicators       []IndicatorResult
	Consensus        Signal
	Confidence       float64
	SignalCounts     SignalCounts
	CalculatedAt     time.Time
	InsufficientData bool
}

// SignalCounts tracks the count of each signal type.
type SignalCounts struct {
	StrongBuy  int
	Buy        int
	Hold       int
	Sell       int
	StrongSell int
}

// Total returns the total number of signals counted.
func (sc SignalCounts) Total() int {
	return sc.StrongBuy + sc.Buy + sc.Hold + sc.Sell + sc.StrongSell
}

// Indicator calculation parameters.
const (
	// RSI parameters
	RSIPeriod           = 14
	RSIOverbought       = 70.0
	RSIOversold         = 30.0
	RSIStrongOverbought = 80.0
	RSIStrongOversold   = 20.0

	// MACD parameters
	MACDFastPeriod   = 12
	MACDSlowPeriod   = 26
	MACDSignalPeriod = 9

	// Moving Average parameters
	MAShortPeriod = 50
	MALongPeriod  = 200

	// Bollinger Bands parameters
	BollingerPeriod = 20
	BollingerStdDev = 2.0

	// Minimum data requirements
	MinDataForRSI       = RSIPeriod + 1
	MinDataForMACD      = MACDSlowPeriod + MACDSignalPeriod
	MinDataForMA        = MALongPeriod
	MinDataForBollinger = BollingerPeriod
)

// Indicator weights for consensus calculation.
const (
	WeightMACD      = 0.30
	WeightRSI       = 0.25
	WeightMA        = 0.25
	WeightBollinger = 0.20
)
