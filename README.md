# WirePusher Go SDK

[![Go Reference](https://pkg.go.dev/badge/gitlab.com/wirepusher/go-sdk.svg)](https://pkg.go.dev/gitlab.com/wirepusher/go-sdk)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/wirepusher/go-sdk)](https://goreportcard.com/report/gitlab.com/wirepusher/go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Official Go SDK for [WirePusher](https://wirepusher.dev) push notifications.

## Features

- ✅ **Zero Dependencies** - Uses only Go standard library
- ✅ **Context Support** - First-class context.Context for cancellation and timeouts
- ✅ **Type-Safe** - Full Go type safety with comprehensive error types
- ✅ **Functional Options** - Flexible client configuration
- ✅ **Production-Ready** - >95% test coverage with race detector

## Installation

```bash
go get gitlab.com/wirepusher/go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/go-sdk"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token, "")

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

    "gitlab.com/wirepusher/go-sdk"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token, "")

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

    "gitlab.com/wirepusher/go-sdk"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token, "")

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

    "gitlab.com/wirepusher/go-sdk"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")

    // Custom timeout
    client := wirepusher.NewClient(token, "",
        wirepusher.WithTimeout(10*time.Second),
    )

    // Custom HTTP client (for proxies, TLS config, etc.)
    httpClient := &http.Client{
        Timeout: 15 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
        },
    }
    client = wirepusher.NewClient(token, "",
        wirepusher.WithHTTPClient(httpClient),
    )

    // Custom API URL (for testing)
    client = wirepusher.NewClient(token, "",
        wirepusher.WithAPIURL("https://custom.example.com/api"),
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

    "gitlab.com/wirepusher/go-sdk"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token, "")

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

    "gitlab.com/wirepusher/go-sdk"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    password := os.Getenv("WIREPUSHER_ENCRYPTION_PASSWORD")
    client := wirepusher.NewClient(token, "")

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

## API Reference

### Client

```go
type Client struct {
    Token      string        // WirePusher token (required)
    UserID     string        // Legacy user ID (not recommended)
    APIURL     string        // API endpoint (defaults to production)
    HTTPClient *http.Client  // HTTP client (can be customized)
}
```

### NewClient

```go
func NewClient(token, userID string, opts ...ClientOption) *Client
```

Creates a new WirePusher client. Use `token` for authentication (recommended). The `userID` parameter is legacy and not recommended for new integrations.

**Parameters:**
- `token` (string): Your WirePusher token (starts with `wpu_` or `wpt_`)
- `userID` (string): Legacy user ID (pass empty string for token-based auth)
- `opts` (...ClientOption): Optional configuration options

**Returns:** `*Client`

**Panics:** If both token and userID are empty

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

    "gitlab.com/wirepusher/go-sdk"
)

func main() {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token, "")

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

    "gitlab.com/wirepusher/go-sdk"
)

func notifyDeployment(version, environment string) {
    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token, "")

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

    "gitlab.com/wirepusher/go-sdk"
)

func checkServerHealth(cpu, memory float64) {
    if cpu > 80 || memory > 80 {
        token := os.Getenv("WIREPUSHER_TOKEN")
        client := wirepusher.NewClient(token, "")

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

    "gitlab.com/wirepusher/go-sdk"
)

func deployHandler(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        Version string `json:"version"`
    }
    json.NewDecoder(r.Body).Decode(&payload)

    // Your deployment logic here

    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token, "")

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

    "gitlab.com/wirepusher/go-sdk"
)

func processBatch(records int) {
    // Your batch processing logic here

    token := os.Getenv("WIREPUSHER_TOKEN")
    client := wirepusher.NewClient(token, "")

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
git clone https://gitlab.com/wirepusher/go-sdk.git
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

- **Documentation**: https://pkg.go.dev/gitlab.com/wirepusher/go-sdk
- **Repository**: https://gitlab.com/wirepusher/go-sdk
- **Issues**: https://gitlab.com/wirepusher/go-sdk/-/issues
- **Website**: https://wirepusher.dev

## License

MIT License - see [LICENSE](LICENSE) file for details.
