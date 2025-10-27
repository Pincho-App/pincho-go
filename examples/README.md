# WirePusher Go SDK Examples

This directory contains runnable examples demonstrating various features of the WirePusher Go SDK.

## Prerequisites

Before running any examples, you need:

1. A WirePusher account with API credentials
2. Environment variables set (choose either token OR user ID):
   ```bash
   # Personal notifications
   export WIREPUSHER_USER_ID="your-user-id"

   # OR team notifications
   export WIREPUSHER_TOKEN="wpt_your_token"

   # For encryption examples, also set:
   export WIREPUSHER_ENCRYPTION_PASSWORD="your_password"
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
- Encrypting message content with password
- Only message is encrypted (title/type/tags remain plaintext)
- Password management with environment variables
- Multiple encryption scenarios
- Backward compatibility (unencrypted messages still work)
- Uses only Go standard library (no external dependencies)

## Example Code Structure

Each example follows the same structure:

```go
package main

import (
    "context"
    "log"
    "os"

    "gitlab.com/wirepusher/go-sdk"
)

func main() {
    // 1. Get credentials from environment (choose either token OR userID)
    userID := os.Getenv("WIREPUSHER_USER_ID")

    // 2. Create client
    client := wirepusher.NewClient("", userID)

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

# Run the built binary
./basic
```

## Integration with Your Project

To use the SDK in your own project:

1. Install the SDK:
   ```bash
   go get gitlab.com/wirepusher/go-sdk
   ```

2. Import and use:
   ```go
   import "gitlab.com/wirepusher/go-sdk"

   client := wirepusher.NewClient("", userID)  // Use "" for either token or userID
   err := client.SendSimple(ctx, "Title", "Message")
   ```

## Tips

- **Environment Variables**: Use a `.env` file or a tool like `direnv` to manage your credentials
- **Error Handling**: Always check errors and use type assertions to handle different error types
- **Context**: Always pass a context (use `context.Background()` if you don't need cancellation/timeout)
- **Testing**: Use a test user ID when trying out examples to avoid spamming your production devices

## Additional Resources

- [API Reference](https://pkg.go.dev/gitlab.com/wirepusher/go-sdk)
- [Main README](../README.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
