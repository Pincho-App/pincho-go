# Pincho Go Library

Official Go client for [Pincho](https://pincho.app) push notifications.

## Installation

```bash
go get github.com/Pincho-App/pincho-go
```

## Quick Start

```go
import "github.com/Pincho-App/pincho-go"

// Auto-load token from PINCHO_TOKEN env var
client := pincho.NewClient("")
err := client.SendSimple(ctx, "Deploy Complete", "Version 1.2.3 deployed")

// Or provide token explicitly
client := pincho.NewClient("YOUR_TOKEN")
err := client.SendSimple(ctx, "Alert", "Server CPU at 95%")
```

## Features

```go
// Full parameters
err := client.Send(ctx, &pincho.SendOptions{
    Title:     "Deploy Complete",
    Message:   "Version 1.2.3 deployed",
    Type:      "deployment",
    Tags:      []string{"production", "backend"},
    ImageURL:  "https://example.com/success.png",
    ActionURL: "https://example.com/deploy/123",
})

// AI-powered notifications (NotifAI)
response, err := client.NotifAI(ctx, &pincho.NotifAIOptions{
    Text: "deployment finished, v2.1.3 is live",
})
// response.Notification contains AI-generated title, message, tags

// Encrypted notifications (title, message, URLs encrypted; type, tags unencrypted)
err := client.Send(ctx, &pincho.SendOptions{
    Title:              "Security Alert",
    Message:            "Sensitive data",
    Type:               "security",
    EncryptionPassword: "your_password",
})
```

## Configuration

```go
// Environment variables (recommended)
// PINCHO_TOKEN - API token (required if not passed to constructor)
// PINCHO_TIMEOUT - Request timeout in seconds (default: 30)
// PINCHO_MAX_RETRIES - Retry attempts (default: 3)

// Or explicit configuration
client := pincho.NewClient(
    "abc12345",
    pincho.WithTimeout(60*time.Second),
    pincho.WithMaxRetries(5),
)
```

## Error Handling

Use sentinel errors with `errors.Is()` or type assertions with `errors.As()`:

```go
err := client.Send(ctx, options)
if err != nil {
    // Check error type with errors.Is()
    if errors.Is(err, pincho.ErrAuth) {
        log.Printf("Authentication failed")
    } else if errors.Is(err, pincho.ErrRateLimit) {
        log.Printf("Rate limited")
    }

    // Or extract error details with errors.As()
    var rateLimitErr *pincho.RateLimitError
    if errors.As(err, &rateLimitErr) {
        log.Printf("Retry after %d seconds", rateLimitErr.RetryAfter)
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
- **Documentation**: https://pincho.app/help
- **Repository**: https://github.com/Pincho-App/pincho-go
- **pkg.go.dev**: https://pkg.go.dev/github.com/Pincho-App/pincho-go

## License

MIT
