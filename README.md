# WirePusher Go Library

Official Go client for [WirePusher](https://wirepusher.dev) push notifications.

## Installation

```bash
go get gitlab.com/wirepusher/wirepusher-go
```

## Quick Start

```go
import "gitlab.com/wirepusher/wirepusher-go"

// Auto-load token from WIREPUSHER_TOKEN env var
client := wirepusher.NewClient("")
err := client.SendSimple(ctx, "Deploy Complete", "Version 1.2.3 deployed")

// Or provide token explicitly
client := wirepusher.NewClient("YOUR_TOKEN")
err := client.SendSimple(ctx, "Alert", "Server CPU at 95%")
```

## Features

```go
// Full parameters
err := client.Send(ctx, &wirepusher.SendOptions{
    Title:     "Deploy Complete",
    Message:   "Version 1.2.3 deployed",
    Type:      "deployment",
    Tags:      []string{"production", "backend"},
    ImageURL:  "https://example.com/success.png",
    ActionURL: "https://example.com/deploy/123",
})

// AI-powered notifications (NotifAI)
response, err := client.NotifAI(ctx, &wirepusher.NotifAIOptions{
    Text: "deployment finished, v2.1.3 is live",
})
// response.Notification contains AI-generated title, message, tags

// Encrypted messages
err := client.Send(ctx, &wirepusher.SendOptions{
    Title:              "Security Alert",
    Message:            "Sensitive data",
    Type:               "security",
    EncryptionPassword: "your_password",
})
```

## Configuration

```go
// Environment variables (recommended)
// WIREPUSHER_TOKEN - API token (required if not passed to constructor)
// WIREPUSHER_TIMEOUT - Request timeout in seconds (default: 30)
// WIREPUSHER_MAX_RETRIES - Retry attempts (default: 3)

// Or explicit configuration
client := wirepusher.NewClient(
    "abc12345",
    wirepusher.WithTimeout(60*time.Second),
    wirepusher.WithMaxRetries(5),
)
```

## Error Handling

```go
err := client.Send(ctx, options)
if err != nil {
    switch e := err.(type) {
    case *wirepusher.AuthError:
        log.Printf("Invalid token: %s", e.Message)
    case *wirepusher.ValidationError:
        log.Printf("Invalid parameters: %s", e.Message)
    case *wirepusher.RateLimitError:
        log.Printf("Rate limited: %s", e.Message)
    case *wirepusher.ServerError:
        log.Printf("Server error: %s", e.Message)
    case *wirepusher.NetworkError:
        log.Printf("Network error: %v", e.Unwrap())
    }
}
```

Automatic retry with exponential backoff for network errors, 5xx, and 429 (rate limit).

## Smart Rate Limiting

The client automatically respects `Retry-After` headers and tracks rate limit information:

```go
err := client.Send(ctx, options)
// Check rate limit info after any request
if info := client.LastRateLimit; info != nil {
    fmt.Printf("Remaining: %d/%d, Resets: %s\n", info.Remaining, info.Limit, info.Reset)
}
```

## Requirements

- Go 1.18+
- Zero runtime dependencies (stdlib only)
- Context support for cancellation and timeouts

## Links

- **Get Token**: App → Settings → Help → copy token
- **Documentation**: https://wirepusher.dev/help
- **Repository**: https://gitlab.com/wirepusher/wirepusher-go
- **pkg.go.dev**: https://pkg.go.dev/gitlab.com/wirepusher/wirepusher-go

## License

MIT
