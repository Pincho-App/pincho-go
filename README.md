# WirePusher Go SDK

[![Go Reference](https://pkg.go.dev/badge/gitlab.com/wirepusher/go-sdk.svg)](https://pkg.go.dev/gitlab.com/wirepusher/go-sdk)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/wirepusher/go-sdk)](https://goreportcard.com/report/gitlab.com/wirepusher/go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Official Go SDK for [WirePusher](https://wirepusher.com) push notifications.

Send push notifications to your mobile devices with just a few lines of Go code. Zero dependencies, uses only the Go standard library.

## Features

- **Zero dependencies** - Uses only the Go standard library (`net/http`)
- **Context support** - First-class context.Context for cancellation and timeouts
- **Functional options** - Flexible client configuration
- **Type-safe** - Full Go type safety with comprehensive error types
- **Customizable** - Bring your own `http.Client` for advanced use cases
- **Well-tested** - >95% test coverage with race detector
- **Production-ready** - Used in production environments

## Installation

```bash
go get gitlab.com/wirepusher/go-sdk
```

**Requirements:**
- Go 1.18 or higher

## Quick Start

```go
package main

import (
    "context"
    "log"

    "gitlab.com/wirepusher/go-sdk"
)

func main() {
    client := wirepusher.NewClient("your-token", "your-user-id")

    err := client.SendSimple(context.Background(), "Hello", "World")
    if err != nil {
        log.Fatal(err)
    }
}
```

## Usage

### Basic Send

Send a simple notification with just a title and message:

```go
client := wirepusher.NewClient("your-token", "your-user-id")

err := client.SendSimple(context.Background(), "Server Alert", "CPU usage high")
if err != nil {
    log.Fatal(err)
}
```

### Advanced Send with Options

Send a notification with additional options (type, tags, image, action URL):

```go
client := wirepusher.NewClient("your-token", "your-user-id")

err := client.Send(context.Background(), &wirepusher.SendOptions{
    Title:     "Deployment Complete",
    Message:   "Version 2.1.0 deployed to production",
    Type:      "deployment",
    Tags:      []string{"production", "backend"},
    ImageURL:  "https://example.com/success.png",
    ActionURL: "https://dashboard.example.com/deployments/123",
})
if err != nil {
    log.Fatal(err)
}
```

### Custom Configuration

#### Custom Timeout

```go
client := wirepusher.NewClient(
    "your-token",
    "your-user-id",
    wirepusher.WithTimeout(10*time.Second),
)
```

#### Custom HTTP Client

Useful for proxies, custom TLS configuration, or other advanced scenarios:

```go
httpClient := &http.Client{
    Timeout: 15 * time.Second,
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
    },
}

client := wirepusher.NewClient(
    "your-token",
    "your-user-id",
    wirepusher.WithHTTPClient(httpClient),
)
```

#### Custom API URL

For testing or self-hosted environments:

```go
client := wirepusher.NewClient(
    "your-token",
    "your-user-id",
    wirepusher.WithAPIURL("https://custom.example.com/api"),
)
```

### Context Usage

#### With Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := client.SendSimple(ctx, "Title", "Message")
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Request timed out")
    } else {
        log.Println("Request failed:", err)
    }
}
```

#### With Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel in another goroutine
go func() {
    time.Sleep(1 * time.Second)
    cancel()
}()

err := client.SendSimple(ctx, "Title", "Message")
if err != nil {
    if ctx.Err() == context.Canceled {
        log.Println("Request was canceled")
    }
}
```

### Error Handling

The SDK provides typed errors for different scenarios:

```go
err := client.Send(ctx, options)

switch e := err.(type) {
case *wirepusher.ValidationError:
    // 400 - Invalid request (missing required fields, etc.)
    log.Printf("Validation error: %s (status: %d)", e.Message, e.StatusCode)

case *wirepusher.AuthError:
    // 401/403 - Authentication failed (invalid token/user ID)
    log.Printf("Auth error: %s (status: %d)", e.Message, e.StatusCode)

case *wirepusher.RateLimitError:
    // 429 - Rate limit exceeded
    log.Printf("Rate limit error: %s (status: %d)", e.Message, e.StatusCode)

case *wirepusher.Error:
    // Other errors (5xx server errors, network errors, etc.)
    log.Printf("Error: %s", e.Message)

default:
    // Shouldn't happen, but handle anyway
    log.Printf("Unknown error: %v", err)
}
```

## API Reference

### Client

```go
type Client struct {
    Token      string        // WirePusher API token (required)
    UserID     string        // WirePusher user ID (required)
    APIURL     string        // API endpoint (defaults to production)
    HTTPClient *http.Client  // HTTP client (can be customized)
}
```

### NewClient

```go
func NewClient(token, userID string, opts ...ClientOption) *Client
```

Creates a new WirePusher client with the given token and user ID.

**Parameters:**
- `token` (string) - Your WirePusher API token
- `userID` (string) - Your WirePusher user ID
- `opts` (...ClientOption) - Optional configuration options

**Returns:** `*Client`

### SendSimple

```go
func (c *Client) SendSimple(ctx context.Context, title, message string) error
```

Sends a simple notification with just a title and message.

**Parameters:**
- `ctx` (context.Context) - Context for cancellation and timeouts
- `title` (string) - Notification title (max 256 characters)
- `message` (string) - Notification message (max 4096 characters)

**Returns:** `error`

### Send

```go
func (c *Client) Send(ctx context.Context, options *SendOptions) error
```

Sends a notification with full options.

**Parameters:**
- `ctx` (context.Context) - Context for cancellation and timeouts
- `options` (*SendOptions) - Notification options

**Returns:** `error`

### SendOptions

```go
type SendOptions struct {
    Title     string   // Required: Notification title (max 256 chars)
    Message   string   // Required: Notification message (max 4096 chars)
    Type      string   // Optional: Notification type (e.g., "alert", "info")
    Tags      []string // Optional: Tags for categorization (max 10)
    ImageURL  string   // Optional: URL to an image
    ActionURL string   // Optional: URL to open on tap
}
```

### Configuration Options

#### WithTimeout

```go
func WithTimeout(timeout time.Duration) ClientOption
```

Sets a custom HTTP timeout.

#### WithHTTPClient

```go
func WithHTTPClient(client *http.Client) ClientOption
```

Sets a custom HTTP client.

#### WithAPIURL

```go
func WithAPIURL(url string) ClientOption
```

Sets a custom API URL.

### Error Types

#### Error

```go
type Error struct {
    Message    string
    StatusCode int
}
```

General error type for API errors.

#### AuthError

```go
type AuthError struct {
    Message    string
    StatusCode int
}
```

Authentication error (401/403).

#### ValidationError

```go
type ValidationError struct {
    Message    string
    StatusCode int
}
```

Validation error (400).

#### RateLimitError

```go
type RateLimitError struct {
    Message    string
    StatusCode int
}
```

Rate limit error (429).

## Examples

See the [examples/](examples/) directory for complete, runnable examples:

- [Basic usage](examples/basic/main.go)
- [Advanced options](examples/advanced/main.go)
- [Context usage](examples/context/main.go)
- [Error handling](examples/errors/main.go)

## Testing

Run tests:

```bash
go test
```

Run tests with coverage:

```bash
go test -cover
```

Run tests with race detector:

```bash
go test -race
```

Generate coverage report:

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Development

### Project Structure

```
.
├── client.go           # Main client implementation
├── types.go            # Request/response types
├── errors.go           # Custom error types
├── client_test.go      # Comprehensive tests
├── examples/           # Usage examples
│   ├── basic/
│   ├── advanced/
│   ├── context/
│   └── errors/
├── go.mod              # Go module definition
└── README.md           # This file
```

### Building

```bash
go build
```

### Linting

```bash
go vet ./...
golangci-lint run  # If golangci-lint is installed
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Security

For security vulnerabilities, please email security@wirepusher.com. See [SECURITY.md](SECURITY.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Documentation:** [pkg.go.dev/gitlab.com/wirepusher/go-sdk](https://pkg.go.dev/gitlab.com/wirepusher/go-sdk)
- **Issues:** [GitLab Issues](https://gitlab.com/wirepusher/go-sdk/-/issues)
- **Email:** support@wirepusher.com
- **Website:** [wirepusher.com](https://wirepusher.com)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and changes.

---

Made with ❤️ by the WirePusher team
