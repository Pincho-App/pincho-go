package wirepusher

import "time"

// RateLimitInfo contains rate limit information from the API response.
type RateLimitInfo struct {
	// Limit is the maximum number of requests allowed in the window.
	Limit int
	// Remaining is the number of requests remaining in the current window.
	Remaining int
	// Reset is the time when the rate limit window resets.
	Reset time.Time
}

// SendOptions contains all parameters for sending a notification.
//
// Required fields:
//   - Title: The notification title (max 256 characters)
//   - Message: The notification message body (max 4096 characters)
//
// Optional fields:
//   - Type: Notification type for filtering/grouping (e.g., "billing", "alert")
//   - Tags: List of tags for categorization (max 10 tags, 128 chars each)
//   - ImageURL: URL to an image to display with the notification
//   - ActionURL: URL to open when user taps the notification
//   - EncryptionPassword: Password for AES-128-CBC encryption (must match type configuration in app)
type SendOptions struct {
	Title              string   `json:"title"`
	Message            string   `json:"message"`
	Type               string   `json:"type,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	ImageURL           string   `json:"imageURL,omitempty"`
	ActionURL          string   `json:"actionURL,omitempty"`
	EncryptionPassword string   `json:"-"` // Not sent to API, used locally for encryption
}

// SendResponse is the response from the WirePusher API for a send operation.
type SendResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ErrorResponse represents the API error response with nested structure.
type ErrorResponse struct {
	Status string       `json:"status"`
	Error  ErrorDetails `json:"error"`
}

// ErrorDetails contains the nested error information.
type ErrorDetails struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

// NotificationFilter contains parameters for filtering notifications.
//
// All fields are optional. If no filters are provided, all notifications are returned.
type NotificationFilter struct {
	Type  string   `json:"type,omitempty"`
	Tags  []string `json:"tags,omitempty"`
	Limit int      `json:"limit,omitempty"`
}

// Notification represents a stored notification from the WirePusher API.
type Notification struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Message   string   `json:"message"`
	Type      string   `json:"type,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	ImageURL  string   `json:"imageURL,omitempty"`
	ActionURL string   `json:"actionURL,omitempty"`
	Timestamp string   `json:"timestamp"`
}

// NotificationListResponse is the response from the API when listing notifications.
type NotificationListResponse struct {
	Status        string         `json:"status"`
	Notifications []Notification `json:"notifications"`
}

// DeleteResponse is the response from the API when deleting a notification.
type DeleteResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NotifAIOptions contains parameters for generating AI-powered notifications.
//
// The NotifAI endpoint uses AI to convert free-form text into a structured notification.
type NotifAIOptions struct {
	Text string `json:"text"`           // Free-form text to convert to notification (required)
	Type string `json:"type,omitempty"` // Optional type override
}

// NotifAINotification represents the AI-generated notification data.
type NotifAINotification struct {
	Title     string   `json:"title"`
	Message   string   `json:"message"`
	Type      string   `json:"type,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	ActionURL string   `json:"actionURL,omitempty"`
}

// NotifAIResponse is the response from the NotifAI endpoint.
type NotifAIResponse struct {
	Status       string              `json:"status"`
	Message      string              `json:"message"`
	Notification NotifAINotification `json:"notification"`
}
