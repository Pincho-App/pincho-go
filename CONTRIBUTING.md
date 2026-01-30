# Contributing to Pincho Go Client Library

Thank you for your interest in contributing to the Pincho Go Client Library! We welcome contributions from the community.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

Before creating a bug report:
- Check the existing issues to avoid duplicates
- Collect information about your environment (Go version, OS, etc.)
- Provide a minimal reproduction case

When creating a bug report, use the Bug issue template and include:
- Clear description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment details
- Code examples
- Error messages (with sensitive information removed)

### Suggesting Features

Before suggesting a feature:
- Check if it has already been requested
- Consider if it fits the project scope
- Think about how it would benefit the SDK users

When suggesting a feature, use the Feature issue template and include:
- Clear description of the feature
- Problem it solves
- Proposed solution
- Example usage
- Benefits to users

### Contributing Code

We welcome code contributions! Here's how to get started:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting
5. Commit your changes (follow commit conventions)
6. Push to your fork
7. Open a Merge Request

## Development Setup

### Prerequisites

- Go 1.18 or higher
- Git

### Initial Setup

```bash
# Clone your fork
git clone git@gitlab.com:your-username/go-sdk.git
cd go-sdk

# Download dependencies
go mod download

# Verify setup
go test ./...
```

### Recommended Tools

- **gofmt**: Code formatting (built-in with Go)
- **go vet**: Static analysis (built-in with Go)
- **golangci-lint**: Advanced linting (optional but recommended)

Install golangci-lint:
```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Windows
# See: https://golangci-lint.run/usage/install/
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test improvements

### 2. Make Changes

- Write clear, idiomatic Go code
- Follow the [Coding Standards](#coding-standards)
- Add tests for new functionality
- Update documentation as needed

### 3. Format Code

```bash
# Format all Go files
gofmt -w .

# Verify formatting
gofmt -l .
```

### 4. Run Static Analysis

```bash
# Run go vet
go vet ./...

# Run golangci-lint (if installed)
golangci-lint run
```

### 5. Run Tests

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 6. Commit Changes

Follow conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting (no code change)
- `refactor`: Code refactoring
- `test`: Adding/updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(client): add support for custom HTTP headers

Allows users to set custom headers on all requests.

Closes #123
```

```
fix(send): handle nil context gracefully

Previously would panic if nil context was passed.
Now returns a ValidationError instead.

Fixes #456
```

### 7. Push and Create MR

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create MR on GitLab
# Go to: https://gitlab.com/pincho/pincho-go/-/merge_requests/new
```

## Pull Request Process

1. **Ensure all checks pass**:
   - All tests pass
   - Code is formatted with `gofmt`
   - No warnings from `go vet`
   - Coverage is maintained (≥90%)

2. **Update documentation**:
   - Update README if API changes
   - Add/update godoc comments
   - Update examples if needed
   - Add entry to CHANGELOG.md

3. **Fill out the MR template**:
   - Describe the changes
   - Link related issues
   - List breaking changes (if any)
   - Add testing notes

4. **Request review**:
   - Wait for review from maintainers
   - Address feedback promptly
   - Make requested changes

5. **Merge**:
   - Squash commits if needed
   - Maintainer will merge after approval

## Coding Standards

### General Principles

- **Simplicity**: Keep code simple and readable
- **Idiomatic Go**: Follow Go conventions and best practices
- **Clear naming**: Use descriptive names for variables, functions, types
- **Error handling**: Always handle errors explicitly
- **Documentation**: Document exported functions and types

### Go-Specific Guidelines

#### Formatting

- Use `gofmt` for all code
- Use tabs for indentation
- Keep lines under 100 characters when practical

#### Naming Conventions

```go
// ✅ Good
var userID string
func SendNotification() error
type HTTPClient struct {}

// ❌ Bad
var user_id string
func send_notification() error
type Http_Client struct {}
```

#### Error Handling

```go
// ✅ Good - Always handle errors
resp, err := client.Send(ctx, options)
if err != nil {
    return fmt.Errorf("send failed: %w", err)
}

// ❌ Bad - Ignoring errors
resp, _ := client.Send(ctx, options)
```

#### Context Usage

```go
// ✅ Good - Context as first parameter
func (c *Client) Send(ctx context.Context, options *SendOptions) error

// ❌ Bad - Context not first or missing
func (c *Client) Send(options *SendOptions, ctx context.Context) error
func (c *Client) Send(options *SendOptions) error
```

#### Comments

```go
// ✅ Good - Complete sentences, proper godoc format
// Send sends a notification with full options.
//
// The options parameter must include at least Title and Message.
// Returns an error if the request fails.
func (c *Client) Send(ctx context.Context, options *SendOptions) error

// ❌ Bad - Incomplete or unclear
// send notification
func (c *Client) Send(ctx context.Context, options *SendOptions) error
```

#### Struct Tags

```go
// ✅ Good - Use json tags for API types
type SendOptions struct {
    Title   string `json:"title"`
    Message string `json:"message"`
}
```

#### Exported vs Unexported

```go
// ✅ Good - Public API is exported
type Client struct {
    Token  string // Exported
    apiURL string // Internal
}

// Public method
func (c *Client) Send(ctx context.Context, options *SendOptions) error

// Internal helper
func (c *Client) buildRequest(options *SendOptions) (*http.Request, error)
```

## Testing

### Test Coverage

- Maintain ≥90% test coverage
- Test happy paths and error cases
- Test edge cases and boundary conditions

### Test Organization

```go
func TestClient_Send(t *testing.T) {
    // Use subtests for organization
    t.Run("successful send", func(t *testing.T) {
        // Test code
    })

    t.Run("validation error", func(t *testing.T) {
        // Test code
    })
}
```

### Table-Driven Tests

```go
// ✅ Good - Table-driven tests for similar cases
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {"valid input", "test", true, false},
        {"empty input", "", false, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr %v, got %v", tt.wantErr, err)
            }
            if got != tt.want {
                t.Errorf("want %v, got %v", tt.want, got)
            }
        })
    }
}
```

### Test Helpers

```go
// Use testify or custom helpers for cleaner tests
func assertEqual(t *testing.T, got, want interface{}) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

## Documentation

### godoc Comments

All exported types and functions must have godoc comments:

```go
// Client is the Pincho API client.
//
// Use NewClient() to create a new instance.
type Client struct {
    // Token is the Pincho API token (required).
    Token string

    // UserID is the Pincho user ID (required).
    UserID string
}

// NewClient creates a new Pincho client.
//
// The token and userID parameters are required.
// Optional configuration can be provided using ClientOption functions.
//
// Example:
//
//	client := pincho.NewClient("token", "user-id")
func NewClient(token, userID string, opts ...ClientOption) *Client
```

### README Updates

When adding features:
- Update the README with new examples
- Add to the API Reference section
- Update the Quick Start if relevant

### Examples

When adding features:
- Add code examples to `examples/` directory
- Ensure examples are runnable
- Document example usage in `examples/README.md`

### Changelog

For every change, add an entry to CHANGELOG.md:

```markdown
## [Unreleased]

### Added
- New feature X that does Y

### Changed
- Modified behavior of Z

### Fixed
- Fixed bug where A caused B
```

## Questions?

If you have questions about contributing:
- Open a discussion on GitLab
- Email support@pincho.com
- Check existing issues and MRs

Thank you for contributing to Pincho Go Client Library!
