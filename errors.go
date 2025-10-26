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

// AuthError represents an authentication error (401/403).
type AuthError struct {
	Message    string
	StatusCode int
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("wirepusher auth error: %s (status: %d)", e.Message, e.StatusCode)
}

// ValidationError represents a validation error (400).
type ValidationError struct {
	Message    string
	StatusCode int
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("wirepusher validation error: %s (status: %d)", e.Message, e.StatusCode)
}

// RateLimitError represents a rate limit error (429).
type RateLimitError struct {
	Message    string
	StatusCode int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("wirepusher rate limit error: %s (status: %d)", e.Message, e.StatusCode)
}
