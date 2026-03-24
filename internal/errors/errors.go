// Package errors provides custom error types for the crypto bot.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
var (
	// ErrNoPriceData indicates no price data is available for the requested coin/period.
	ErrNoPriceData = errors.New("no price data available")

	// ErrEmptyInput indicates an empty or nil input was provided.
	ErrEmptyInput = errors.New("empty input provided")

	// ErrRateLimited indicates the API rate limit was exceeded.
	ErrRateLimited = errors.New("rate limit exceeded")
)

// NotFoundError represents a resource not found error.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// NewNotFoundError creates a new NotFoundError.
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id}
}

// IsNotFound checks if an error is a NotFoundError.
func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}

// APIError represents an error from an external API call.
type APIError struct {
	StatusCode int
	Message    string
	Cause      error
}

func (e *APIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("API error (status %d): %s: %v", e.StatusCode, e.Message, e.Cause)
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Cause
}

// NewAPIError creates a new APIError.
func NewAPIError(statusCode int, message string, cause error) *APIError {
	return &APIError{StatusCode: statusCode, Message: message, Cause: cause}
}

// IsAPIError checks if an error is an APIError.
func IsAPIError(err error) bool {
	var ae *APIError
	return errors.As(err, &ae)
}

// DatabaseError represents a database operation error.
type DatabaseError struct {
	Operation string
	Cause     error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Cause)
}

func (e *DatabaseError) Unwrap() error {
	return e.Cause
}

// NewDatabaseError creates a new DatabaseError.
func NewDatabaseError(operation string, cause error) *DatabaseError {
	return &DatabaseError{Operation: operation, Cause: cause}
}

// IsDatabaseError checks if an error is a DatabaseError.
func IsDatabaseError(err error) bool {
	var de *DatabaseError
	return errors.As(err, &de)
}

// ValidationError represents an input validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// IsValidationError checks if an error is a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
