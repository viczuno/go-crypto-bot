// Package domain contains the core business types and interfaces for the crypto bot.
package domain

// Constants for time-based thresholds used in price change calculations.
// These define minimum data requirements for valid historical comparisons.
const (
	// Hours24Threshold is the minimum hours of data required for valid 24h comparison.
	Hours24Threshold = 20.0

	// Days7Threshold is the minimum days of data required for valid 7d comparison.
	Days7Threshold = 6.5

	// Days30Threshold is the minimum days of data required for valid 30d comparison.
	Days30Threshold = 28.0

	// Days30Fallback is the fallback threshold when exact 30d data is unavailable.
	Days30Fallback = 25.0
)

// History period constants for standardized time periods.
const (
	Days7  = 7
	Days30 = 30
)
