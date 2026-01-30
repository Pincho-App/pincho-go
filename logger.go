package pincho

import (
	"fmt"
	"log"
	"os"
)

// Logger is an interface for logging within the Pincho client.
// This allows users to integrate their own logging solution.
//
// The interface matches Go's standard log.Logger methods for compatibility.
type Logger interface {
	// Printf logs a formatted message (INFO level).
	Printf(format string, v ...interface{})

	// Println logs a message with a newline (INFO level).
	Println(v ...interface{})
}

// NoOpLogger is a logger that discards all log messages.
// This is the default logger for the client.
type NoOpLogger struct{}

// Printf discards the log message.
func (l *NoOpLogger) Printf(format string, v ...interface{}) {}

// Println discards the log message.
func (l *NoOpLogger) Println(v ...interface{}) {}

// StdLogger is a logger that writes to standard output.
// This provides basic console logging similar to Python's logging module.
type StdLogger struct {
	logger *log.Logger
}

// NewStdLogger creates a new standard logger that writes to stdout.
// Prefix is optional and will be prepended to all log messages.
//
// Example:
//
//	logger := pincho.NewStdLogger("pincho")
//	client := pincho.NewClient("abc12345", pincho.WithLogger(logger))
func NewStdLogger(prefix string) *StdLogger {
	if prefix != "" {
		prefix = prefix + ": "
	}
	return &StdLogger{
		logger: log.New(os.Stdout, prefix, log.LstdFlags),
	}
}

// Printf logs a formatted message to stdout.
func (l *StdLogger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

// Println logs a message with a newline to stdout.
func (l *StdLogger) Println(v ...interface{}) {
	l.logger.Println(v...)
}

// WithLogger sets a custom logger for the client.
//
// Example with standard logger:
//
//	logger := pincho.NewStdLogger("pincho")
//	client := pincho.NewClient("abc12345", pincho.WithLogger(logger))
//
// Example with custom logger:
//
//	type MyLogger struct{}
//	func (l *MyLogger) Printf(format string, v ...interface{}) {
//	    // Custom logging implementation
//	}
//	func (l *MyLogger) Println(v ...interface{}) {
//	    // Custom logging implementation
//	}
//
//	logger := &MyLogger{}
//	client := pincho.NewClient("abc12345", pincho.WithLogger(logger))
func WithLogger(logger Logger) ClientOption {
	return func(c *Client) {
		c.Logger = logger
	}
}

// logf is a helper method for formatted logging.
func (c *Client) logf(format string, v ...interface{}) {
	if c.Logger != nil {
		c.Logger.Printf(format, v...)
	}
}

// logln is a helper method for line logging.
func (c *Client) logln(v ...interface{}) {
	if c.Logger != nil {
		c.Logger.Println(v...)
	}
}

// logDebug logs a debug message if logging is enabled.
// This mimics Python's logger.debug() behavior.
func (c *Client) logDebug(message string) {
	c.logf("DEBUG: %s", message)
}

// logInfo logs an info message if logging is enabled.
// This mimics Python's logger.info() behavior.
func (c *Client) logInfo(message string) {
	c.logf("INFO: %s", message)
}

// logWarning logs a warning message if logging is enabled.
// This mimics Python's logger.warning() behavior.
func (c *Client) logWarning(message string) {
	c.logf("WARNING: %s", message)
}

// logError logs an error message if logging is enabled.
// This mimics Python's logger.error() behavior.
func (c *Client) logError(message string) {
	c.logf("ERROR: %s", message)
}

// truncateToken returns a truncated token for safe logging.
// Shows first 4 chars only: "abc12345" -> "abc1..."
func truncateToken(token string) string {
	if len(token) <= 4 {
		return token
	}
	return fmt.Sprintf("%s...", token[:4])
}
