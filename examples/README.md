# Pincho Go Client Library Examples

This directory contains runnable examples demonstrating various features of the Pincho Go Client Library.

## Prerequisites

Before running any examples, you need:

1. A Pincho account with API token
2. Environment variable set:
   ```bash
   export PINCHO_TOKEN="abc12345"  # Your 8-character token

   # For encryption examples, also set:
   export PINCHO_ENCRYPTION_PASSWORD="your_password"
   ```

## Running Examples

Each example can be run independently using `go run`:

### Basic Usage

Simple notification with just a title and message:

```bash
cd examples/basic
go run main.go
```

**What it demonstrates:**
- Creating a client
- Sending a simple notification with `SendSimple()`

### Advanced Options

Notifications with types, tags, images, and action URLs:

```bash
cd examples/advanced
go run main.go
```

**What it demonstrates:**
- Using `Send()` with full options
- Setting notification type
- Adding tags
- Including image URLs
- Including action URLs

### Context Usage

Using contexts for timeouts and cancellation:

```bash
cd examples/context
go run main.go
```

**What it demonstrates:**
- Request timeouts with `context.WithTimeout()`
- Request cancellation with `context.WithCancel()`
- Custom client timeout with `WithTimeout()` option

### Error Handling

Handling different error types:

```bash
cd examples/errors
go run main.go
```

**What it demonstrates:**
- Validation errors (empty title/message)
- Authentication errors (invalid token)
- Type switching for different error types
- Graceful error handling patterns

### Encrypted Notifications

Sending notifications with AES-128-CBC encryption:

```bash
cd examples/encryption
go run main.go
```

**What it demonstrates:**
- Encrypting title, message, imageURL, actionURL with password
- Type and tags remain unencrypted (needed for filtering/routing)
- Password management with environment variables
- Multiple encryption scenarios
- Backward compatibility (unencrypted notifications still work)
- Uses only Go standard library (no external dependencies)

### Rate Limit Monitoring

Monitoring and utilizing rate limit information:

```bash
cd examples/rate-limits
go run main.go
```

**What it demonstrates:**
- Accessing `client.LastRateLimit` after requests
- Reading limit, remaining, and reset time
- Proactive rate limit checking before exhaustion
- Smart scheduling based on available quota
- Calculating sustainable sending rates

## Example Code Structure

Each example follows the same structure:

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/pincho/pincho-go"
)

func main() {
    // 1. Get token from environment
    token := os.Getenv("PINCHO_TOKEN")
    if token == "" {
        token = "abc12345" // Fallback for testing
    }

    // 2. Create client
    client := pincho.NewClient(token)

    // 3. Send notification
    err := client.SendSimple(context.Background(), "Title", "Message")
    if err != nil {
        log.Fatal(err)
    }
}
```

## Building Examples

You can also build the examples into standalone binaries:

```bash
# Build all examples
cd examples/basic && go build -o basic
cd ../advanced && go build -o advanced
cd ../context && go build -o context
cd ../errors && go build -o errors
cd ../encryption && go build -o encryption
cd ../rate-limits && go build -o rate-limits

# Run the built binary
./basic
```

## Integration with Your Project

To use the SDK in your own project:

1. Install the SDK:
   ```bash
   go get gitlab.com/pincho/pincho-go
   ```

2. Import and use:
   ```go
   import "gitlab.com/pincho/pincho-go"

   client := pincho.NewClient(token)
   err := client.SendSimple(ctx, "Title", "Message")
   ```

## Tips

- **Environment Variables**: Use a `.env` file or a tool like `direnv` to manage your credentials
- **Error Handling**: Always check errors and use type assertions to handle different error types
- **Context**: Always pass a context (use `context.Background()` if you don't need cancellation/timeout)
- **Fallback Token**: Examples use "abc12345" as fallback if PINCHO_TOKEN is not set

## Additional Resources

- [API Reference](https://pkg.go.dev/gitlab.com/pincho/pincho-go)
- [Main README](../README.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
