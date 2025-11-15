# WirePusher Go Client Library

[![Go Reference](https://pkg.go.dev/badge/gitlab.com/wirepusher/wirepusher-go.svg)](https://pkg.go.dev/gitlab.com/wirepusher/wirepusher-go)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/wirepusher/wirepusher-go)](https://goreportcard.com/report/gitlab.com/wirepusher/wirepusher-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Official Go Client Library for [WirePusher](https://wirepusher.dev) push notifications.

## Features

- ✅ **Zero Dependencies** - Uses only Go standard library
- ✅ **Context Support** - First-class context.Context for cancellation and timeouts
- ✅ **Type-Safe** - Full Go type safety with comprehensive error types
- ✅ **Functional Options** - Flexible client configuration
- ✅ **Production-Ready** - >95% test coverage with race detector
- ✅ **Automatic Retries** - Exponential backoff with configurable retry logic
- ✅ **Debug Logging** - Optional logging interface for debugging and monitoring
- ✅ **Tag Validation** - Automatic tag normalization and validation

## Installation

```bash
go get gitlab.com/wirepusher/wirepusher-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token)

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

### Basic Example

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token)

    err := client.SendSimple(context.Background(),
        "Deploy Complete",
        "Version 1.2.3 deployed to production",
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### All Parameters

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token)

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

### Custom Configuration

```go
package main

import (
    "context"
    "crypto/tls"
    "net/http"
    "os"
    "time"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")

    // Custom timeout
    client := wirepusher.NewClient(token,
        wirepusher.WithTimeout(10*time.Second),
    )

    // Custom HTTP client (for proxies, TLS config, etc.)
    httpClient := &http.Client{
        Timeout: 15 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
        },
    }
    client = wirepusher.NewClient(token,
        wirepusher.WithHTTPClient(httpClient),
    )

    // Custom API URL (for testing)
    client = wirepusher.NewClient(token,
        wirepusher.WithAPIURL("https://custom.example.com/api"),
    )

    // Enable debug logging
    logger := wirepusher.NewStdLogger("wirepusher")
    client = wirepusher.NewClient(token,
        wirepusher.WithLogger(logger),
    )

    // Configure retry behavior
    client = wirepusher.NewClient(token,
        wirepusher.WithMaxRetries(5), // Default: 3
    )
}
```

### Context Usage

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token)

    // With timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := client.SendSimple(ctx, "Deploy Complete", "Version 1.2.3 deployed")
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            log.Println("Request timed out")
        } else {
            log.Println("Request failed:", err)
        }
    }

    // With cancellation
    ctx2, cancel2 := context.WithCancel(context.Background())
    go func() {
        time.Sleep(1 * time.Second)
        cancel2()
    }()

    err = client.SendSimple(ctx2, "Deploy Complete", "Version 1.2.3 deployed")
    if err != nil {
        if ctx2.Err() == context.Canceled {
            log.Println("Request was canceled")
        }
    }
}
```

## Encryption

Encrypt notification messages using AES-128-CBC. Only the `message` field is encrypted—`title`, `type`, `tags`, `image_url`, and `action_url` remain unencrypted for filtering and display.

**Setup:**
1. In the app, create a notification type
2. Set an encryption password for that type
3. Pass the same `type` and password when sending

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    password := os.Getenv("WIREPUSHER_ENCRYPTION_PASSWORD")
    client := wirepusher.NewClient(token)

    err := client.Send(context.Background(), &wirepusher.SendOptions{
        Title:              "Security Alert",
        Message:            "Unauthorized access attempt detected",
        Type:               "security",
        EncryptionPassword: password,
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

**Security notes:**
- Use strong passwords (minimum 12 characters)
- Store passwords securely (environment variables, secret managers)
- Password must match the type configuration in the app

## Logging

Enable logging to debug issues or monitor API interactions. By default, no logging is performed for zero overhead.

### Basic Logging

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")

    // Enable standard output logging
    logger := wirepusher.NewStdLogger("wirepusher")
    client := wirepusher.NewClient(token,
        wirepusher.WithLogger(logger),
    )

    err := client.SendSimple(context.Background(),
        "Test",
        "This will log debug information",
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### Custom Logger

Implement the `Logger` interface to integrate with your existing logging solution:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

// CustomLogger integrates with your logging system
type CustomLogger struct {
    // Your logger instance here
}

func (l *CustomLogger) Printf(format string, v ...interface{}) {
    // Your logging implementation
    fmt.Printf(format, v...)
}

func (l *CustomLogger) Println(v ...interface{}) {
    // Your logging implementation
    fmt.Println(v...)
}

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")

    customLogger := &CustomLogger{}
    client := wirepusher.NewClient(token,
        wirepusher.WithLogger(customLogger),
    )

    client.SendSimple(context.Background(), "Test", "Message")
}
```

### What Gets Logged

When logging is enabled, the client logs:
- **DEBUG**: Send/NotifAI calls, tag normalization, encryption operations, retry attempts
- **INFO**: Successful operations
- **WARNING**: Rate limits, max retries exceeded
- **ERROR**: Encryption failures, request errors

## Tag Normalization

Tags are automatically normalized for consistency:
- Converted to lowercase
- Whitespace trimmed
- Only alphanumeric characters, hyphens, and underscores allowed
- Duplicates removed (case-insensitive)
- Empty tags filtered out

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token)

    err := client.Send(context.Background(), &wirepusher.SendOptions{
        Title:   "Deploy Complete",
        Message: "Version 1.2.3 deployed",
        // These tags will be normalized:
        // "Production" -> "production"
        // "  Release  " -> "release"
        // "production" (duplicate) -> removed
        // "invalid tag" (has space) -> removed
        Tags: []string{"Production", "  Release  ", "production", "invalid tag"},
    })
    // Final tags sent: ["production", "release"]

    if err != nil {
        log.Fatal(err)
    }
}
```

## Automatic Retries

The client automatically retries failed requests with exponential backoff:
- **Network errors**: Retried with 1s, 2s, 4s backoff
- **Server errors (5xx)**: Retried with exponential backoff
- **Rate limits (429)**: Retried with longer backoff (5s, 10s, 20s)
- **Default**: 3 retry attempts (configurable)

Non-retryable errors (400, 401, 403) fail immediately.

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")

    // Customize retry behavior
    client := wirepusher.NewClient(token,
        wirepusher.WithMaxRetries(5), // Default: 3
    )

    // Disable retries
    client = wirepusher.NewClient(token,
        wirepusher.WithMaxRetries(0),
    )

    err := client.SendSimple(context.Background(), "Test", "Message")
    if err != nil {
        log.Fatal(err)
    }
}
```

## API Reference

### Client

```go
type Client struct {
    Token      string        // WirePusher token (required)
    APIURL     string        // API endpoint (defaults to production)
    HTTPClient *http.Client  // HTTP client (can be customized)
    MaxRetries int          // Maximum retry attempts (default: 3)
    Logger     Logger       // Logger for debug/info messages (default: NoOpLogger)
}
```

### NewClient

```go
func NewClient(token string, opts ...ClientOption) *Client
```

Creates a new WirePusher client.

**Parameters:**
- `token` (string): Your WirePusher token (8-character alphanumeric)
- `opts` (...ClientOption): Optional configuration options

**Returns:** `*Client`

**Panics:** If token is empty

### SendSimple

```go
func (c *Client) SendSimple(ctx context.Context, title, message string) error
```

Sends a simple notification with just a title and message.

**Parameters:**
- `ctx` (context.Context): Context for cancellation and timeouts
- `title` (string): Notification title (max 256 characters)
- `message` (string): Notification message (max 4096 characters)

**Returns:** `error`

### Send

```go
func (c *Client) Send(ctx context.Context, options *SendOptions) error
```

Sends a notification with full options.

**Parameters:**
- `ctx` (context.Context): Context for cancellation and timeouts
- `options` (*SendOptions): Notification options

**Returns:** `error`

### SendOptions

```go
type SendOptions struct {
    Title              string   // Required: Notification title (max 256 chars)
    Message            string   // Required: Notification message (max 4096 chars)
    Type               string   // Optional: Notification type (e.g., "deployment", "alert")
    Tags               []string // Optional: Tags for categorization (max 10)
    ImageURL           string   // Optional: URL to an image
    ActionURL          string   // Optional: URL to open on tap
    EncryptionPassword string   // Optional: Password for AES-128-CBC encryption
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

#### WithMaxRetries

```go
func WithMaxRetries(maxRetries int) ClientOption
```

Sets the maximum number of retry attempts. Set to 0 to disable retries.

#### WithLogger

```go
func WithLogger(logger Logger) ClientOption
```

Sets a custom logger for debug/info messages.

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

## Error Handling

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token)

    err := client.SendSimple(context.Background(),
        "Deploy Complete",
        "Version 1.2.3 deployed to production",
    )

    switch e := err.(type) {
    case *wirepusher.ValidationError:
        // 400 - Invalid request
        log.Printf("Validation error: %s (status: %d)", e.Message, e.StatusCode)

    case *wirepusher.AuthError:
        // 401/403 - Authentication failed
        log.Printf("Auth error: %s (status: %d)", e.Message, e.StatusCode)

    case *wirepusher.RateLimitError:
        // 429 - Rate limit exceeded
        log.Printf("Rate limit error: %s (status: %d)", e.Message, e.StatusCode)

    case *wirepusher.Error:
        // Other errors (5xx server errors, network errors, etc.)
        log.Printf("Error: %s", e.Message)

    case nil:
        log.Println("Notification sent successfully")

    default:
        log.Printf("Unknown error: %v", err)
    }
}
```

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
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token)

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
        token := os.Getenv("WIREPUSHER_TOKEN")
        client := wirepusher.NewClient(token)

        err := client.Send(context.Background(), &wirepusher.SendOptions{
            Title:   "Server Alert",
            Message: fmt.Sprintf("CPU: %.1f%%, Memory: %.1f%%", cpu, memory),
            Type:    "alert",
            Tags:    []string{"server", "critical"},
        })
        if err != nil {
            log.Println("Failed to send alert:", err)
        }
    }
}

func main() {
    // Your monitoring logic here
    checkServerHealth(85.0, 75.0)
}
```

### HTTP Handler

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func deployHandler(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        Version string `json:"version"`
    }
    json.NewDecoder(r.Body).Decode(&payload)

    // Your deployment logic here

    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token)

    err := client.Send(context.Background(), &wirepusher.SendOptions{
        Title:   "Deploy Complete",
        Message: "Version " + payload.Version + " deployed to production",
        Type:    "deployment",
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func main() {
    http.HandleFunc("/deploy", deployHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Batch Processing

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "gitlab.com/wirepusher/wirepusher-go"
)

func processBatch(records int) {
    // Your batch processing logic here

    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token)

    err := client.Send(context.Background(), &wirepusher.SendOptions{
        Title:   "Batch Job Complete",
        Message: fmt.Sprintf("Processed %d records", records),
        Type:    "batch",
        Tags:    []string{"data-pipeline", "success"},
    })
    if err != nil {
        log.Println("Failed to send notification:", err)
    }
}

func main() {
    processBatch(10000)
}
```

## Development

### Setup

```bash
# Clone repository
git clone https://gitlab.com/wirepusher/wirepusher-go.git
cd go-sdk

# Install dependencies
go mod download

# Run tests
go test -v
```

### Testing

```bash
# Run tests
go test

# Run tests with coverage
go test -cover

# Run tests with race detector
go test -race

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Linting

```bash
go vet ./...
golangci-lint run  # If golangci-lint is installed
```

## Requirements

- Go 1.18 or higher

## Links

- **Documentation**: https://pkg.go.dev/gitlab.com/wirepusher/wirepusher-go
- **Repository**: https://gitlab.com/wirepusher/wirepusher-go
- **Issues**: https://gitlab.com/wirepusher/wirepusher-go/-/issues
- **Website**: https://wirepusher.dev

## License

MIT License - see [LICENSE](LICENSE) file for details.
