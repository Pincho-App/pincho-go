// Package wirepusher provides a Go client for the WirePusher push notification API.
//
// Example usage:
//
//	// Personal notifications
//	client := wirepusher.NewClient("", "your-user-id")
//
//	// Team notifications
//	client := wirepusher.NewClient("wpt_your_token", "")
//
//	// Simple send
//	err := client.SendSimple(ctx, "Hello", "World")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Advanced send with options
//	err = client.Send(ctx, &wirepusher.SendOptions{
//	    Title:     "Server Alert",
//	    Message:   "CPU usage high",
//	    Type:      "alert",
//	    Tags:      []string{"monitoring", "production"},
//	    ActionURL: "https://dashboard.example.com",
//	})
package wirepusher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultAPIURL is the default WirePusher API endpoint.
	DefaultAPIURL = "https://wirepusher.com/send"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second
)

// Client is the WirePusher API client.
type Client struct {
	// Token is the WirePusher team token (mutually exclusive with UserID).
	// Use this for team notifications (starts with "wpt_").
	Token string

	// UserID is the WirePusher user ID (mutually exclusive with Token).
	// Use this for personal notifications.
	UserID string

	// APIURL is the WirePusher API endpoint (defaults to DefaultAPIURL).
	APIURL string

	// HTTPClient is the HTTP client used for requests.
	// Can be customized to use different timeouts, proxies, etc.
	HTTPClient *http.Client
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithAPIURL sets a custom API URL.
func WithAPIURL(url string) ClientOption {
	return func(c *Client) {
		c.APIURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = client
	}
}

// WithTimeout sets a custom HTTP timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Timeout = timeout
	}
}

// NewClient creates a new WirePusher client.
//
// You must specify EITHER token OR userID, not both:
//   - token: Team token (starts with "wpt_") for team-wide notifications
//   - userID: User ID for personal notifications
//
// Panics if both token and userID are provided, or if neither is provided.
//
// Examples:
//
//	// Personal notifications
//	client := wirepusher.NewClient("", "user_abc123")
//
//	// Team notifications
//	client := wirepusher.NewClient("wpt_abc123...", "")
//
//	// With custom timeout
//	client := wirepusher.NewClient(
//	    "",
//	    "user_abc123",
//	    wirepusher.WithTimeout(10*time.Second),
//	)
func NewClient(token, userID string, opts ...ClientOption) *Client {
	// Validate mutual exclusivity
	if token != "" && userID != "" {
		panic("wirepusher: cannot specify both token and userID - they are mutually exclusive. Use token for team notifications or userID for personal notifications")
	}
	if token == "" && userID == "" {
		panic("wirepusher: must specify either token or userID. Use token for team notifications or userID for personal notifications")
	}

	client := &Client{
		Token:  token,
		UserID: userID,
		APIURL: DefaultAPIURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
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
//	err := client.Send(ctx, &wirepusher.SendOptions{
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

	if options.Title == "" {
		return &ValidationError{Message: "title is required", StatusCode: 0}
	}

	if options.Message == "" {
		return &ValidationError{Message: "message is required", StatusCode: 0}
	}

	// Handle encryption if password provided
	finalMessage := options.Message
	var ivHex string

	if options.EncryptionPassword != "" {
		iv, ivStr, err := generateIV()
		if err != nil {
			return &Error{Message: fmt.Sprintf("failed to generate IV: %v", err), StatusCode: 0}
		}

		encryptedMessage, err := encryptMessage(options.Message, options.EncryptionPassword, iv)
		if err != nil {
			return &Error{Message: fmt.Sprintf("failed to encrypt message: %v", err), StatusCode: 0}
		}

		finalMessage = encryptedMessage
		ivHex = ivStr
	}

	// Build request body
	body := map[string]interface{}{
		"title":   options.Title,
		"message": finalMessage,
		"id":      c.UserID,
		"token":   c.Token,
	}

	if options.Type != "" {
		body["type"] = options.Type
	}
	if len(options.Tags) > 0 {
		body["tags"] = options.Tags
	}
	if options.ImageURL != "" {
		body["imageURL"] = options.ImageURL
	}
	if options.ActionURL != "" {
		body["actionURL"] = options.ActionURL
	}
	if ivHex != "" {
		body["iv"] = ivHex
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return &Error{Message: fmt.Sprintf("failed to marshal request: %v", err), StatusCode: 0}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return &Error{Message: fmt.Sprintf("failed to create request: %v", err), StatusCode: 0}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return &Error{Message: fmt.Sprintf("request failed: %v", err), StatusCode: 0}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Error{Message: fmt.Sprintf("failed to read response: %v", err), StatusCode: resp.StatusCode}
	}

	// Handle non-2xx status codes
	if resp.StatusCode >= 400 {
		var errorMsg string

		// Try to parse error response
		var apiResponse SendResponse
		if err := json.Unmarshal(bodyBytes, &apiResponse); err == nil && apiResponse.Message != "" {
			errorMsg = apiResponse.Message
		} else {
			errorMsg = string(bodyBytes)
		}

		switch resp.StatusCode {
		case 400:
			return &ValidationError{Message: errorMsg, StatusCode: resp.StatusCode}
		case 401, 403:
			return &AuthError{Message: errorMsg, StatusCode: resp.StatusCode}
		case 429:
			return &RateLimitError{Message: errorMsg, StatusCode: resp.StatusCode}
		default:
			return &Error{Message: errorMsg, StatusCode: resp.StatusCode}
		}
	}

	// Parse success response (optional)
	var apiResponse SendResponse
	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		// Non-fatal: response was successful but couldn't parse
		return nil
	}

	return nil
}
