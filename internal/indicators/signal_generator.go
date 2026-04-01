// Package indicators provides technical indicator calculations for crypto analysis.
package indicators

import (
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// SignalGenerator aggregates multiple indicators and generates consensus signals.
type SignalGenerator struct {
	calculators []weightedCalculator
}

type weightedCalculator struct {
	calculator domain.IndicatorCalculator
	weight     float64
}

// NewSignalGenerator creates a new signal generator with all standard indicators.
func NewSignalGenerator() *SignalGenerator {
	return &SignalGenerator{
		calculators: []weightedCalculator{
			{calculator: NewMACDCalculator(), weight: domain.WeightMACD},
			{calculator: NewRSICalculator(), weight: domain.WeightRSI},
			{calculator: NewMovingAverageCalculator(), weight: domain.WeightMA},
			{calculator: NewBollingerBandsCalculator(), weight: domain.WeightBollinger},
		},
	}
}

var _ domain.SignalGenerator = (*SignalGenerator)(nil)

// GenerateSignals calculates all indicators and returns an aggregated summary.
func (sg *SignalGenerator) GenerateSignals(coinID string, prices []domain.CryptoPrice) domain.IndicatorSummary {
	summary := domain.IndicatorSummary{
		CoinID:       coinID,
		Indicators:   make([]domain.IndicatorResult, 0, len(sg.calculators)),
		CalculatedAt: time.Now().UTC(),
	}

	// Check if we have enough data for any indicator
	maxRequired := sg.maxDataPointsRequired()
	if len(prices) < maxRequired {
		summary.InsufficientData = true
		summary.Consensus = domain.SignalHold
		summary.Confidence = 0.0
		return summary
	}

	// Calculate all indicators
	for _, wc := range sg.calculators {
		result := wc.calculator.Calculate(coinID, prices)
		summary.Indicators = append(summary.Indicators, result)
		sg.updateSignalCounts(&summary.SignalCounts, result.Signal)
	}

	// Calculate consensus signal
	summary.Consensus, summary.Confidence = sg.calculateConsensus(summary.Indicators)

	return summary
}

func (sg *SignalGenerator) maxDataPointsRequired() int {
	max := 0
	for _, wc := range sg.calculators {
		if wc.calculator.MinDataPoints() > max {
			max = wc.calculator.MinDataPoints()
		}
	}
	return max
}

func (sg *SignalGenerator) updateSignalCounts(counts *domain.SignalCounts, signal domain.Signal) {
	switch signal {
	case domain.SignalStrongBuy:
		counts.StrongBuy++
	case domain.SignalBuy:
		counts.Buy++
	case domain.SignalHold:
		counts.Hold++
	case domain.SignalSell:
		counts.Sell++
	case domain.SignalStrongSell:
		counts.StrongSell++
	}
}

func (sg *SignalGenerator) calculateConsensus(results []domain.IndicatorResult) (domain.Signal, float64) {
	if len(results) == 0 {
		return domain.SignalHold, 0.0
	}

	// Calculate weighted signal score
	// Strong Buy = +2, Buy = +1, Hold = 0, Sell = -1, Strong Sell = -2
	var weightedScore float64
	var totalWeight float64
	var weightedConfidence float64

	for i, result := range results {
		if result.Error != nil {
			continue
		}

		weight := sg.calculators[i].weight
		score := float64(result.Signal)

		weightedScore += score * weight
		totalWeight += weight
		weightedConfidence += result.Confidence * weight
	}

	if totalWeight == 0 {
		return domain.SignalHold, 0.0
	}

	averageScore := weightedScore / totalWeight
	averageConfidence := weightedConfidence / totalWeight

	// Convert average score to discrete signal
	consensus := sg.scoreToSignal(averageScore)

	// Adjust confidence based on agreement between indicators
	agreementFactor := sg.calculateAgreement(results, consensus)
	finalConfidence := averageConfidence * agreementFactor

	return consensus, finalConfidence
}

func (sg *SignalGenerator) scoreToSignal(score float64) domain.Signal {
	switch {
	case score >= 1.5:
		return domain.SignalStrongBuy
	case score >= 0.5:
		return domain.SignalBuy
	case score <= -1.5:
		return domain.SignalStrongSell
	case score <= -0.5:
		return domain.SignalSell
	default:
		return domain.SignalHold
	}
}

func (sg *SignalGenerator) calculateAgreement(results []domain.IndicatorResult, consensus domain.Signal) float64 {
	if len(results) == 0 {
		return 1.0
	}

	agreementCount := 0
	validCount := 0

	for _, result := range results {
		if result.Error != nil {
			continue
		}
		validCount++

		// Check if indicator agrees with consensus direction
		// Strong signals in same direction count as agreement
		consensusDirection := signDirection(consensus)
		indicatorDirection := signDirection(result.Signal)

		if consensusDirection == indicatorDirection || consensusDirection == 0 || indicatorDirection == 0 {
			agreementCount++
		}
	}

	if validCount == 0 {
		return 1.0
	}

	// Agreement factor: 0.5 (all disagree) to 1.0 (all agree)
	return 0.5 + (0.5 * float64(agreementCount) / float64(validCount))
}

func signDirection(signal domain.Signal) int {
	switch signal {
	case domain.SignalStrongBuy, domain.SignalBuy:
		return 1
	case domain.SignalStrongSell, domain.SignalSell:
		return -1
	default:
		return 0
	}
}
