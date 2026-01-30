package pincho

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNoOpLogger verifies that NoOpLogger discards all log messages.
func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}

	// These should not panic and should do nothing
	logger.Printf("test message %s", "arg")
	logger.Println("test message")
}

// TestStdLogger verifies that StdLogger writes to the provided writer.
func TestStdLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := &StdLogger{
		logger: log.New(&buf, "wirepusher: ", log.LstdFlags),
	}

	logger.Printf("test %s", "message")
	output := buf.String()

	if !strings.Contains(output, "wirepusher:") {
		t.Errorf("Expected prefix 'wirepusher:' in output, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected 'test message' in output, got: %s", output)
	}
}

// TestNewStdLogger verifies that NewStdLogger creates a logger with the correct prefix.
func TestNewStdLogger(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		expectedPrefix string
	}{
		{
			name:           "with prefix",
			prefix:         "wirepusher",
			expectedPrefix: "wirepusher: ",
		},
		{
			name:           "empty prefix",
			prefix:         "",
			expectedPrefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewStdLogger(tt.prefix)
			if logger == nil {
				t.Fatal("NewStdLogger returned nil")
			}
			if logger.logger == nil {
				t.Fatal("NewStdLogger.logger is nil")
			}
		})
	}
}

// TestWithLogger verifies that WithLogger option sets the logger.
func TestWithLogger(t *testing.T) {
	customLogger := &NoOpLogger{}
	client := NewClient("test-token", WithLogger(customLogger))

	if client.Logger != customLogger {
		t.Errorf("Expected custom logger to be set")
	}
}

// TestClientLoggingMethods verifies that client logging methods call the logger.
func TestClientLoggingMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := &StdLogger{
		logger: log.New(&buf, "", 0), // No prefix or timestamp for easier testing
	}

	client := NewClient("test-token", WithLogger(logger))

	tests := []struct {
		name           string
		logFunc        func()
		expectedPrefix string
	}{
		{
			name:           "logDebug",
			logFunc:        func() { client.logDebug("debug message") },
			expectedPrefix: "DEBUG:",
		},
		{
			name:           "logInfo",
			logFunc:        func() { client.logInfo("info message") },
			expectedPrefix: "INFO:",
		},
		{
			name:           "logWarning",
			logFunc:        func() { client.logWarning("warning message") },
			expectedPrefix: "WARNING:",
		},
		{
			name:           "logError",
			logFunc:        func() { client.logError("error message") },
			expectedPrefix: "ERROR:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()

			if !strings.Contains(output, tt.expectedPrefix) {
				t.Errorf("Expected %s in output, got: %s", tt.expectedPrefix, output)
			}
		})
	}
}

// TestLoggingInSend verifies that Send() logs appropriately.
func TestLoggingInSend(t *testing.T) {
	var buf bytes.Buffer
	logger := &StdLogger{
		logger: log.New(&buf, "", 0),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewClient("test-token",
		WithLogger(logger),
		WithAPIURL(server.URL),
	)

	// Test: Send with tags that need normalization
	buf.Reset()
	err := client.Send(context.Background(), &SendOptions{
		Title:   "Test",
		Message: "Test message",
		Tags:    []string{"Production", "RELEASE", "production"},
	})

	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	output := buf.String()

	// Should log the Send() call
	if !strings.Contains(output, "DEBUG: Send() called with title: Test") {
		t.Errorf("Expected Send() debug log, got: %s", output)
	}

	// Should log tag normalization
	if !strings.Contains(output, "Tags normalized:") {
		t.Errorf("Expected tag normalization log, got: %s", output)
	}
}

// TestLoggingInSendWithEncryption verifies that Send() logs encryption.
func TestLoggingInSendWithEncryption(t *testing.T) {
	var buf bytes.Buffer
	logger := &StdLogger{
		logger: log.New(&buf, "", 0),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewClient("test-token",
		WithLogger(logger),
		WithAPIURL(server.URL),
	)

	// Test: Send with encryption
	buf.Reset()
	err := client.Send(context.Background(), &SendOptions{
		Title:              "Test",
		Message:            "Test message",
		EncryptionPassword: "secret123",
	})

	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	output := buf.String()

	// Should log encryption
	if !strings.Contains(output, "DEBUG: Encrypting message") {
		t.Errorf("Expected encryption log, got: %s", output)
	}
}

// TestLoggingInNotifAI verifies that NotifAI() logs appropriately.
func TestLoggingInNotifAI(t *testing.T) {
	var buf bytes.Buffer
	logger := &StdLogger{
		logger: log.New(&buf, "", 0),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "title": "Test", "message": "Test message", "type": "info", "tags": ["test"]}`))
	}))
	defer server.Close()

	client := NewClient("test-token",
		WithLogger(logger),
		WithAPIURL(server.URL+"/send"),
	)

	// Test: NotifAI call
	buf.Reset()
	_, err := client.NotifAI(context.Background(), &NotifAIOptions{
		Text: "test notification",
	})

	if err != nil {
		t.Fatalf("NotifAI failed: %v", err)
	}

	output := buf.String()

	// Should log the NotifAI() call
	if !strings.Contains(output, "DEBUG: NotifAI() called with text: test notification") {
		t.Errorf("Expected NotifAI() debug log, got: %s", output)
	}
}

// TestLoggingInRetryWithBackoff verifies retry logging.
func TestLoggingInRetryWithBackoff(t *testing.T) {
	var buf bytes.Buffer
	logger := &StdLogger{
		logger: log.New(&buf, "", 0),
	}

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			// First attempt fails with 500
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "server error"}`))
		} else {
			// Second attempt succeeds
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		}
	}))
	defer server.Close()

	client := NewClient("test-token",
		WithLogger(logger),
		WithAPIURL(server.URL),
		WithMaxRetries(3),
	)

	buf.Reset()
	err := client.SendSimple(context.Background(), "Test", "Test message")

	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	output := buf.String()

	// Should log retry attempt
	if !strings.Contains(output, "Retry attempt") {
		t.Errorf("Expected retry log, got: %s", output)
	}

	// Should log backoff
	if !strings.Contains(output, "backing off for") {
		t.Errorf("Expected backoff log, got: %s", output)
	}
}

// TestLoggingInRetryMaxRetriesExceeded verifies max retries logging.
func TestLoggingInRetryMaxRetriesExceeded(t *testing.T) {
	var buf bytes.Buffer
	logger := &StdLogger{
		logger: log.New(&buf, "", 0),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always fail with 500
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "server error"}`))
	}))
	defer server.Close()

	client := NewClient("test-token",
		WithLogger(logger),
		WithAPIURL(server.URL),
		WithMaxRetries(2),
	)

	buf.Reset()
	err := client.SendSimple(context.Background(), "Test", "Test message")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	output := buf.String()

	// Should log max retries exceeded
	if !strings.Contains(output, "Max retries") && !strings.Contains(output, "exceeded") {
		t.Errorf("Expected max retries log, got: %s", output)
	}
}

// TestLoggingInRetryNonRetryableError verifies non-retryable error logging.
func TestLoggingInRetryNonRetryableError(t *testing.T) {
	var buf bytes.Buffer
	logger := &StdLogger{
		logger: log.New(&buf, "", 0),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 401 (non-retryable)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "invalid token"}`))
	}))
	defer server.Close()

	client := NewClient("test-token",
		WithLogger(logger),
		WithAPIURL(server.URL),
		WithMaxRetries(3),
	)

	buf.Reset()
	err := client.SendSimple(context.Background(), "Test", "Test message")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	output := buf.String()

	// Should log non-retryable error
	if !strings.Contains(output, "Error not retryable") {
		t.Errorf("Expected non-retryable error log, got: %s", output)
	}

	// Should NOT log retry attempts (since error is not retryable)
	if strings.Contains(output, "Retry attempt 1") {
		t.Errorf("Should not retry non-retryable errors, got: %s", output)
	}
}

// TestLoggingInRetryRateLimit verifies rate limit logging.
func TestLoggingInRetryRateLimit(t *testing.T) {
	var buf bytes.Buffer
	logger := &StdLogger{
		logger: log.New(&buf, "", 0),
	}

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			// First attempt fails with 429
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message": "rate limit exceeded"}`))
		} else {
			// Second attempt succeeds
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		}
	}))
	defer server.Close()

	client := NewClient("test-token",
		WithLogger(logger),
		WithAPIURL(server.URL),
		WithMaxRetries(3),
	)

	buf.Reset()
	err := client.SendSimple(context.Background(), "Test", "Test message")

	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	output := buf.String()

	// Should log rate limit
	if !strings.Contains(output, "Rate limit hit") {
		t.Errorf("Expected rate limit log, got: %s", output)
	}
}

// TestTruncateToken verifies token truncation for safe logging.
func TestTruncateToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "long token",
			token:    "abc12345678",
			expected: "abc1...",
		},
		{
			name:     "short token",
			token:    "abc",
			expected: "abc",
		},
		{
			name:     "exactly 4 chars",
			token:    "abcd",
			expected: "abcd",
		},
		{
			name:     "5 chars",
			token:    "abcde",
			expected: "abcd...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateToken(tt.token)
			if result != tt.expected {
				t.Errorf("truncateToken(%s) = %s, want %s", tt.token, result, tt.expected)
			}
		})
	}
}
