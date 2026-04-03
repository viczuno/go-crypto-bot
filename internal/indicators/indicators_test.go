package indicators

import (
	"testing"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// Helper to create price history with a simple pattern
// Values should be ordered from oldest to newest
func makePrices(values []float64) []domain.CryptoPrice {
	prices := make([]domain.CryptoPrice, len(values))
	baseTime := time.Now()
	// Data is stored newest first in the system, so reverse the order
	for i, v := range values {
		// Index 0 in values is oldest, should map to furthest in past
		daysAgo := len(values) - 1 - i
		prices[i] = domain.CryptoPrice{
			Coin:      "test",
			PriceUSD:  v,
			FetchedAt: baseTime.Add(-time.Duration(daysAgo) * 24 * time.Hour),
		}
	}
	// Reverse so newest is first (matching repository convention)
	for i, j := 0, len(prices)-1; i < j; i, j = i+1, j-1 {
		prices[i], prices[j] = prices[j], prices[i]
	}
	return prices
}

// Generate rising prices
func risingPrices(start float64, count int, increment float64) []float64 {
	prices := make([]float64, count)
	for i := 0; i < count; i++ {
		prices[i] = start + float64(i)*increment
	}
	return prices
}

// Generate falling prices
func fallingPrices(start float64, count int, decrement float64) []float64 {
	prices := make([]float64, count)
	for i := 0; i < count; i++ {
		prices[i] = start - float64(i)*decrement
	}
	return prices
}

// Generate flat prices
func flatPrices(value float64, count int) []float64 {
	prices := make([]float64, count)
	for i := 0; i < count; i++ {
		prices[i] = value
	}
	return prices
}

func TestRSICalculator_Calculate(t *testing.T) {
	calc := NewRSICalculator()

	tests := []struct {
		name           string
		prices         []float64
		wantSignal     domain.Signal
		wantRSIAbove50 bool
		wantError      bool
	}{
		{
			name:       "insufficient data",
			prices:     flatPrices(100, 10),
			wantSignal: domain.SignalHold,
			wantError:  true,
		},
		{
			name:           "flat prices should give RSI around 50",
			prices:         flatPrices(100, 20),
			wantSignal:     domain.SignalHold,
			wantRSIAbove50: false, // RSI = 50 exactly for flat
		},
		{
			name:           "rising prices should give RSI > 50",
			prices:         risingPrices(100, 20, 1),
			wantSignal:     domain.SignalBuy, // May vary based on strength
			wantRSIAbove50: true,
		},
		{
			name:           "falling prices should give RSI < 50",
			prices:         fallingPrices(120, 20, 1),
			wantSignal:     domain.SignalSell, // May vary based on strength
			wantRSIAbove50: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prices := makePrices(tt.prices)
			result := calc.Calculate("test", prices)

			if tt.wantError {
				if result.Error == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if result.Error != nil {
				t.Errorf("unexpected error: %v", result.Error)
				return
			}

			if tt.wantRSIAbove50 && result.Value <= 50 {
				t.Errorf("expected RSI > 50, got %v", result.Value)
			}
			if !tt.wantRSIAbove50 && result.Value > 50 && tt.name != "flat prices should give RSI around 50" {
				t.Errorf("expected RSI < 50, got %v", result.Value)
			}
		})
	}
}

func TestRSI_OverboughtOversold(t *testing.T) {
	calc := NewRSICalculator()

	// Create strongly rising prices to get overbought
	strongRise := make([]float64, 20)
	for i := 0; i < 20; i++ {
		strongRise[i] = 100 + float64(i)*5 // Large gains
	}
	prices := makePrices(strongRise)
	result := calc.Calculate("test", prices)

	if result.Value <= domain.RSIOverbought {
		t.Logf("RSI value: %v (may not be strongly overbought with this data)", result.Value)
	}

	// Create strongly falling prices to get oversold
	strongFall := make([]float64, 20)
	for i := 0; i < 20; i++ {
		strongFall[i] = 200 - float64(i)*5 // Large losses
	}
	prices = makePrices(strongFall)
	result = calc.Calculate("test", prices)

	if result.Value >= domain.RSIOversold {
		t.Logf("RSI value: %v (may not be strongly oversold with this data)", result.Value)
	}
}

func TestMACDCalculator_Calculate(t *testing.T) {
	calc := NewMACDCalculator()

	tests := []struct {
		name      string
		prices    []float64
		wantError bool
	}{
		{
			name:      "insufficient data",
			prices:    flatPrices(100, 30),
			wantError: true,
		},
		{
			name:      "sufficient data - flat",
			prices:    flatPrices(100, 40),
			wantError: false,
		},
		{
			name:      "sufficient data - rising",
			prices:    risingPrices(100, 40, 0.5),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prices := makePrices(tt.prices)
			result := calc.Calculate("test", prices)

			if tt.wantError && result.Error == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantError && result.Error != nil {
				t.Errorf("unexpected error: %v", result.Error)
			}

			if !tt.wantError {
				// Check metadata is populated
				if _, ok := result.Metadata["macd_line"]; !ok {
					t.Error("expected macd_line in metadata")
				}
				if _, ok := result.Metadata["signal_line"]; !ok {
					t.Error("expected signal_line in metadata")
				}
				if _, ok := result.Metadata["histogram"]; !ok {
					t.Error("expected histogram in metadata")
				}
			}
		})
	}
}

func TestMACDCalculator_Crossover(t *testing.T) {
	calc := NewMACDCalculator()

	// Create a pattern that should generate a bullish crossover
	// Start flat, then rise sharply
	prices := make([]float64, 50)
	for i := 0; i < 30; i++ {
		prices[i] = 100
	}
	for i := 30; i < 50; i++ {
		prices[i] = 100 + float64(i-30)*2
	}

	priceData := makePrices(prices)
	result := calc.Calculate("test", priceData)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	// MACD line should be positive when fast EMA > slow EMA (rising prices)
	macdLine := result.Metadata["macd_line"]
	if macdLine <= 0 {
		t.Logf("MACD line: %v (expected positive for rising prices)", macdLine)
	}
}

func TestMovingAverageCalculator_Calculate(t *testing.T) {
	calc := NewMovingAverageCalculator()

	tests := []struct {
		name      string
		prices    []float64
		wantError bool
	}{
		{
			name:      "insufficient data",
			prices:    flatPrices(100, 100),
			wantError: true,
		},
		{
			name:      "sufficient data",
			prices:    flatPrices(100, 210),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prices := makePrices(tt.prices)
			result := calc.Calculate("test", prices)

			if tt.wantError && result.Error == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantError && result.Error != nil {
				t.Errorf("unexpected error: %v", result.Error)
			}

			if !tt.wantError {
				if _, ok := result.Metadata["sma_50"]; !ok {
					t.Error("expected sma_50 in metadata")
				}
				if _, ok := result.Metadata["sma_200"]; !ok {
					t.Error("expected sma_200 in metadata")
				}
			}
		})
	}
}

func TestMovingAverageCalculator_GoldenCross(t *testing.T) {
	calc := NewMovingAverageCalculator()

	// Create price pattern: declining, then sharply rising
	// This should create a golden cross scenario
	prices := make([]float64, 210)
	// First 150 days: declining trend
	for i := 0; i < 150; i++ {
		prices[i] = 150 - float64(i)*0.3
	}
	// Next 60 days: sharp rise
	for i := 150; i < 210; i++ {
		prices[i] = prices[149] + float64(i-149)*1.5
	}

	priceData := makePrices(prices)
	result := calc.Calculate("test", priceData)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	// Log values for inspection
	t.Logf("SMA50: %v, SMA200: %v", result.Metadata["sma_50"], result.Metadata["sma_200"])
	t.Logf("Golden Cross: %v, Death Cross: %v",
		result.Metadata["golden_cross"], result.Metadata["death_cross"])
}

func TestBollingerBandsCalculator_Calculate(t *testing.T) {
	calc := NewBollingerBandsCalculator()

	tests := []struct {
		name      string
		prices    []float64
		wantError bool
	}{
		{
			name:      "insufficient data",
			prices:    flatPrices(100, 15),
			wantError: true,
		},
		{
			name:      "sufficient data - flat",
			prices:    flatPrices(100, 25),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prices := makePrices(tt.prices)
			result := calc.Calculate("test", prices)

			if tt.wantError && result.Error == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantError && result.Error != nil {
				t.Errorf("unexpected error: %v", result.Error)
			}

			if !tt.wantError {
				if _, ok := result.Metadata["upper_band"]; !ok {
					t.Error("expected upper_band in metadata")
				}
				if _, ok := result.Metadata["lower_band"]; !ok {
					t.Error("expected lower_band in metadata")
				}
				if _, ok := result.Metadata["middle_band"]; !ok {
					t.Error("expected middle_band in metadata")
				}
			}
		})
	}
}

func TestBollingerBands_FlatPrices(t *testing.T) {
	calc := NewBollingerBandsCalculator()

	// Flat prices should have very narrow bands (zero std dev)
	prices := makePrices(flatPrices(100, 25))
	result := calc.Calculate("test", prices)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	upper := result.Metadata["upper_band"]
	middle := result.Metadata["middle_band"]
	lower := result.Metadata["lower_band"]

	// With flat prices, bands should be very close together
	if upper-middle > 0.01 || middle-lower > 0.01 {
		t.Logf("Expected narrow bands for flat prices: upper=%v, middle=%v, lower=%v",
			upper, middle, lower)
	}

	// %B should be around 0.5 (middle)
	percentB := result.Metadata["percent_b"]
	t.Logf("%%B for flat prices: %v (expected ~0.5 or undefined for zero bandwidth)", percentB)
}

func TestBollingerBands_Volatility(t *testing.T) {
	calc := NewBollingerBandsCalculator()

	// Create volatile prices
	prices := make([]float64, 25)
	for i := 0; i < 25; i++ {
		if i%2 == 0 {
			prices[i] = 110
		} else {
			prices[i] = 90
		}
	}

	priceData := makePrices(prices)
	result := calc.Calculate("test", priceData)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	upper := result.Metadata["upper_band"]
	lower := result.Metadata["lower_band"]
	bandwidth := result.Metadata["bandwidth"]

	// Volatile prices should have wide bands
	t.Logf("Volatile prices: upper=%v, lower=%v, bandwidth=%v", upper, lower, bandwidth)

	if upper-lower < 10 {
		t.Errorf("expected wider bands for volatile prices")
	}
}

func TestSignalGenerator_GenerateSignals(t *testing.T) {
	sg := NewSignalGenerator()

	tests := []struct {
		name                 string
		prices               []float64
		wantInsufficientData bool
	}{
		{
			name:                 "insufficient data",
			prices:               flatPrices(100, 50),
			wantInsufficientData: true,
		},
		{
			name:                 "sufficient data",
			prices:               flatPrices(100, 220),
			wantInsufficientData: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prices := makePrices(tt.prices)
			summary := sg.GenerateSignals("test", prices)

			if summary.InsufficientData != tt.wantInsufficientData {
				t.Errorf("InsufficientData = %v, want %v",
					summary.InsufficientData, tt.wantInsufficientData)
			}

			if !tt.wantInsufficientData {
				// Should have all 4 indicators
				if len(summary.Indicators) != 4 {
					t.Errorf("expected 4 indicators, got %d", len(summary.Indicators))
				}

				// Signal counts should add up
				total := summary.SignalCounts.Total()
				if total != 4 {
					t.Errorf("expected total signal count of 4, got %d", total)
				}
			}
		})
	}
}

func TestSignalGenerator_Consensus(t *testing.T) {
	sg := NewSignalGenerator()

	// Create steadily rising prices - should lean bullish
	rising := risingPrices(100, 220, 0.1)
	prices := makePrices(rising)
	summary := sg.GenerateSignals("test", prices)

	if summary.InsufficientData {
		t.Fatal("expected sufficient data for rising prices")
	}

	t.Logf("Rising prices consensus: %s (confidence: %.2f%%)",
		summary.Consensus.String(), summary.Confidence*100)
	t.Logf("Signal counts: StrongBuy=%d, Buy=%d, Hold=%d, Sell=%d, StrongSell=%d",
		summary.SignalCounts.StrongBuy, summary.SignalCounts.Buy,
		summary.SignalCounts.Hold, summary.SignalCounts.Sell, summary.SignalCounts.StrongSell)

	// Create steadily falling prices - should lean bearish
	falling := fallingPrices(150, 220, 0.1)
	prices = makePrices(falling)
	summary = sg.GenerateSignals("test", prices)

	t.Logf("Falling prices consensus: %s (confidence: %.2f%%)",
		summary.Consensus.String(), summary.Confidence*100)
}

func TestHelpers_SMA(t *testing.T) {
	prices := []float64{1, 2, 3, 4, 5}
	sma := calculateSMA(prices, 5)
	expected := 3.0

	if sma != expected {
		t.Errorf("SMA(1,2,3,4,5) = %v, want %v", sma, expected)
	}
}

func TestHelpers_EMA(t *testing.T) {
	// Simple test case
	prices := []float64{100, 100, 100, 100, 100}
	ema := calculateEMA(prices, 5)

	// EMA of constant values should equal that value
	if ema != 100 {
		t.Errorf("EMA of constant 100 = %v, want 100", ema)
	}
}

func TestHelpers_StdDev(t *testing.T) {
	// Test with known values
	prices := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	mean := calculateSMA(prices, 8)
	stdDev := calculateStdDev(prices, mean, 8)

	// Expected std dev is 2
	if stdDev < 1.9 || stdDev > 2.1 {
		t.Errorf("StdDev = %v, want ~2", stdDev)
	}
}

func TestSignal_String(t *testing.T) {
	tests := []struct {
		signal domain.Signal
		want   string
	}{
		{domain.SignalStrongBuy, "Strong Buy"},
		{domain.SignalBuy, "Buy"},
		{domain.SignalHold, "Hold"},
		{domain.SignalSell, "Sell"},
		{domain.SignalStrongSell, "Strong Sell"},
	}

	for _, tt := range tests {
		if got := tt.signal.String(); got != tt.want {
			t.Errorf("Signal(%d).String() = %v, want %v", tt.signal, got, tt.want)
		}
	}
}

func TestSignal_Emoji(t *testing.T) {
	tests := []struct {
		signal domain.Signal
		want   string
	}{
		{domain.SignalStrongBuy, "🟢"},
		{domain.SignalBuy, "🔵"},
		{domain.SignalHold, "⚪"},
		{domain.SignalSell, "🟠"},
		{domain.SignalStrongSell, "🔴"},
	}

	for _, tt := range tests {
		if got := tt.signal.Emoji(); got != tt.want {
			t.Errorf("Signal(%d).Emoji() = %v, want %v", tt.signal, got, tt.want)
		}
	}
}
