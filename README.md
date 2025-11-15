# WirePusher Go Library

[![Go Reference](https://pkg.go.dev/badge/gitlab.com/wirepusher/wirepusher-go.svg)](https://pkg.go.dev/gitlab.com/wirepusher/wirepusher-go)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/wirepusher/wirepusher-go)](https://goreportcard.com/report/gitlab.com/wirepusher/wirepusher-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Official Go client library for [WirePusher](https://wirepusher.dev) push notifications.

## Features

- ✅ **Zero Dependencies** - Uses only Go standard library
- ✅ **Type-Safe** - Full Go type safety with comprehensive error types
- ✅ **AI-Powered** - NotifAI endpoint for generating notifications from text
- ✅ **Context Support** - First-class context.Context for cancellation and timeouts
- ✅ **Automatic Retries** - Exponential backoff with smart error handling
- ✅ **Production-Ready** - >95% test coverage with race detector

## Quick Start

```bash
go get gitlab.com/wirepusher/wirepusher-go
```

```go
package main

import (
	"context"
	"log"

	"gitlab.com/wirepusher/wirepusher-go"
)

func main() {
	client := wirepusher.NewClient("YOUR_TOKEN")

	err := client.SendSimple(context.Background(),
		"Deploy Complete",
		"Version 1.2.3 deployed to production",
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

**Get your token:** Open app → Settings → Help → copy token

## Usage

### All Parameters

```go
package main

import (
	"context"
	"log"

	"gitlab.com/wirepusher/wirepusher-go"
)

func main() {
	client := wirepusher.NewClient("abc12345")

	err := client.Send(context.Background(), &wirepusher.SendOptions{
		Title:     "Deploy Complete",
		Message:   "Version 1.2.3 deployed to production",
		Type:      "deployment",
		Tags:      []string{"production", "backend"},
		ImageURL:  "https://cdn.example.com/success.png",
		ActionURL: "https://dash.example.com/deploy/123",
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

### NotifAI - AI-Powered Notifications

Let AI generate structured notifications from free-form text using Gemini:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"gitlab.com/wirepusher/wirepusher-go"
)

func main() {
	client := wirepusher.NewClient("abc12345")

	response, err := client.NotifAI(context.Background(), &wirepusher.NotifAIOptions{
		Text: "deployment finished successfully, v2.1.3 is live on prod",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Title: %s\n", response.Notification.Title)
	fmt.Printf("Message: %s\n", response.Notification.Message)
	fmt.Printf("Tags: %v\n", response.Notification.Tags)
}
```

The AI automatically generates:
- Title and message
- Relevant tags
- Action URL (when applicable)

Override the AI-generated type:

```go
response, err := client.NotifAI(context.Background(), &wirepusher.NotifAIOptions{
	Text: "cpu at 95% on web-3",
	Type: "alert", // Override AI type
})
```

### Automatic Retries

The library automatically retries failed requests with exponential backoff (default: 3 retries). Retries network errors, 5xx errors, and 429 (rate limit). Client errors (400, 401, 403) are not retried.

```go
// Configure retries
client := wirepusher.NewClient("abc12345",
	wirepusher.WithMaxRetries(5), // Custom
)

// Disable retries
client = wirepusher.NewClient("abc12345",
	wirepusher.WithMaxRetries(0), // Disable
)
```

### Debug Logging

Enable debug logging using the built-in Logger interface:

```go
package main

import (
	"context"
	"log"

	"gitlab.com/wirepusher/wirepusher-go"
)

func main() {
	// Enable standard output logging
	logger := wirepusher.NewStdLogger("wirepusher")
	client := wirepusher.NewClient("abc12345",
		wirepusher.WithLogger(logger),
	)

	err := client.SendSimple(context.Background(), "Test", "Message")
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// wirepusher: 2024/01/15 10:30:00 DEBUG: Send() called with title: Test
	// wirepusher: 2024/01/15 10:30:00 INFO: Notification sent successfully
}
```

## Encryption

Encrypt notification messages using AES-128-CBC. Only the `message` is encrypted—`title`, `type`, and `tags` remain visible for filtering.

```go
package main

import (
	"context"
	"log"
	"os"

	"gitlab.com/wirepusher/wirepusher-go"
)

func main() {
	// 1. In app: create notification type with encryption password
	// 2. Send with matching type and password
	client := wirepusher.NewClient("abc12345")

	err := client.Send(context.Background(), &wirepusher.SendOptions{
		Title:              "Security Alert",
		Message:            "Sensitive data here",
		Type:               "security",
		EncryptionPassword: os.Getenv("ENCRYPTION_PASSWORD"),
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

## API Reference

### Client

**Constructor:**

```go
func NewClient(token string, opts ...ClientOption) *Client
```

**Parameters:**
- `token` (string, required): Your WirePusher token (8-character alphanumeric string)
- `opts` (...ClientOption, optional): Configuration options
  - `WithTimeout(duration)` - Request timeout (default: 30s)
  - `WithMaxRetries(n)` - Maximum retry attempts (default: 3, set to 0 to disable)
  - `WithHTTPClient(client)` - Custom HTTP client
  - `WithAPIURL(url)` - Custom base URL for testing
  - `WithLogger(logger)` - Enable logging

### send()

Send a notification.

```go
func (c *Client) Send(ctx context.Context, options *SendOptions) error
```

**Parameters:**
- `ctx` (context.Context, required): Context for cancellation and timeouts
- `options` (*SendOptions, required): Notification options
  - `Title` (string, required): Notification title
  - `Message` (string, optional): Notification message
  - `Type` (string, optional): Category for organization
  - `Tags` ([]string, optional): Tags for filtering (automatically normalized)
  - `ImageURL` (string, optional): Image URL to display
  - `ActionURL` (string, optional): URL to open when tapped
  - `EncryptionPassword` (string, optional): Password for encryption

**Returns:** `error`

**Raises:**
- `*AuthError`: Invalid token (401, 403)
- `*ValidationError`: Invalid parameters (400)
- `*RateLimitError`: Rate limit exceeded (429)
- `*ServerError`: Server error (5xx)
- `*NetworkError`: Network/timeout error
- `*Error`: Other API errors

### notifai()

Generate AI-powered notification from free-form text.

```go
func (c *Client) NotifAI(ctx context.Context, options *NotifAIOptions) (*NotifAIResponse, error)
```

**Parameters:**
- `ctx` (context.Context, required): Context for cancellation and timeouts
- `options` (*NotifAIOptions, required): NotifAI options
  - `Text` (string, required): Free-form text to convert
  - `Type` (string, optional): Override AI-generated type

**Returns:** `(*NotifAIResponse, error)`

**NotifAIResponse fields:**
- `Status` (string): Response status
- `Message` (string): Response message
- `Notification` (Notification): AI-generated notification
  - `Title` (string): Generated title
  - `Message` (string): Generated message
  - `Type` (string): Generated or overridden type
  - `Tags` ([]string): Generated tags
  - `ActionURL` (string): Generated action URL (if applicable)

**Raises:**
- Same exceptions as `Send()`

## Error Handling

```go
package main

import (
	"context"
	"log"

	"gitlab.com/wirepusher/wirepusher-go"
)

func main() {
	client := wirepusher.NewClient("abc12345")

	err := client.SendSimple(context.Background(), "Title", "Message")

	switch e := err.(type) {
	case *wirepusher.AuthError:
		log.Printf("Invalid token: %s", e.Message)
	case *wirepusher.ValidationError:
		log.Printf("Invalid parameters: %s", e.Message)
	case *wirepusher.RateLimitError:
		log.Printf("Rate limit exceeded: %s", e.Message)
	case *wirepusher.ServerError:
		log.Printf("Server error: %s", e.Message)
	case *wirepusher.NetworkError:
		log.Printf("Network error: %s", e.Message)
	case nil:
		log.Println("Success!")
	default:
		log.Printf("Error: %v", err) // Auto-retry handles transient errors
	}
}
```

**Error Types:** `AuthError`, `ValidationError`, `RateLimitError`, `ServerError`, `NetworkError`, `Error`

## Validation Philosophy

This library performs **minimal client-side validation** to ensure the API remains the source of truth:

### ✅ We Validate

- **Required parameters**: `title` and `token`
- **Parameter types**: Ensuring correct Go types

### ✅ We Normalize

- **Tags**: Lowercase conversion, whitespace trimming, deduplication, and invalid character filtering
- **Logging**: Debug logs when normalization occurs

### ❌ We Don't Validate

- **Message**: Optional parameter (not required by API)
- **Tag limits**: API validates max 10 tags, 50 characters each
- **Business rules**: Rules that may change server-side

### Why This Approach?

**The API is the source of truth.** Client-side validation of business rules can create false negatives when API rules evolve independently of client library updates. By performing minimal validation:

- ✅ Valid requests are never rejected due to outdated client logic
- ✅ API error messages provide detailed context (error codes, param names)
- ✅ Less maintenance burden across client libraries
- ✅ Consistent behavior as API evolves

The API returns comprehensive error responses with `type`, `code`, `message`, and `param` fields to help you debug validation failures.

## Examples

### CI/CD Pipeline

```go
package main

import (
	"context"
	"log"
	"os"

	"gitlab.com/wirepusher/wirepusher-go"
)

func notifyDeployment(version, environment string) {
	client := wirepusher.NewClient(os.Getenv("WIREPUSHER_TOKEN"))

	err := client.Send(context.Background(), &wirepusher.SendOptions{
		Title:   "Deploy Complete",
		Message: "Version " + version + " deployed to " + environment,
		Type:    "deployment",
		Tags:    []string{environment, version},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	notifyDeployment("1.2.3", "production")
}
```

### Server Monitoring

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gitlab.com/wirepusher/wirepusher-go"
)

func checkServerHealth(cpu, memory float64) {
	if cpu > 80 || memory > 80 {
		client := wirepusher.NewClient(os.Getenv("WIREPUSHER_TOKEN"))

		err := client.Send(context.Background(), &wirepusher.SendOptions{
			Title:   "Server Alert",
			Message: fmt.Sprintf("CPU: %.1f%%, Memory: %.1f%%", cpu, memory),
			Type:    "alert",
			Tags:    []string{"critical"},
		})
		if err != nil {
			log.Println("Failed to send alert:", err)
		}
	}
}

func main() {
	checkServerHealth(85.0, 75.0)
}
```

## Development

```bash
# Setup
git clone https://gitlab.com/wirepusher/wirepusher-go.git
cd wirepusher-go

# Test
go test -v
go test -race        # With race detector
go test -cover       # With coverage

# Lint
go fmt ./...
go vet ./...
```

## Requirements

- Go 1.18+

## Links

- **Documentation**: https://wirepusher.dev/help
- **Repository**: https://gitlab.com/wirepusher/wirepusher-go
- **Issues**: https://gitlab.com/wirepusher/wirepusher-go/-/issues
- **pkg.go.dev**: https://pkg.go.dev/gitlab.com/wirepusher/wirepusher-go

## License

MIT License - see [LICENSE](LICENSE) file for details.
