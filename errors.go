package wirepusher

import (
	"fmt"
)

// Error represents a general WirePusher API error.
type Error struct {
	Message    string
	StatusCode int
}

func (e *Error) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("wirepusher: %s (status: %d)", e.Message, e.StatusCode)
	}
	return fmt.Sprintf("wirepusher: %s", e.Message)
}

// IsRetryable returns false for generic errors.
// Use ServerError or NetworkError for retryable errors.
func (e *Error) IsRetryable() bool {
	return false
}

// ServerError represents a server error (5xx).
type ServerError struct {
	Message    string
	StatusCode int
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("wirepusher server error: %s (status: %d)", e.Message, e.StatusCode)
}

// IsRetryable returns true - server errors should be retried.
func (e *ServerError) IsRetryable() bool {
	return true
}

// NetworkError represents a network or connection error.
type NetworkError struct {
	Message string
	Err     error // Original error
}

func (e *NetworkError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("wirepusher network error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("wirepusher network error: %s", e.Message)
}

// IsRetryable returns true - network errors should be retried.
func (e *NetworkError) IsRetryable() bool {
	return true
}

// Unwrap returns the original error for error chain support.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// AuthError represents an authentication error (401/403).
type AuthError struct {
	Message    string
	StatusCode int
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("wirepusher auth error: %s (status: %d)", e.Message, e.StatusCode)
}

// IsRetryable returns false - authentication errors are not retryable.
func (e *AuthError) IsRetryable() bool {
	return false
}

// ValidationError represents a validation error (400).
type ValidationError struct {
	Message    string
	StatusCode int
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("wirepusher validation error: %s (status: %d)", e.Message, e.StatusCode)
}

// IsRetryable returns false - validation errors are not retryable.
func (e *ValidationError) IsRetryable() bool {
	return false
}

// RateLimitError represents a rate limit error (429).
type RateLimitError struct {
	Message    string
	StatusCode int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("wirepusher rate limit error: %s (status: %d)", e.Message, e.StatusCode)
}

// IsRetryable returns true - rate limit errors should be retried with backoff.
func (e *RateLimitError) IsRetryable() bool {
	return true
}

// RetryableError is an interface for errors that can be retried.
type RetryableError interface {
	error
	IsRetryable() bool
}

// IsErrorRetryable checks if an error should be retried.
func IsErrorRetryable(err error) bool {
	if retryable, ok := err.(RetryableError); ok {
		return retryable.IsRetryable()
	}
	return false
}
