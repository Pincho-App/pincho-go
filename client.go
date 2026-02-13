// Package pincho provides a Go client for the Pincho push notification API.
//
// Example usage:
//
//	// Create client with token
//	client := pincho.NewClient("abc12345")
//
//	// Simple send
//	err := client.SendSimple(ctx, "Hello", "World")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Advanced send with options
//	err = client.Send(ctx, &pincho.SendOptions{
//	    Title:     "Server Alert",
//	    Message:   "CPU usage high",
//	    Type:      "alert",
//	    Tags:      []string{"monitoring", "production"},
//	    ActionURL: "https://dashboard.example.com",
//	})
package pincho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	// Version is the library version.
	Version = "1.0.0-alpha.7"

	// DefaultAPIURL is the default Pincho API endpoint.
	DefaultAPIURL = "https://api.pincho.app/send"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second

	// MaxBackoff is the maximum backoff duration for retries.
	MaxBackoff = 30 * time.Second
)

// Client is the Pincho API client.
type Client struct {
	// Token is the Pincho API token.
	Token string

	// APIURL is the Pincho API endpoint (defaults to DefaultAPIURL).
	APIURL string

	// HTTPClient is the HTTP client used for requests.
	// Can be customized to use different timeouts, proxies, etc.
	HTTPClient *http.Client

	// MaxRetries is the maximum number of retry attempts for failed requests.
	// Defaults to 3. Set to 0 to disable retries.
	MaxRetries int

	// Logger is the logger for debug/info messages.
	// Defaults to NoOpLogger (no logging). Use WithLogger() to enable logging.
	Logger Logger

	// LastRateLimit contains rate limit information from the most recent API response.
	// This is updated after each successful request.
	LastRateLimit *RateLimitInfo
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithAPIURL sets a custom API URL.
// The URL must not be empty.
func WithAPIURL(url string) ClientOption {
	return func(c *Client) {
		if url == "" {
			panic("pincho: API URL cannot be empty")
		}
		c.APIURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
// The client must not be nil.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		if client == nil {
			panic("pincho: HTTP client cannot be nil")
		}
		c.HTTPClient = client
	}
}

// WithTimeout sets a custom HTTP timeout.
// The timeout must be positive.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if timeout <= 0 {
			panic("pincho: timeout must be positive")
		}
		c.HTTPClient.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
// Set to 0 to disable retries. Negative values are not allowed.
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		if maxRetries < 0 {
			panic("pincho: max retries cannot be negative")
		}
		c.MaxRetries = maxRetries
	}
}

// NewClient creates a new Pincho client.
//
// The token parameter is your Pincho API token. If empty, it reads from
// PINCHO_TOKEN environment variable.
//
// Environment variables (used as defaults):
//   - PINCHO_TOKEN: API token (required if token param is empty)
//   - PINCHO_TIMEOUT: Request timeout in seconds (default: 30)
//   - PINCHO_MAX_RETRIES: Maximum retry attempts (default: 3)
//
// Examples:
//
//	// Auto-load token from PINCHO_TOKEN
//	client := pincho.NewClient("")
//
//	// Explicit token
//	client := pincho.NewClient("abc12345")
//
//	// With custom timeout
//	client := pincho.NewClient(
//	    "abc12345",
//	    pincho.WithTimeout(10*time.Second),
//	)
//
//	// With custom HTTP client
//	client := pincho.NewClient(
//	    "abc12345",
//	    pincho.WithHTTPClient(customHTTPClient),
//	)
func NewClient(token string, opts ...ClientOption) *Client {
	// Auto-load token from environment variable if not provided
	if token == "" {
		token = os.Getenv("PINCHO_TOKEN")
	}
	if token == "" {
		panic("pincho: token is required (set PINCHO_TOKEN or pass to NewClient)")
	}

	// Default timeout (can be overridden by env var or option)
	timeout := DefaultTimeout
	if envTimeout := os.Getenv("PINCHO_TIMEOUT"); envTimeout != "" {
		if seconds, err := strconv.Atoi(envTimeout); err == nil && seconds > 0 {
			timeout = time.Duration(seconds) * time.Second
		}
	}

	// Default max retries (can be overridden by env var or option)
	maxRetries := 3
	if envRetries := os.Getenv("PINCHO_MAX_RETRIES"); envRetries != "" {
		if retries, err := strconv.Atoi(envRetries); err == nil && retries >= 0 {
			maxRetries = retries
		}
	}

	client := &Client{
		Token:  token,
		APIURL: DefaultAPIURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		MaxRetries: maxRetries,
		Logger:     &NoOpLogger{}, // Default: no logging
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// retryWithBackoff executes a function with exponential backoff retry logic.
// It retries on retryable errors (network errors, 5xx, 429) up to maxRetries times.
// For rate limit errors (429), it uses longer backoff periods.
func (c *Client) retryWithBackoff(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		// Log attempt
		if attempt > 0 {
			c.logDebug(fmt.Sprintf("Retry attempt %d/%d", attempt, c.MaxRetries))
		}

		// Execute the operation
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsErrorRetryable(err) {
			c.logDebug(fmt.Sprintf("Error not retryable: %v", err))
			return err
		}

		// Don't retry if we've exhausted all attempts
		if attempt == c.MaxRetries {
			c.logWarning(fmt.Sprintf("Max retries (%d) exceeded: %v", c.MaxRetries, err))
			return err
		}

		// Calculate backoff duration
		var backoff time.Duration
		if rateLimitErr, isRateLimit := err.(*RateLimitError); isRateLimit {
			// Use Retry-After header value if provided by server
			if rateLimitErr.RetryAfter > 0 {
				backoff = time.Duration(rateLimitErr.RetryAfter) * time.Second
				if backoff > MaxBackoff {
					backoff = MaxBackoff
				}
				c.logWarning(fmt.Sprintf("Rate limit hit, using server Retry-After: %s", backoff))
			} else {
				// Rate limit: use longer backoff (5s, 10s, 20s, capped at 30s)
				backoff = time.Duration(5*(1<<uint(attempt))) * time.Second
				if backoff > MaxBackoff {
					backoff = MaxBackoff
				}
				c.logWarning(fmt.Sprintf("Rate limit hit, backing off for %s", backoff))
			}
		} else {
			// Network/server error: exponential backoff (1s, 2s, 4s, 8s, capped at 30s)
			backoff = time.Duration(1<<uint(attempt)) * time.Second
			if backoff > MaxBackoff {
				backoff = MaxBackoff
			}
			c.logDebug(fmt.Sprintf("Retryable error, backing off for %s: %v", backoff, err))
		}

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			c.logDebug("Context cancelled during retry backoff")
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next retry
		}
	}

	return lastErr
}

// SendSimple sends a simple notification with just a title and message.
//
// This is a convenience method that wraps Send() with minimal options.
//
// Example:
//
//	err := client.SendSimple(ctx, "Hello", "World")
func (c *Client) SendSimple(ctx context.Context, title, message string) error {
	return c.Send(ctx, &SendOptions{
		Title:   title,
		Message: message,
	})
}

// Send sends a notification with full options.
//
// The options parameter must include at least Title and Message.
// Optional fields include Type, Tags, ImageURL, and ActionURL.
//
// Example:
//
//	err := client.Send(ctx, &pincho.SendOptions{
//	    Title:     "Server Alert",
//	    Message:   "CPU usage at 95%",
//	    Type:      "alert",
//	    Tags:      []string{"monitoring", "production"},
//	    ImageURL:  "https://example.com/graph.png",
//	    ActionURL: "https://dashboard.example.com",
//	})
func (c *Client) Send(ctx context.Context, options *SendOptions) error {
	if options == nil {
		return &ValidationError{Message: "options cannot be nil", StatusCode: 0}
	}

	c.logDebug(fmt.Sprintf("Send() called with title: %s", options.Title))

	if options.Title == "" {
		return &ValidationError{Message: "title is required", StatusCode: 0}
	}

	// Normalize tags
	normalizedTags := NormalizeTags(options.Tags)
	if normalizedTags != nil && len(normalizedTags) != len(options.Tags) {
		c.logDebug(fmt.Sprintf("Tags normalized: %v -> %v", options.Tags, normalizedTags))
	}

	// Handle encryption if password provided
	// Encrypted fields: title, message, imageURL, actionURL
	// NOT encrypted: type, tags (needed for filtering/routing)
	finalTitle := options.Title
	finalMessage := options.Message
	finalImageURL := options.ImageURL
	finalActionURL := options.ActionURL
	var ivHex string

	if options.EncryptionPassword != "" {
		c.logDebug("Encrypting title, message, imageURL, actionURL")
		iv, ivStr, err := GenerateIV()
		if err != nil {
			c.logError(fmt.Sprintf("Failed to generate IV: %v", err))
			return &Error{Message: fmt.Sprintf("failed to generate IV: %v", err), StatusCode: 0}
		}

		encryptedTitle, err := EncryptMessage(options.Title, options.EncryptionPassword, iv)
		if err != nil {
			c.logError(fmt.Sprintf("Failed to encrypt title: %v", err))
			return &Error{Message: fmt.Sprintf("failed to encrypt title: %v", err), StatusCode: 0}
		}
		finalTitle = encryptedTitle

		encryptedMessage, err := EncryptMessage(options.Message, options.EncryptionPassword, iv)
		if err != nil {
			c.logError(fmt.Sprintf("Failed to encrypt message: %v", err))
			return &Error{Message: fmt.Sprintf("failed to encrypt message: %v", err), StatusCode: 0}
		}
		finalMessage = encryptedMessage

		if options.ImageURL != "" {
			encryptedImageURL, err := EncryptMessage(options.ImageURL, options.EncryptionPassword, iv)
			if err != nil {
				c.logError(fmt.Sprintf("Failed to encrypt imageURL: %v", err))
				return &Error{Message: fmt.Sprintf("failed to encrypt imageURL: %v", err), StatusCode: 0}
			}
			finalImageURL = encryptedImageURL
		}

		if options.ActionURL != "" {
			encryptedActionURL, err := EncryptMessage(options.ActionURL, options.EncryptionPassword, iv)
			if err != nil {
				c.logError(fmt.Sprintf("Failed to encrypt actionURL: %v", err))
				return &Error{Message: fmt.Sprintf("failed to encrypt actionURL: %v", err), StatusCode: 0}
			}
			finalActionURL = encryptedActionURL
		}

		ivHex = ivStr
	}

	// Build request body
	body := map[string]interface{}{
		"title":   finalTitle,
		"message": finalMessage,
	}

	if options.Type != "" {
		body["type"] = options.Type
	}
	if normalizedTags != nil {
		body["tags"] = normalizedTags
	}
	if finalImageURL != "" {
		body["imageURL"] = finalImageURL
	}
	if finalActionURL != "" {
		body["actionURL"] = finalActionURL
	}
	if ivHex != "" {
		body["iv"] = ivHex
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return &Error{Message: fmt.Sprintf("failed to marshal request: %v", err), StatusCode: 0}
	}

	// Wrap HTTP request in retry logic
	return c.retryWithBackoff(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", c.APIURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return &NetworkError{Message: "failed to create request", Err: err}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("User-Agent", "pincho-go/"+Version)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return &NetworkError{Message: "request failed", Err: err}
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return &NetworkError{Message: "failed to read response", Err: err}
		}

		// Handle non-2xx status codes
		if resp.StatusCode >= 400 {
			var errorMsg string

			// Try to parse nested error response
			var errorResp ErrorResponse
			if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error.Message != "" {
				// Format error message with details
				errorMsg = errorResp.Error.Message
				if errorResp.Error.Param != "" {
					errorMsg = fmt.Sprintf("%s (parameter: %s)", errorMsg, errorResp.Error.Param)
				}
				if errorResp.Error.Code != "" {
					errorMsg = fmt.Sprintf("%s [%s]", errorMsg, errorResp.Error.Code)
				}
			} else {
				// Fallback to raw response if parsing fails
				errorMsg = string(bodyBytes)
			}

			switch resp.StatusCode {
			case 400:
				return &ValidationError{Message: errorMsg, StatusCode: resp.StatusCode}
			case 401, 403:
				return &AuthError{Message: errorMsg, StatusCode: resp.StatusCode}
			case 429:
				retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
				return &RateLimitError{Message: errorMsg, StatusCode: resp.StatusCode, RetryAfter: retryAfter}
			default:
				if resp.StatusCode >= 500 {
					return &ServerError{Message: errorMsg, StatusCode: resp.StatusCode}
				}
				return &Error{Message: errorMsg, StatusCode: resp.StatusCode}
			}
		}

		// Parse rate limit headers from successful response
		c.parseRateLimitHeaders(resp)

		// Parse success response (optional)
		var apiResponse SendResponse
		if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
			// Non-fatal: response was successful but couldn't parse
			return nil
		}

		return nil
	})
}

// NotifAI generates and sends an AI-powered notification from free-form text.
//
// The NotifAI endpoint uses AI (Gemini) to convert natural language into a
// structured notification with title, message, type, and tags.
//
// Example:
//
//	response, err := client.NotifAI(ctx, &pincho.NotifAIOptions{
//	    Text: "deployment finished successfully, v2.1.3 is live on prod",
//	    Type: "deployment", // Optional override
//	})
func (c *Client) NotifAI(ctx context.Context, options *NotifAIOptions) (*NotifAIResponse, error) {
	if options == nil {
		return nil, &ValidationError{Message: "options cannot be nil", StatusCode: 0}
	}

	c.logDebug(fmt.Sprintf("NotifAI() called with text: %s", options.Text))

	if options.Text == "" {
		return nil, &ValidationError{Message: "text is required", StatusCode: 0}
	}

	// Build request body
	body := map[string]interface{}{
		"text": options.Text,
	}

	if options.Type != "" {
		body["type"] = options.Type
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("failed to marshal request: %v", err), StatusCode: 0}
	}

	// Build NotifAI endpoint URL using proper URL parsing
	baseURL := c.APIURL
	// Remove "/send" suffix if present
	if len(baseURL) >= 5 && baseURL[len(baseURL)-5:] == "/send" {
		baseURL = baseURL[:len(baseURL)-5]
	}
	// Ensure trailing slash
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}
	apiURL := baseURL + "notifai"

	// Capture response outside retry closure
	var apiResponse NotifAIResponse

	// Wrap HTTP request in retry logic
	err = c.retryWithBackoff(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return &NetworkError{Message: "failed to create request", Err: err}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("User-Agent", "pincho-go/"+Version)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return &NetworkError{Message: "request failed", Err: err}
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return &NetworkError{Message: "failed to read response", Err: err}
		}

		// Handle non-2xx status codes
		if resp.StatusCode >= 400 {
			var errorMsg string

			// Try to parse nested error response
			var errorResp ErrorResponse
			if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error.Message != "" {
				// Format error message with details
				errorMsg = errorResp.Error.Message
				if errorResp.Error.Param != "" {
					errorMsg = fmt.Sprintf("%s (parameter: %s)", errorMsg, errorResp.Error.Param)
				}
				if errorResp.Error.Code != "" {
					errorMsg = fmt.Sprintf("%s [%s]", errorMsg, errorResp.Error.Code)
				}
			} else {
				// Fallback to raw response if parsing fails
				errorMsg = string(bodyBytes)
			}

			switch resp.StatusCode {
			case 400:
				return &ValidationError{Message: errorMsg, StatusCode: resp.StatusCode}
			case 401, 403:
				return &AuthError{Message: errorMsg, StatusCode: resp.StatusCode}
			case 429:
				retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
				return &RateLimitError{Message: errorMsg, StatusCode: resp.StatusCode, RetryAfter: retryAfter}
			default:
				if resp.StatusCode >= 500 {
					return &ServerError{Message: errorMsg, StatusCode: resp.StatusCode}
				}
				return &Error{Message: errorMsg, StatusCode: resp.StatusCode}
			}
		}

		// Parse rate limit headers from successful response
		c.parseRateLimitHeaders(resp)

		// Parse success response
		if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
			// Non-fatal: response was successful but couldn't parse
			return nil
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

// parseRetryAfter parses the Retry-After header value (in seconds).
// Returns 0 if the header is missing or invalid.
func parseRetryAfter(header string) int {
	if header == "" {
		return 0
	}
	seconds, err := strconv.Atoi(header)
	if err != nil || seconds < 0 {
		return 0
	}
	return seconds
}

// parseIntHeader parses an integer header value.
// Returns 0 if the header is missing or invalid.
func parseIntHeader(header string) int {
	if header == "" {
		return 0
	}
	value, err := strconv.Atoi(header)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

// parseUnixTimestamp parses a Unix timestamp header value.
// Returns zero time if the header is missing or invalid.
func parseUnixTimestamp(header string) time.Time {
	if header == "" {
		return time.Time{}
	}
	timestamp, err := strconv.ParseInt(header, 10, 64)
	if err != nil || timestamp < 0 {
		return time.Time{}
	}
	return time.Unix(timestamp, 0)
}

// parseRateLimitHeaders parses rate limit headers from the response and updates LastRateLimit.
func (c *Client) parseRateLimitHeaders(resp *http.Response) {
	limit := parseIntHeader(resp.Header.Get("RateLimit-Limit"))
	remaining := parseIntHeader(resp.Header.Get("RateLimit-Remaining"))
	reset := parseUnixTimestamp(resp.Header.Get("RateLimit-Reset"))

	// Only update if at least one header is present
	if limit > 0 || remaining > 0 || !reset.IsZero() {
		c.LastRateLimit = &RateLimitInfo{
			Limit:     limit,
			Remaining: remaining,
			Reset:     reset,
		}
	}
}

// GetRateLimitInfo returns the rate limit information from the most recent API response.
// Returns nil if no rate limit information is available.
func (c *Client) GetRateLimitInfo() *RateLimitInfo {
	return c.LastRateLimit
}
