package pincho

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("with token", func(t *testing.T) {
		client := NewClient("abc12345")

		if client.Token != "abc12345" {
			t.Errorf("expected token 'abc12345', got '%s'", client.Token)
		}

		if client.APIURL != DefaultAPIURL {
			t.Errorf("expected APIURL '%s', got '%s'", DefaultAPIURL, client.APIURL)
		}

		if client.HTTPClient == nil {
			t.Error("expected HTTPClient to be initialized")
		}

		if client.HTTPClient.Timeout != DefaultTimeout {
			t.Errorf("expected timeout %v, got %v", DefaultTimeout, client.HTTPClient.Timeout)
		}
	})

	t.Run("panics with empty token", func(t *testing.T) {
		// Clear environment variable to ensure panic
		oldToken := os.Getenv("PINCHO_TOKEN")
		os.Unsetenv("PINCHO_TOKEN")
		defer os.Setenv("PINCHO_TOKEN", oldToken)

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected NewClient to panic when token is empty")
			}
		}()
		NewClient("")
	})

	t.Run("with custom API URL", func(t *testing.T) {
		customURL := "https://custom.example.com/api"
		client := NewClient("abc12345", WithAPIURL(customURL))

		if client.APIURL != customURL {
			t.Errorf("expected APIURL '%s', got '%s'", customURL, client.APIURL)
		}
	})

	t.Run("with custom timeout", func(t *testing.T) {
		customTimeout := 5 * time.Second
		client := NewClient("abc12345", WithTimeout(customTimeout))

		if client.HTTPClient.Timeout != customTimeout {
			t.Errorf("expected timeout %v, got %v", customTimeout, client.HTTPClient.Timeout)
		}
	})

	t.Run("with custom HTTP client", func(t *testing.T) {
		customClient := &http.Client{Timeout: 10 * time.Second}
		client := NewClient("abc12345", WithHTTPClient(customClient))

		if client.HTTPClient != customClient {
			t.Error("expected custom HTTP client to be used")
		}
	})

	t.Run("panics with empty API URL", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected WithAPIURL to panic when URL is empty")
			}
		}()
		NewClient("abc12345", WithAPIURL(""))
	})

	t.Run("panics with nil HTTP client", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected WithHTTPClient to panic when client is nil")
			}
		}()
		NewClient("abc12345", WithHTTPClient(nil))
	})

	t.Run("panics with zero timeout", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected WithTimeout to panic when timeout is zero")
			}
		}()
		NewClient("abc12345", WithTimeout(0))
	})

	t.Run("panics with negative timeout", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected WithTimeout to panic when timeout is negative")
			}
		}()
		NewClient("abc12345", WithTimeout(-1*time.Second))
	})

	t.Run("panics with negative max retries", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected WithMaxRetries to panic when max retries is negative")
			}
		}()
		NewClient("abc12345", WithMaxRetries(-1))
	})

	t.Run("allows zero max retries", func(t *testing.T) {
		client := NewClient("abc12345", WithMaxRetries(0))
		if client.MaxRetries != 0 {
			t.Errorf("expected MaxRetries to be 0, got %d", client.MaxRetries)
		}
	})

	t.Run("reads token from env var", func(t *testing.T) {
		os.Setenv("PINCHO_TOKEN", "env_token_123")
		defer os.Unsetenv("PINCHO_TOKEN")

		client := NewClient("")

		if client.Token != "env_token_123" {
			t.Errorf("expected token 'env_token_123', got '%s'", client.Token)
		}
	})

	t.Run("reads timeout from env var", func(t *testing.T) {
		os.Setenv("PINCHO_TIMEOUT", "60")
		defer os.Unsetenv("PINCHO_TIMEOUT")

		client := NewClient("abc12345")

		expectedTimeout := 60 * time.Second
		if client.HTTPClient.Timeout != expectedTimeout {
			t.Errorf("expected timeout %v, got %v", expectedTimeout, client.HTTPClient.Timeout)
		}
	})

	t.Run("reads max retries from env var", func(t *testing.T) {
		os.Setenv("PINCHO_MAX_RETRIES", "10")
		defer os.Unsetenv("PINCHO_MAX_RETRIES")

		client := NewClient("abc12345")

		if client.MaxRetries != 10 {
			t.Errorf("expected MaxRetries 10, got %d", client.MaxRetries)
		}
	})

	t.Run("explicit token overrides env var", func(t *testing.T) {
		os.Setenv("PINCHO_TOKEN", "env_token_123")
		defer os.Unsetenv("PINCHO_TOKEN")

		client := NewClient("explicit_token")

		if client.Token != "explicit_token" {
			t.Errorf("expected token 'explicit_token', got '%s'", client.Token)
		}
	})

	t.Run("option overrides env var timeout", func(t *testing.T) {
		os.Setenv("PINCHO_TIMEOUT", "60")
		defer os.Unsetenv("PINCHO_TIMEOUT")

		client := NewClient("abc12345", WithTimeout(120*time.Second))

		expectedTimeout := 120 * time.Second
		if client.HTTPClient.Timeout != expectedTimeout {
			t.Errorf("expected timeout %v, got %v", expectedTimeout, client.HTTPClient.Timeout)
		}
	})

	t.Run("option overrides env var max retries", func(t *testing.T) {
		os.Setenv("PINCHO_MAX_RETRIES", "10")
		defer os.Unsetenv("PINCHO_MAX_RETRIES")

		client := NewClient("abc12345", WithMaxRetries(20))

		if client.MaxRetries != 20 {
			t.Errorf("expected MaxRetries 20, got %d", client.MaxRetries)
		}
	})

	t.Run("ignores invalid timeout env var", func(t *testing.T) {
		os.Setenv("PINCHO_TIMEOUT", "invalid")
		defer os.Unsetenv("PINCHO_TIMEOUT")

		client := NewClient("abc12345")

		if client.HTTPClient.Timeout != DefaultTimeout {
			t.Errorf("expected default timeout %v, got %v", DefaultTimeout, client.HTTPClient.Timeout)
		}
	})

	t.Run("ignores negative timeout env var", func(t *testing.T) {
		os.Setenv("PINCHO_TIMEOUT", "-5")
		defer os.Unsetenv("PINCHO_TIMEOUT")

		client := NewClient("abc12345")

		if client.HTTPClient.Timeout != DefaultTimeout {
			t.Errorf("expected default timeout %v, got %v", DefaultTimeout, client.HTTPClient.Timeout)
		}
	})

	t.Run("ignores invalid max retries env var", func(t *testing.T) {
		os.Setenv("PINCHO_MAX_RETRIES", "invalid")
		defer os.Unsetenv("PINCHO_MAX_RETRIES")

		client := NewClient("abc12345")

		if client.MaxRetries != 3 {
			t.Errorf("expected default MaxRetries 3, got %d", client.MaxRetries)
		}
	})

	t.Run("ignores negative max retries env var", func(t *testing.T) {
		os.Setenv("PINCHO_MAX_RETRIES", "-5")
		defer os.Unsetenv("PINCHO_MAX_RETRIES")

		client := NewClient("abc12345")

		if client.MaxRetries != 3 {
			t.Errorf("expected default MaxRetries 3, got %d", client.MaxRetries)
		}
	})
}

func TestClient_SendSimple(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		message       string
		serverStatus  int
		serverBody    string
		expectError   bool
		errorContains string
	}{
		{
			name:         "successful send",
			title:        "Test Title",
			message:      "Test Message",
			serverStatus: 200,
			serverBody:   `{"status": "success", "message": "Notification sent"}`,
			expectError:  false,
		},
		{
			name:          "validation error",
			title:         "",
			message:       "Test Message",
			serverStatus:  400,
			serverBody:    `{"status": "error", "error": {"type": "validation_error", "code": "missing_field", "message": "title is required", "param": "title"}}`,
			expectError:   true,
			errorContains: "title is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverBody))
			}))
			defer server.Close()

			client := NewClient("abc12345", WithAPIURL(server.URL))
			err := client.SendSimple(context.Background(), tt.title, tt.message)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestClient_Send(t *testing.T) {
	t.Run("successful send with all options", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify method
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}

			// Verify Content-Type
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			// Verify Authorization header
			if r.Header.Get("Authorization") != "Bearer abc12345" {
				t.Errorf("expected Authorization 'Bearer abc12345', got '%s'", r.Header.Get("Authorization"))
			}

			// Verify User-Agent header
			expectedUA := "pincho-go/" + Version
			if r.Header.Get("User-Agent") != expectedUA {
				t.Errorf("expected User-Agent '%s', got '%s'", expectedUA, r.Header.Get("User-Agent"))
			}

			// Parse body
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)

			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success", "message": "Notification sent"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		options := &SendOptions{
			Title:     "Test Title",
			Message:   "Test Message",
			Type:      "alert",
			Tags:      []string{"tag1", "tag2"},
			ImageURL:  "https://example.com/image.png",
			ActionURL: "https://example.com/action",
		}

		err := client.Send(context.Background(), options)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Verify all fields were sent
		if receivedBody["title"] != "Test Title" {
			t.Errorf("expected title 'Test Title', got '%v'", receivedBody["title"])
		}

		if receivedBody["message"] != "Test Message" {
			t.Errorf("expected message 'Test Message', got '%v'", receivedBody["message"])
		}

		if receivedBody["type"] != "alert" {
			t.Errorf("expected type 'alert', got '%v'", receivedBody["type"])
		}

		tags := receivedBody["tags"].([]interface{})
		if len(tags) != 2 || tags[0] != "tag1" || tags[1] != "tag2" {
			t.Errorf("expected tags [tag1 tag2], got %v", tags)
		}

		if receivedBody["imageURL"] != "https://example.com/image.png" {
			t.Errorf("expected imageURL, got '%v'", receivedBody["imageURL"])
		}

		if receivedBody["actionURL"] != "https://example.com/action" {
			t.Errorf("expected actionURL, got '%v'", receivedBody["actionURL"])
		}

		// Token should NOT be in body (now sent via Authorization header)
		if _, hasToken := receivedBody["token"]; hasToken {
			t.Errorf("token should not be in body, got '%v'", receivedBody["token"])
		}
	})

	t.Run("message is optional", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)

			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success", "message": "Notification sent"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		// Send with title only, no message
		err := client.Send(context.Background(), &SendOptions{
			Title: "Test Title",
			// No message provided
		})

		if err != nil {
			t.Fatalf("expected no error for title-only notification, got: %v", err)
		}

		if receivedBody["title"] != "Test Title" {
			t.Errorf("expected title 'Test Title', got '%v'", receivedBody["title"])
		}

		// Message field should be empty string (not sent or sent as "")
		if msg, ok := receivedBody["message"]; ok && msg != "" {
			t.Errorf("expected message to be empty or absent, got '%v'", msg)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		client := NewClient("abc12345")

		tests := []struct {
			name          string
			options       *SendOptions
			errorContains string
			errorType     interface{}
		}{
			{
				name:          "nil options",
				options:       nil,
				errorContains: "options cannot be nil",
				errorType:     &ValidationError{},
			},
			{
				name:          "empty title",
				options:       &SendOptions{Title: "", Message: "test"},
				errorContains: "title is required",
				errorType:     &ValidationError{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := client.Send(context.Background(), tt.options)

				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}

				// Check error type
				switch tt.errorType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("expected ValidationError, got %T", err)
					}
				}
			})
		}
	})

	t.Run("HTTP error responses", func(t *testing.T) {
		tests := []struct {
			name          string
			statusCode    int
			responseBody  string
			errorType     interface{}
			errorContains string
		}{
			{
				name:          "400 bad request",
				statusCode:    400,
				responseBody:  `{"status": "error", "error": {"type": "validation_error", "code": "invalid_request", "message": "Invalid request"}}`,
				errorType:     &ValidationError{},
				errorContains: "Invalid request",
			},
			{
				name:          "401 unauthorized",
				statusCode:    401,
				responseBody:  `{"status": "error", "error": {"type": "auth_error", "code": "unauthorized", "message": "Unauthorized"}}`,
				errorType:     &AuthError{},
				errorContains: "Unauthorized",
			},
			{
				name:          "403 forbidden",
				statusCode:    403,
				responseBody:  `{"status": "error", "error": {"type": "auth_error", "code": "forbidden", "message": "Forbidden"}}`,
				errorType:     &AuthError{},
				errorContains: "Forbidden",
			},
			{
				name:          "429 rate limit",
				statusCode:    429,
				responseBody:  `{"status": "error", "error": {"type": "rate_limit_error", "code": "rate_limit_exceeded", "message": "Rate limit exceeded"}}`,
				errorType:     &RateLimitError{},
				errorContains: "Rate limit exceeded",
			},
			{
				name:          "500 server error",
				statusCode:    500,
				responseBody:  `{"status": "error", "error": {"type": "server_error", "code": "internal_error", "message": "Internal server error"}}`,
				errorType:     &ServerError{},
				errorContains: "Internal server error",
			},
			{
				name:          "502 bad gateway",
				statusCode:    502,
				responseBody:  `{"status": "error", "error": {"type": "server_error", "code": "bad_gateway", "message": "Bad Gateway"}}`,
				errorType:     &ServerError{},
				errorContains: "Bad Gateway",
			},
			{
				name:          "503 service unavailable",
				statusCode:    503,
				responseBody:  "Service Unavailable",
				errorType:     &ServerError{},
				errorContains: "Service Unavailable",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
					w.Write([]byte(tt.responseBody))
				}))
				defer server.Close()

				// Disable retries for error testing
				client := NewClient("abc12345", WithAPIURL(server.URL), WithMaxRetries(0))

				err := client.Send(context.Background(), &SendOptions{
					Title:   "Test",
					Message: "Test",
				})

				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}

				// Check error type
				switch tt.errorType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("expected ValidationError, got %T", err)
					}
				case *AuthError:
					if _, ok := err.(*AuthError); !ok {
						t.Errorf("expected AuthError, got %T", err)
					}
				case *RateLimitError:
					if _, ok := err.(*RateLimitError); !ok {
						t.Errorf("expected RateLimitError, got %T", err)
					}
				case *ServerError:
					if _, ok := err.(*ServerError); !ok {
						t.Errorf("expected ServerError, got %T", err)
					}
				case *Error:
					if _, ok := err.(*Error); !ok {
						t.Errorf("expected Error, got %T", err)
					}
				}
			})
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := client.Send(ctx, &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err == nil {
			t.Fatal("expected error due to context cancellation, got nil")
		}

		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("expected context cancellation error, got: %v", err)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := client.Send(ctx, &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err == nil {
			t.Fatal("expected error due to context timeout, got nil")
		}

		if !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("expected context timeout error, got: %v", err)
		}
	})

	t.Run("network error", func(t *testing.T) {
		// Use invalid URL to trigger network error
		client := NewClient("abc12345", WithAPIURL("http://localhost:1"))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err == nil {
			t.Fatal("expected network error, got nil")
		}

		if _, ok := err.(*NetworkError); !ok {
			t.Errorf("expected NetworkError type, got %T", err)
		}

		// Verify error wrapping
		netErr := err.(*NetworkError)
		if netErr.Err == nil {
			t.Error("expected NetworkError.Err to be set")
		}
	})

	t.Run("successful send with non-JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		// Should not error even though response isn't JSON
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}

func TestEncryption(t *testing.T) {
	t.Run("derive encryption key", func(t *testing.T) {
		password := "test_password_123"
		key, err := DeriveEncryptionKey(password)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(key) != 16 {
			t.Errorf("expected key length 16, got %d", len(key))
		}

		// Verify key derivation is deterministic
		key2, _ := DeriveEncryptionKey(password)
		if string(key) != string(key2) {
			t.Error("key derivation is not deterministic")
		}
	})

	t.Run("encrypt message with fixed IV", func(t *testing.T) {
		plaintext := "This is a secret message that needs to be encrypted securely."
		password := "test_password_123"
		ivHex := "0123456789abcdef0123456789abcdef"

		iv, err := hex.DecodeString(ivHex)
		if err != nil {
			t.Fatalf("failed to decode IV: %v", err)
		}

		encrypted, err := EncryptMessage(plaintext, password, iv)
		if err != nil {
			t.Fatalf("encryption failed: %v", err)
		}

		// Verify it uses custom Base64 encoding (contains -, ., or _)
		hasCustomChars := strings.Contains(encrypted, "-") || strings.Contains(encrypted, ".") || strings.Contains(encrypted, "_")
		if !hasCustomChars {
			t.Error("encrypted output should use custom Base64 encoding")
		}

		// Verify encryption is deterministic with same IV
		encrypted2, _ := EncryptMessage(plaintext, password, iv)
		if encrypted != encrypted2 {
			t.Error("encryption with same IV should be deterministic")
		}

		// Expected value verified with reference implementation
		// Key: 5a0aee0f3af308cd6d74d617fde6592c (from SHA1 of password)
		// Plaintext length: 61 bytes, padded to 64 bytes (pad_length=3)
		expected := "y2fzGqnZSgdMqkwYhAUEZi30VFBYvwcCmrQ6BmSliPpPGHXMdMRsLCtG-cfwhhxN4HSIk5Y3UMjM6XoBWPqiHw__"
		if encrypted != expected {
			t.Errorf("encrypted output doesn't match expected.\nGot:      %s\nExpected: %s", encrypted, expected)
		}
	})

	t.Run("encrypt different messages produce different outputs", func(t *testing.T) {
		password := "test_password"
		iv, _, _ := GenerateIV()

		encrypted1, _ := EncryptMessage("message 1", password, iv)
		encrypted2, _ := EncryptMessage("message 2", password, iv)

		if encrypted1 == encrypted2 {
			t.Error("different messages should produce different encrypted outputs")
		}
	})

	t.Run("generate IV", func(t *testing.T) {
		iv1, ivHex1, err := GenerateIV()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(iv1) != 16 {
			t.Errorf("expected IV length 16, got %d", len(iv1))
		}

		if len(ivHex1) != 32 {
			t.Errorf("expected IV hex length 32, got %d", len(ivHex1))
		}

		// Verify IVs are random (different each time)
		_, ivHex2, _ := GenerateIV()
		if ivHex1 == ivHex2 {
			t.Error("generated IVs should be random")
		}
	})

	t.Run("custom base64 encoding", func(t *testing.T) {
		// Test that custom encoding replaces special characters correctly
		iv, _, _ := GenerateIV()
		plaintext := "Test message with padding"
		password := "password"

		encrypted, _ := EncryptMessage(plaintext, password, iv)

		// Should not contain standard Base64 special characters
		if strings.Contains(encrypted, "+") {
			t.Error("encrypted output should not contain '+' (should be '-')")
		}
		if strings.Contains(encrypted, "/") {
			t.Error("encrypted output should not contain '/' (should be '.')")
		}
		if strings.Contains(encrypted, "=") {
			t.Error("encrypted output should not contain '=' (should be '_')")
		}
	})

	t.Run("send with encryption", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		err := client.Send(context.Background(), &SendOptions{
			Title:              "Test",
			Message:            "Secret message",
			EncryptionPassword: "test_password",
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Verify title is encrypted (not plaintext)
		if receivedBody["title"] == "Test" {
			t.Error("title should be encrypted, got plaintext")
		}

		// Verify message is encrypted (not plaintext)
		if receivedBody["message"] == "Secret message" {
			t.Error("message should be encrypted, got plaintext")
		}

		// Verify IV was included
		if receivedBody["iv"] == nil || receivedBody["iv"] == "" {
			t.Error("IV should be included in request when encryption is used")
		}

		ivHex, ok := receivedBody["iv"].(string)
		if !ok || len(ivHex) != 32 {
			t.Errorf("IV should be 32-character hex string, got: %v", receivedBody["iv"])
		}
	})

	t.Run("send_with_encryption_all_fields", func(t *testing.T) {
		var receivedBody map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		err := client.Send(context.Background(), &SendOptions{
			Title:              "Test Title",
			Message:            "Secret message",
			Type:               "secure",
			Tags:               []string{"test", "encryption"},
			ImageURL:           "https://example.com/image.png",
			ActionURL:          "https://example.com/action",
			EncryptionPassword: "test_password",
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Verify all encrypted fields are not plaintext
		if receivedBody["title"] == "Test Title" {
			t.Error("title should be encrypted")
		}
		if receivedBody["message"] == "Secret message" {
			t.Error("message should be encrypted")
		}
		if receivedBody["imageURL"] == "https://example.com/image.png" {
			t.Error("imageURL should be encrypted")
		}
		if receivedBody["actionURL"] == "https://example.com/action" {
			t.Error("actionURL should be encrypted")
		}

		// Verify type and tags remain unencrypted
		if receivedBody["type"] != "secure" {
			t.Errorf("type should remain unencrypted, got: %v", receivedBody["type"])
		}
		if tags, ok := receivedBody["tags"].([]interface{}); !ok || len(tags) != 2 {
			t.Errorf("tags should remain unencrypted, got: %v", receivedBody["tags"])
		}
	})
}

func TestErrorTypes(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		err := &Error{Message: "test error", StatusCode: 500}
		expected := "pincho: test error (status: 500)"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("Error without status code", func(t *testing.T) {
		err := &Error{Message: "test error", StatusCode: 0}
		expected := "pincho: test error"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("AuthError", func(t *testing.T) {
		err := &AuthError{Message: "unauthorized", StatusCode: 401}
		expected := "pincho auth error: unauthorized (status: 401)"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := &ValidationError{Message: "invalid input", StatusCode: 400}
		expected := "pincho validation error: invalid input (status: 400)"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("RateLimitError", func(t *testing.T) {
		err := &RateLimitError{Message: "rate limit exceeded", StatusCode: 429}
		expected := "pincho rate limit error: rate limit exceeded (status: 429)"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("ServerError", func(t *testing.T) {
		err := &ServerError{Message: "internal server error", StatusCode: 500}
		expected := "pincho server error: internal server error (status: 500)"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("NetworkError", func(t *testing.T) {
		originalErr := fmt.Errorf("connection refused")
		err := &NetworkError{Message: "request failed", Err: originalErr}
		expected := "pincho network error: request failed: connection refused"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("NetworkError without wrapped error", func(t *testing.T) {
		err := &NetworkError{Message: "request failed", Err: nil}
		expected := "pincho network error: request failed"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("NetworkError Unwrap", func(t *testing.T) {
		originalErr := fmt.Errorf("connection refused")
		err := &NetworkError{Message: "request failed", Err: originalErr}
		unwrapped := err.Unwrap()
		if unwrapped != originalErr {
			t.Errorf("expected unwrapped error to be original error")
		}
	})
}

func TestErrorsIs(t *testing.T) {
	t.Run("AuthError matches ErrAuth", func(t *testing.T) {
		err := &AuthError{Message: "unauthorized", StatusCode: 401}
		if !errors.Is(err, ErrAuth) {
			t.Error("expected AuthError to match ErrAuth")
		}
	})

	t.Run("AuthError does not match ErrValidation", func(t *testing.T) {
		err := &AuthError{Message: "unauthorized", StatusCode: 401}
		if errors.Is(err, ErrValidation) {
			t.Error("expected AuthError to not match ErrValidation")
		}
	})

	t.Run("ValidationError matches ErrValidation", func(t *testing.T) {
		err := &ValidationError{Message: "invalid", StatusCode: 400}
		if !errors.Is(err, ErrValidation) {
			t.Error("expected ValidationError to match ErrValidation")
		}
	})

	t.Run("RateLimitError matches ErrRateLimit", func(t *testing.T) {
		err := &RateLimitError{Message: "too many requests", StatusCode: 429}
		if !errors.Is(err, ErrRateLimit) {
			t.Error("expected RateLimitError to match ErrRateLimit")
		}
	})

	t.Run("ServerError matches ErrServer", func(t *testing.T) {
		err := &ServerError{Message: "internal error", StatusCode: 500}
		if !errors.Is(err, ErrServer) {
			t.Error("expected ServerError to match ErrServer")
		}
	})

	t.Run("NetworkError matches ErrNetwork", func(t *testing.T) {
		err := &NetworkError{Message: "connection failed", Err: nil}
		if !errors.Is(err, ErrNetwork) {
			t.Error("expected NetworkError to match ErrNetwork")
		}
	})

	t.Run("sentinel errors work directly", func(t *testing.T) {
		if !errors.Is(ErrAuth, ErrAuth) {
			t.Error("expected ErrAuth to match itself")
		}
		if !errors.Is(ErrValidation, ErrValidation) {
			t.Error("expected ErrValidation to match itself")
		}
		if !errors.Is(ErrRateLimit, ErrRateLimit) {
			t.Error("expected ErrRateLimit to match itself")
		}
		if !errors.Is(ErrServer, ErrServer) {
			t.Error("expected ErrServer to match itself")
		}
		if !errors.Is(ErrNetwork, ErrNetwork) {
			t.Error("expected ErrNetwork to match itself")
		}
	})
}

func TestErrorsAs(t *testing.T) {
	t.Run("AuthError can be extracted with errors.As", func(t *testing.T) {
		var err error = &AuthError{Message: "unauthorized", StatusCode: 401}
		var authErr *AuthError
		if !errors.As(err, &authErr) {
			t.Error("expected to extract AuthError with errors.As")
		}
		if authErr.StatusCode != 401 {
			t.Errorf("expected StatusCode 401, got %d", authErr.StatusCode)
		}
	})

	t.Run("ValidationError can be extracted with errors.As", func(t *testing.T) {
		var err error = &ValidationError{Message: "invalid", StatusCode: 400}
		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Error("expected to extract ValidationError with errors.As")
		}
		if validationErr.Message != "invalid" {
			t.Errorf("expected Message 'invalid', got '%s'", validationErr.Message)
		}
	})

	t.Run("RateLimitError can be extracted with errors.As", func(t *testing.T) {
		var err error = &RateLimitError{Message: "too many", StatusCode: 429, RetryAfter: 60}
		var rateLimitErr *RateLimitError
		if !errors.As(err, &rateLimitErr) {
			t.Error("expected to extract RateLimitError with errors.As")
		}
		if rateLimitErr.RetryAfter != 60 {
			t.Errorf("expected RetryAfter 60, got %d", rateLimitErr.RetryAfter)
		}
	})

	t.Run("ServerError can be extracted with errors.As", func(t *testing.T) {
		var err error = &ServerError{Message: "internal", StatusCode: 500}
		var serverErr *ServerError
		if !errors.As(err, &serverErr) {
			t.Error("expected to extract ServerError with errors.As")
		}
		if serverErr.StatusCode != 500 {
			t.Errorf("expected StatusCode 500, got %d", serverErr.StatusCode)
		}
	})

	t.Run("NetworkError can be extracted with errors.As", func(t *testing.T) {
		originalErr := fmt.Errorf("connection refused")
		var err error = &NetworkError{Message: "failed", Err: originalErr}
		var networkErr *NetworkError
		if !errors.As(err, &networkErr) {
			t.Error("expected to extract NetworkError with errors.As")
		}
		if networkErr.Err != originalErr {
			t.Error("expected Err to be original error")
		}
	})
}

func TestIsErrorRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "ServerError is retryable",
			err:       &ServerError{Message: "internal error", StatusCode: 500},
			retryable: true,
		},
		{
			name:      "NetworkError is retryable",
			err:       &NetworkError{Message: "connection failed", Err: fmt.Errorf("test")},
			retryable: true,
		},
		{
			name:      "RateLimitError is retryable",
			err:       &RateLimitError{Message: "rate limit", StatusCode: 429},
			retryable: true,
		},
		{
			name:      "AuthError is not retryable",
			err:       &AuthError{Message: "unauthorized", StatusCode: 401},
			retryable: false,
		},
		{
			name:      "ValidationError is not retryable",
			err:       &ValidationError{Message: "invalid", StatusCode: 400},
			retryable: false,
		},
		{
			name:      "Error is not retryable",
			err:       &Error{Message: "generic error", StatusCode: 0},
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsErrorRetryable(tt.err)
			if result != tt.retryable {
				t.Errorf("expected IsErrorRetryable() = %v, got %v", tt.retryable, result)
			}
		})
	}
}

func TestRetryAfterParsing(t *testing.T) {
	t.Run("parses valid Retry-After header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(429)
			w.Write([]byte(`{"status": "error", "error": {"type": "rate_limit_error", "code": "rate_limit_exceeded", "message": "Rate limit exceeded"}}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL), WithMaxRetries(0))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		rateLimitErr, ok := err.(*RateLimitError)
		if !ok {
			t.Fatalf("expected RateLimitError, got %T", err)
		}

		if rateLimitErr.RetryAfter != 60 {
			t.Errorf("expected RetryAfter to be 60, got %d", rateLimitErr.RetryAfter)
		}
	})

	t.Run("handles missing Retry-After header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(429)
			w.Write([]byte(`{"status": "error", "error": {"type": "rate_limit_error", "code": "rate_limit_exceeded", "message": "Rate limit exceeded"}}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL), WithMaxRetries(0))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		rateLimitErr, ok := err.(*RateLimitError)
		if !ok {
			t.Fatalf("expected RateLimitError, got %T", err)
		}

		if rateLimitErr.RetryAfter != 0 {
			t.Errorf("expected RetryAfter to be 0 when header is missing, got %d", rateLimitErr.RetryAfter)
		}
	})

	t.Run("handles invalid Retry-After header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "invalid")
			w.WriteHeader(429)
			w.Write([]byte(`{"status": "error", "error": {"type": "rate_limit_error", "code": "rate_limit_exceeded", "message": "Rate limit exceeded"}}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL), WithMaxRetries(0))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		rateLimitErr, ok := err.(*RateLimitError)
		if !ok {
			t.Fatalf("expected RateLimitError, got %T", err)
		}

		if rateLimitErr.RetryAfter != 0 {
			t.Errorf("expected RetryAfter to be 0 when header is invalid, got %d", rateLimitErr.RetryAfter)
		}
	})

	t.Run("handles negative Retry-After header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "-10")
			w.WriteHeader(429)
			w.Write([]byte(`{"status": "error", "error": {"type": "rate_limit_error", "code": "rate_limit_exceeded", "message": "Rate limit exceeded"}}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL), WithMaxRetries(0))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		rateLimitErr, ok := err.(*RateLimitError)
		if !ok {
			t.Fatalf("expected RateLimitError, got %T", err)
		}

		if rateLimitErr.RetryAfter != 0 {
			t.Errorf("expected RetryAfter to be 0 when header is negative, got %d", rateLimitErr.RetryAfter)
		}
	})
}

func TestRateLimitInfoParsing(t *testing.T) {
	t.Run("parses all rate limit headers", func(t *testing.T) {
		resetTime := time.Now().Add(1 * time.Hour).Unix()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("RateLimit-Limit", "100")
			w.Header().Set("RateLimit-Remaining", "95")
			w.Header().Set("RateLimit-Reset", fmt.Sprintf("%d", resetTime))
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success", "message": "Notification sent"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		info := client.GetRateLimitInfo()
		if info == nil {
			t.Fatal("expected rate limit info, got nil")
		}

		if info.Limit != 100 {
			t.Errorf("expected Limit to be 100, got %d", info.Limit)
		}

		if info.Remaining != 95 {
			t.Errorf("expected Remaining to be 95, got %d", info.Remaining)
		}

		if info.Reset.Unix() != resetTime {
			t.Errorf("expected Reset to be %d, got %d", resetTime, info.Reset.Unix())
		}
	})

	t.Run("handles missing rate limit headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success", "message": "Notification sent"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		info := client.GetRateLimitInfo()
		if info != nil {
			t.Errorf("expected nil rate limit info when headers are missing, got %+v", info)
		}
	})

	t.Run("handles partial rate limit headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("RateLimit-Limit", "100")
			// Missing RateLimit-Remaining and RateLimit-Reset
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success", "message": "Notification sent"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		info := client.GetRateLimitInfo()
		if info == nil {
			t.Fatal("expected rate limit info with partial headers, got nil")
		}

		if info.Limit != 100 {
			t.Errorf("expected Limit to be 100, got %d", info.Limit)
		}

		if info.Remaining != 0 {
			t.Errorf("expected Remaining to be 0 when header is missing, got %d", info.Remaining)
		}

		if !info.Reset.IsZero() {
			t.Errorf("expected Reset to be zero when header is missing, got %v", info.Reset)
		}
	})

	t.Run("handles invalid rate limit headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("RateLimit-Limit", "invalid")
			w.Header().Set("RateLimit-Remaining", "not-a-number")
			w.Header().Set("RateLimit-Reset", "bad-timestamp")
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success", "message": "Notification sent"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		info := client.GetRateLimitInfo()
		if info != nil {
			t.Errorf("expected nil rate limit info when all headers are invalid, got %+v", info)
		}
	})

	t.Run("updates rate limit info on subsequent requests", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			remaining := 100 - requestCount
			w.Header().Set("RateLimit-Limit", "100")
			w.Header().Set("RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("RateLimit-Reset", "1700000000")
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success", "message": "Notification sent"}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		// First request
		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test 1",
			Message: "Test",
		})
		if err != nil {
			t.Fatalf("first request failed: %v", err)
		}

		info1 := client.GetRateLimitInfo()
		if info1 == nil || info1.Remaining != 99 {
			t.Errorf("expected Remaining to be 99 after first request, got %v", info1)
		}

		// Second request
		err = client.Send(context.Background(), &SendOptions{
			Title:   "Test 2",
			Message: "Test",
		})
		if err != nil {
			t.Fatalf("second request failed: %v", err)
		}

		info2 := client.GetRateLimitInfo()
		if info2 == nil || info2.Remaining != 98 {
			t.Errorf("expected Remaining to be 98 after second request, got %v", info2)
		}
	})

	t.Run("rate limit info parsed in NotifAI", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("RateLimit-Limit", "50")
			w.Header().Set("RateLimit-Remaining", "45")
			w.Header().Set("RateLimit-Reset", "1700000000")
			w.WriteHeader(200)
			w.Write([]byte(`{
				"status": "success",
				"message": "Notification generated and sent",
				"notification": {
					"title": "Test",
					"message": "Test message"
				}
			}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		_, err := client.NotifAI(context.Background(), &NotifAIOptions{
			Text: "test notification",
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		info := client.GetRateLimitInfo()
		if info == nil {
			t.Fatal("expected rate limit info, got nil")
		}

		if info.Limit != 50 {
			t.Errorf("expected Limit to be 50, got %d", info.Limit)
		}

		if info.Remaining != 45 {
			t.Errorf("expected Remaining to be 45, got %d", info.Remaining)
		}
	})
}

func TestClient_NotifAI(t *testing.T) {
	t.Run("successful notifai", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify method
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}

			// Verify Content-Type
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			// Verify Authorization header
			if r.Header.Get("Authorization") != "Bearer abc12345" {
				t.Errorf("expected Authorization 'Bearer abc12345', got '%s'", r.Header.Get("Authorization"))
			}

			// Parse body
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)

			w.WriteHeader(200)
			w.Write([]byte(`{
				"status": "success",
				"message": "Notification generated and sent",
				"notification": {
					"title": "Deploy Complete",
					"message": "Version 2.1.3 is live on prod",
					"type": "deployment",
					"tags": ["production", "release"]
				}
			}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		response, err := client.NotifAI(context.Background(), &NotifAIOptions{
			Text: "deployment finished successfully, v2.1.3 is live on prod",
			Type: "deployment",
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Verify request body
		if receivedBody["text"] != "deployment finished successfully, v2.1.3 is live on prod" {
			t.Errorf("expected text, got '%v'", receivedBody["text"])
		}

		if receivedBody["type"] != "deployment" {
			t.Errorf("expected type 'deployment', got '%v'", receivedBody["type"])
		}

		// Token should NOT be in body (now sent via Authorization header)
		if _, hasToken := receivedBody["token"]; hasToken {
			t.Errorf("token should not be in body, got '%v'", receivedBody["token"])
		}

		// Verify response
		if response.Status != "success" {
			t.Errorf("expected status 'success', got '%s'", response.Status)
		}

		if response.Notification.Title != "Deploy Complete" {
			t.Errorf("expected title 'Deploy Complete', got '%s'", response.Notification.Title)
		}

		if response.Notification.Type != "deployment" {
			t.Errorf("expected type 'deployment', got '%s'", response.Notification.Type)
		}

		if len(response.Notification.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(response.Notification.Tags))
		}
	})

	t.Run("successful notifai without type", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)

			w.WriteHeader(200)
			w.Write([]byte(`{
				"status": "success",
				"message": "Notification generated and sent",
				"notification": {
					"title": "CPU Alert",
					"message": "CPU usage at 95% on web-3",
					"type": "alert",
					"tags": ["monitoring", "critical"]
				}
			}`))
		}))
		defer server.Close()

		client := NewClient("abc12345", WithAPIURL(server.URL))

		response, err := client.NotifAI(context.Background(), &NotifAIOptions{
			Text: "cpu at 95% on web-3",
		})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Verify type was not sent
		if _, exists := receivedBody["type"]; exists {
			t.Error("expected type field to be omitted when not provided")
		}

		if response.Notification.Type != "alert" {
			t.Errorf("expected AI-generated type 'alert', got '%s'", response.Notification.Type)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		client := NewClient("abc12345")

		tests := []struct {
			name          string
			options       *NotifAIOptions
			errorContains string
		}{
			{
				name:          "nil options",
				options:       nil,
				errorContains: "options cannot be nil",
			},
			{
				name:          "empty text",
				options:       &NotifAIOptions{Text: ""},
				errorContains: "text is required",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := client.NotifAI(context.Background(), tt.options)

				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if _, ok := err.(*ValidationError); !ok {
					t.Errorf("expected ValidationError, got %T", err)
				}

				if err.Error() == "" || !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			})
		}
	})

	t.Run("HTTP error responses", func(t *testing.T) {
		tests := []struct {
			name           string
			statusCode     int
			responseBody   string
			expectedError  interface{}
			errorSubstring string
		}{
			{
				name:           "400 bad request",
				statusCode:     400,
				responseBody:   `{"status": "error", "error": {"type": "validation_error", "code": "invalid_text", "message": "Invalid text"}}`,
				expectedError:  &ValidationError{},
				errorSubstring: "Invalid text",
			},
			{
				name:           "401 unauthorized",
				statusCode:     401,
				responseBody:   `{"status": "error", "error": {"type": "auth_error", "code": "invalid_token", "message": "Invalid token"}}`,
				expectedError:  &AuthError{},
				errorSubstring: "Invalid token",
			},
			{
				name:           "429 rate limit",
				statusCode:     429,
				responseBody:   `{"status": "error", "error": {"type": "rate_limit_error", "code": "rate_limit_exceeded", "message": "Rate limit exceeded"}}`,
				expectedError:  &RateLimitError{},
				errorSubstring: "Rate limit exceeded",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
					w.Write([]byte(tt.responseBody))
				}))
				defer server.Close()

				// Disable retries for error testing
				client := NewClient("abc12345", WithAPIURL(server.URL), WithMaxRetries(0))

				_, err := client.NotifAI(context.Background(), &NotifAIOptions{
					Text: "test notification",
				})

				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if reflect.TypeOf(err) != reflect.TypeOf(tt.expectedError) {
					t.Errorf("expected error type %T, got %T", tt.expectedError, err)
				}
			})
		}
	})
}
