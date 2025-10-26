package wirepusher

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
type SendOptions struct {
	Title     string   `json:"title"`
	Message   string   `json:"message"`
	Type      string   `json:"type,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	ImageURL  string   `json:"imageURL,omitempty"`
	ActionURL string   `json:"actionURL,omitempty"`
}

// SendResponse is the response from the WirePusher API for a send operation.
type SendResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
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
