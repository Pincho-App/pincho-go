# CLAUDE.md - Pincho Go Client Library

Context file for AI-powered development assistance on the Pincho Go Client Library project.

## Project Overview

**Pincho Go Client Library** is a Go client library for sending push notifications via [Pincho](https://pincho.app).

- **Language**: Go 1.18+
- **Framework**: Standard library only (net/http, encoding/json, context)
- **Purpose**: Send notifications from Go applications, services, and servers
- **Philosophy**: Zero external dependencies, idiomatic Go, context-aware

## Architecture

```
pincho-go/
├── client.go              # Main client implementation
├── errors.go              # Error types with sentinel errors
├── logger.go              # Structured logging
├── validation.go          # Tag normalization
├── crypto.go              # AES-128-CBC encryption
├── client_test.go         # Client tests (errors.Is/errors.As)
├── validation_test.go     # Validation tests
├── logger_test.go         # Logger tests
├── go.mod                  # Module definition
├── cloudbuild.yaml        # Cloud Build CI/CD
├── examples/              # Usage examples
│   ├── basic/             # Simple send example
│   ├── advanced/          # Full options
│   ├── context/           # Context cancellation/timeout
│   ├── errors/            # Error handling patterns
│   ├── encryption/        # Message encryption
│   └── rate-limits/       # Rate limit monitoring
├── docs/                  # Documentation
│   ├── ADVANCED.md        # Rate limits, config, encryption
│   ├── CODE_OF_CONDUCT.md # Code of conduct
│   └── SECURITY.md        # Security policy
└── README.md              # ~120 lines, concise
```

## Key Features

### 1. Zero External Dependencies

Uses only Go standard library:
- `net/http` for HTTP requests
- `encoding/json` for JSON serialization
- `context` for cancellation and timeouts
- `crypto/sha1`, `crypto/aes` for encryption
- `log/slog` for structured logging

### 2. Context-Aware API

All methods accept `context.Context`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := client.Send(ctx, &pincho.SendOptions{
    Title:   "Deploy Complete",
    Message: "v1.2.3 deployed",
})
```

### 3. Functional Options Pattern

Configure client with functional options:

```go
client := pincho.NewClient("abc12345",
    pincho.WithTimeout(60*time.Second),
    pincho.WithMaxRetries(5),
    pincho.WithAPIURL("https://custom.api.com"),
)
```

### 4. Sentinel Errors (Go 1.13+)

Use `errors.Is()` and `errors.As()` patterns:

```go
err := client.Send(ctx, options)
if errors.Is(err, pincho.ErrAuth) {
    log.Println("Authentication failed")
} else if errors.Is(err, pincho.ErrRateLimit) {
    log.Println("Rate limited")
}

// Extract error details
var rateLimitErr *pincho.RateLimitError
if errors.As(err, &rateLimitErr) {
    log.Printf("Retry after %d seconds", rateLimitErr.RetryAfter)
}
```

Sentinel errors:
- `ErrAuth` - Authentication errors (401/403)
- `ErrValidation` - Validation errors (400)
- `ErrRateLimit` - Rate limit errors (429)
- `ErrServer` - Server errors (5xx)
- `ErrNetwork` - Network/connection errors

### 5. Error Types with IsRetryable()

All error types implement `RetryableError` interface:

- `AuthError` - Not retryable
- `ValidationError` - Not retryable
- `RateLimitError` - Retryable with backoff
- `ServerError` - Retryable
- `NetworkError` - Retryable

### 6. Automatic Retry Logic

- **Default**: 3 retries with exponential backoff
- **Backoff strategy**: 1s, 2s, 4s, 8s (capped at 30s)
- **Rate limit handling**: Uses Retry-After header if present, otherwise 5s, 10s, 20s
- **Retryable**: Network errors, 5xx, 429
- **Non-retryable**: 400, 401, 403, 404

### 7. Tag Normalization

Automatic validation in `NormalizeTags()`:
- Lowercase conversion
- Whitespace trimming
- Duplicate removal (case-insensitive)
- Character validation (alphanumeric, hyphens, underscores only)
- Enforces limits: max 10 tags, 50 characters per tag

### 8. Rate Limit Monitoring

```go
err := client.Send(ctx, options)

// Check rate limit info after any request
if client.LastRateLimit != nil {
    fmt.Printf("Remaining: %d/%d, Resets: %s\n",
        client.LastRateLimit.Remaining,
        client.LastRateLimit.Limit,
        client.LastRateLimit.Reset)
}
```

### 9. NotifAI Endpoint

AI-powered notification generation:

```go
response, err := client.NotifAI(ctx, &pincho.NotifAIOptions{
    Text: "deployment finished, v2.1.3 is live",
})
// response.Notification contains AI-generated title, message, tags
```

### 10. AES-128-CBC Encryption

Message encryption matching mobile app:

```go
err := client.Send(ctx, &pincho.SendOptions{
    Title:              "Security Alert",
    Message:            "Sensitive data",
    Type:               "security",
    EncryptionPassword: "your_password",
})
```

## Configuration

Environment variable support with functional options override:

```go
// Auto-load from environment variables
client := pincho.NewClient("")  // reads PINCHO_TOKEN

// Or explicit configuration
client := pincho.NewClient("abc12345",
    pincho.WithTimeout(60*time.Second),
    pincho.WithMaxRetries(5),
)
```

**Environment Variables**:
- `PINCHO_TOKEN` - API token (used if constructor token is empty)
- `PINCHO_TIMEOUT` - Request timeout in seconds (default: 30)
- `PINCHO_MAX_RETRIES` - Maximum retry attempts (default: 3)

## Dependencies

**Runtime**: ZERO (standard library only)
- `net/http`
- `encoding/json`
- `context`
- `crypto/sha1`, `crypto/aes`, `crypto/cipher`, `crypto/rand`
- `log/slog`

**Development**: None required (Go tooling)

## Recent Changes

### v1.0.0 (Current)

**Added**:
- Sentinel errors with `errors.Is()` and `errors.As()` support
- `Is()` methods on all error types for Go 1.13+ patterns
- NotifAI endpoint for AI-powered notifications
- Automatic retry logic with exponential backoff
- Rate limit monitoring via `LastRateLimit`
- Tag normalization (lowercase, trim, dedupe, validate)
- AES-128-CBC encryption support
- Environment variable configuration
- Structured logging with `log/slog`
- Comprehensive documentation and examples

**Error Types**:
- `AuthError` (401/403) - `IsRetryable() = false`
- `ValidationError` (400) - `IsRetryable() = false`
- `RateLimitError` (429) - `IsRetryable() = true`
- `ServerError` (5xx) - `IsRetryable() = true`
- `NetworkError` (network) - `IsRetryable() = true`

## Development

### Setup

```bash
# Clone repository
git clone https://gitlab.com/pincho-app/pincho-go.git
cd pincho-go

# No dependencies to install!
go mod verify
```

### Testing

```bash
go test -v ./...                       # All tests
go test -race ./...                    # With race detector
go test -cover ./...                   # With coverage
go test -coverprofile=coverage.out ./...  # Coverage report
go tool cover -html=coverage.out       # View coverage in browser
```

### Code Quality

```bash
go vet ./...                           # Static analysis
gofmt -l .                             # Check formatting
gofmt -w .                             # Auto-format
```

### Building

```bash
go build ./...                         # Build all packages
go build -v ./examples/basic           # Build specific example
```

## Common Development Tasks

### Adding a Feature

1. Implement in appropriate file (`client.go`, etc.)
2. Add functional option if configurable (`WithXxx`)
3. Update SendOptions/NotifAIOptions if needed
4. Add tests with comprehensive coverage
5. Update README with examples
6. Add to CHANGELOG

### Adding an Error Type

1. Add struct to `errors.go`
2. Implement `Error()` string method
3. Implement `IsRetryable()` bool method
4. Implement `Is(target error)` bool method
5. Add sentinel error variable (e.g., `ErrXxx`)
6. Update `handleHTTPError()` in client.go
7. Add tests for errors.Is() and errors.As()

### Adding Validation

1. Add logic to `validation.go`
2. Write comprehensive tests in `validation_test.go`
3. Integrate into client methods
4. Document behavior in README

## Testing Philosophy

- **Unit tests**: Test functions in isolation
- **Table-driven tests**: Use `[]struct{ name string; ... }` pattern
- **Error testing**: Test both errors.Is() and errors.As() patterns
- **Race detection**: Run with `-race` flag
- **Coverage target**: 85%+ for critical paths
- **Context testing**: Test cancellation and timeouts

## API Integration

### Endpoints

- `POST /send` - Send notifications
- `POST /notifai` - AI-powered notifications

### Authentication

Bearer token via `Authorization` header:
```
Authorization: Bearer {token}
User-Agent: pincho-go/{version}
```

### Response Format

**Success response:**
```json
{
  "status": "success",
  "message": "Notification sent successfully"
}
```

**Error response:**
```json
{
  "status": "error",
  "error": {
    "type": "validation_error",
    "code": "missing_required_field",
    "message": "Title is required",
    "param": "title"
  }
}
```

The library parses the nested error format and builds descriptive error messages with code and param information.

## Notes for AI Assistants

- **Zero dependencies**: Keep runtime dependency-free (stdlib only)
- **Context-aware**: All network methods must accept context.Context
- **Functional options**: Use `With...` pattern for configuration
- **Error patterns**: Implement errors.Is() via `Is()` method
- **Table-driven tests**: Use Go's standard test patterns
- **Idiomatic Go**: Follow Effective Go guidelines
- **No config files**: Configuration via env vars or constructor
- **Test coverage**: Aim for 90%+ on critical paths
- **Documentation**: Update README for user-facing changes
- **Race detector**: Test with `-race` flag regularly

## Project Status

**Current**: Production-ready v1.0.0 with comprehensive feature set

**Completed**:
- ✅ Zero external dependencies (stdlib only)
- ✅ Context-aware API
- ✅ Sentinel errors with errors.Is/errors.As
- ✅ Automatic retry logic with exponential backoff
- ✅ Tag normalization
- ✅ AES-128-CBC encryption
- ✅ NotifAI endpoint
- ✅ Rate limit monitoring
- ✅ Structured logging
- ✅ Environment variable support
- ✅ Comprehensive documentation and examples
- ✅ CI/CD with Cloud Build

**Not Needed**:
- ❌ Config file support (not standard for libraries)
- ❌ CLI tool (separate pincho-cli project)

## Links

- **Repository**: https://gitlab.com/pincho-app/pincho-go
- **Issues**: https://gitlab.com/pincho-app/pincho-go/-/issues
- **pkg.go.dev**: https://pkg.go.dev/gitlab.com/pincho-app/pincho-go
- **API Docs**: https://pincho.app/help
- **App**: https://pincho.app
