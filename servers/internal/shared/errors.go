package shared

import (
	"errors"
	"fmt"
)

// Common error definitions
var (
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrInvalidInput       = errors.New("invalid input provided")
	ErrNotFound           = errors.New("resource not found")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInternalServer     = errors.New("internal server error")
)

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsRetryableError determines if an error can be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific retryable errors
	if errors.Is(err, ErrDatabaseConnection) {
		return true
	}

	// Add more retryable error checks as needed
	return false
}
