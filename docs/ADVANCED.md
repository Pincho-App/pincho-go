# Advanced Usage Guide

This guide covers advanced features of the WirePusher Go client library.

## Rate Limit Monitoring

The client automatically tracks rate limit information from API responses. After each successful request, you can inspect `client.LastRateLimit` to see your current quota:

```go
client := wirepusher.NewClient("your-token")
err := client.Send(ctx, &wirepusher.SendOptions{
    Title:   "Alert",
    Message: "Server CPU high",
})

if info := client.LastRateLimit; info != nil {
    fmt.Printf("Rate Limit: %d/%d requests remaining\n", info.Remaining, info.Limit)
    fmt.Printf("Resets at: %s\n", info.Reset.Format(time.RFC3339))

    // Proactive rate limit checking
    if info.Remaining < 10 {
        log.Println("Warning: Approaching rate limit")
    }
}
```

### RateLimitInfo Structure

```go
type RateLimitInfo struct {
    Limit     int       // Maximum requests allowed in window
    Remaining int       // Requests remaining in current window
    Reset     time.Time // When the rate limit window resets
}
```

## Retry-After Behavior

When you hit the rate limit (HTTP 429), the client intelligently handles the `Retry-After` header:

1. **Server-provided delay**: If the server includes a `Retry-After` header, the client uses that exact delay (capped at 30 seconds)
2. **Exponential backoff**: If no `Retry-After` header is present, uses longer backoff periods (5s, 10s, 20s, capped at 30s)
3. **Automatic retry**: Rate limit errors are automatically retried up to `MaxRetries` times

```go
// Configure retry behavior
client := wirepusher.NewClient(
    "your-token",
    wirepusher.WithMaxRetries(5), // Default is 3
)

// Rate limit handling is automatic
err := client.Send(ctx, options)
if err != nil {
    if rateLimitErr, ok := err.(*wirepusher.RateLimitError); ok {
        // Only reached if all retries exhausted
        log.Printf("Rate limit exceeded after retries. Retry-After: %ds", rateLimitErr.RetryAfter)
    }
}
```

### Backoff Strategy

| Error Type | Attempt 1 | Attempt 2 | Attempt 3 | Maximum |
|------------|-----------|-----------|-----------|---------|
| Network/5xx | 1s | 2s | 4s | 30s |
| Rate Limit (no header) | 5s | 10s | 20s | 30s |
| Rate Limit (with header) | Retry-After value | Retry-After value | Retry-After value | 30s |

## Custom Timeout Configuration

### Per-Client Timeout

```go
// Set timeout during client creation
client := wirepusher.NewClient(
    "your-token",
    wirepusher.WithTimeout(60*time.Second), // 60 second timeout
)
```

### Environment Variable Configuration

```bash
export WIREPUSHER_TIMEOUT=120  # 120 seconds
export WIREPUSHER_MAX_RETRIES=5
```

```go
// Client automatically respects environment variables
client := wirepusher.NewClient("")  // Uses WIREPUSHER_TOKEN from env
// Also uses WIREPUSHER_TIMEOUT and WIREPUSHER_MAX_RETRIES if set
```

### Per-Request Timeout (Context)

```go
// Override timeout for specific request
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := client.Send(ctx, options)
if ctx.Err() == context.DeadlineExceeded {
    log.Println("Request timed out after 10 seconds")
}
```

### Custom HTTP Client

```go
// Full control over HTTP behavior
httpClient := &http.Client{
    Timeout: 45 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

client := wirepusher.NewClient(
    "your-token",
    wirepusher.WithHTTPClient(httpClient),
)
```

## Error Handling Patterns

### Comprehensive Error Type Handling

```go
err := client.Send(ctx, options)
if err != nil {
    switch e := err.(type) {
    case *wirepusher.ValidationError:
        // Invalid input (400)
        log.Printf("Validation failed: %s", e.Message)
        // Fix input and retry

    case *wirepusher.AuthError:
        // Invalid token (401/403)
        log.Printf("Authentication failed: %s", e.Message)
        // Check token, do not retry

    case *wirepusher.RateLimitError:
        // Rate limit exceeded (429)
        log.Printf("Rate limited: %s (retry after %ds)", e.Message, e.RetryAfter)
        // Wait and retry later

    case *wirepusher.ServerError:
        // Server error (5xx) - retryable
        log.Printf("Server error: %s", e.Message)
        // Already retried automatically

    case *wirepusher.NetworkError:
        // Connection issue - retryable
        if originalErr := e.Unwrap(); originalErr != nil {
            log.Printf("Network error: %v", originalErr)
        }
        // Already retried automatically

    default:
        log.Printf("Unexpected error: %v", err)
    }
}
```

### Checking Retryability

```go
err := client.Send(ctx, options)
if err != nil && wirepusher.IsErrorRetryable(err) {
    log.Println("Error was retryable but all attempts exhausted")
}
```

### Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel after some condition
go func() {
    <-someCondition
    cancel()
}()

err := client.Send(ctx, options)
if err == context.Canceled {
    log.Println("Request was cancelled")
}
```

## Logging

Enable debug logging to see retry attempts and internal operations:

```go
import "log"

// Create a custom logger
type DebugLogger struct{}

func (l *DebugLogger) Debug(msg string) {
    log.Printf("[DEBUG] %s", msg)
}

func (l *DebugLogger) Info(msg string) {
    log.Printf("[INFO] %s", msg)
}

func (l *DebugLogger) Warning(msg string) {
    log.Printf("[WARN] %s", msg)
}

func (l *DebugLogger) Error(msg string) {
    log.Printf("[ERROR] %s", msg)
}

// Use with client
client := wirepusher.NewClient(
    "your-token",
    wirepusher.WithLogger(&DebugLogger{}),
)
```

## AES-128-CBC Encryption

Messages can be encrypted client-side before sending:

```go
err := client.Send(ctx, &wirepusher.SendOptions{
    Title:              "Secure Alert",           // NOT encrypted
    Message:            "Sensitive data here",    // ENCRYPTED
    Type:               "security",               // NOT encrypted
    Tags:               []string{"confidential"}, // NOT encrypted
    EncryptionPassword: "your-password",
})
```

Key points:
- Only the `Message` field is encrypted
- Uses AES-128-CBC with random IV
- Password is hashed with SHA-1 to derive key
- IV is transmitted alongside encrypted message
- No external dependencies (uses Go standard library)

## Go-Specific Features

### Zero External Dependencies

The library uses only Go standard library packages:
- `net/http` for HTTP requests
- `encoding/json` for JSON serialization
- `context` for cancellation and timeouts
- `crypto/aes`, `crypto/cipher`, `crypto/sha1` for encryption
- `time` for timing and backoff

### Full Context Support

All public methods accept a `context.Context` as the first parameter, enabling:
- Request cancellation
- Request timeouts
- Request-scoped values
- Graceful shutdown handling

### Type-Safe Error Handling

All errors implement specific types with the `IsRetryable() bool` method:
- Type assertions work reliably
- Error unwrapping via `errors.Unwrap()` for network errors
- Structured error information with status codes and messages

### Functional Options Pattern

Client configuration uses the functional options pattern:
```go
client := wirepusher.NewClient(
    token,
    wirepusher.WithTimeout(30*time.Second),
    wirepusher.WithMaxRetries(5),
    wirepusher.WithHTTPClient(customClient),
    wirepusher.WithLogger(logger),
)
```

This allows:
- Clear, readable configuration
- Optional parameters without breaking changes
- Easy extension for future options

### Tag Normalization

Tags are automatically normalized:
- Trimmed of whitespace
- Empty tags removed
- Preserves order of valid tags

```go
tags := []string{"  production ", "", "backend", "  "}
// Normalized to: ["production", "backend"]
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `WIREPUSHER_TOKEN` | API token | Required if not passed to constructor |
| `WIREPUSHER_TIMEOUT` | Request timeout in seconds | 30 |
| `WIREPUSHER_MAX_RETRIES` | Maximum retry attempts | 3 |

## Best Practices

1. **Use environment variables** for tokens and configuration
2. **Always pass context** - even `context.Background()` for simple cases
3. **Check rate limits proactively** before critical operations
4. **Handle all error types** - provide specific handling for each error type
5. **Use appropriate timeouts** - balance between reliability and responsiveness
6. **Enable logging in development** - helps debug retry behavior
7. **Test error scenarios** - ensure your code handles all failure modes
