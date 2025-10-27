package wirepusher

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("with user ID", func(t *testing.T) {
		client := NewClient("", "test-user")

		if client.Token != "" {
			t.Errorf("expected empty token, got '%s'", client.Token)
		}

		if client.UserID != "test-user" {
			t.Errorf("expected userID 'test-user', got '%s'", client.UserID)
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

	t.Run("with token", func(t *testing.T) {
		client := NewClient("wpt_test123", "")

		if client.Token != "wpt_test123" {
			t.Errorf("expected token 'wpt_test123', got '%s'", client.Token)
		}

		if client.UserID != "" {
			t.Errorf("expected empty userID, got '%s'", client.UserID)
		}

		if client.APIURL != DefaultAPIURL {
			t.Errorf("expected APIURL '%s', got '%s'", DefaultAPIURL, client.APIURL)
		}
	})

	t.Run("panics with both token and userID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected NewClient to panic when both token and userID provided")
			}
		}()
		NewClient("wpt_test123", "test-user")
	})

	t.Run("panics with neither token nor userID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected NewClient to panic when neither token nor userID provided")
			}
		}()
		NewClient("", "")
	})

	t.Run("with custom API URL", func(t *testing.T) {
		customURL := "https://custom.example.com/api"
		client := NewClient("", "user", WithAPIURL(customURL))

		if client.APIURL != customURL {
			t.Errorf("expected APIURL '%s', got '%s'", customURL, client.APIURL)
		}
	})

	t.Run("with custom timeout", func(t *testing.T) {
		customTimeout := 5 * time.Second
		client := NewClient("", "user", WithTimeout(customTimeout))

		if client.HTTPClient.Timeout != customTimeout {
			t.Errorf("expected timeout %v, got %v", customTimeout, client.HTTPClient.Timeout)
		}
	})

	t.Run("with custom HTTP client", func(t *testing.T) {
		customClient := &http.Client{Timeout: 10 * time.Second}
		client := NewClient("", "user", WithHTTPClient(customClient))

		if client.HTTPClient != customClient {
			t.Error("expected custom HTTP client to be used")
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
			serverBody:    `{"status": "error", "message": "Title is required"}`,
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

			client := NewClient("", "test-user", WithAPIURL(server.URL))
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

			// Parse body
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)

			w.WriteHeader(200)
			w.Write([]byte(`{"status": "success", "message": "Notification sent"}`))
		}))
		defer server.Close()

		client := NewClient("", "test-user", WithAPIURL(server.URL))

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

		if receivedBody["id"] != "test-user" {
			t.Errorf("expected id 'test-user', got '%v'", receivedBody["id"])
		}

		// Token should be nil/empty since we're using user_id
		if receivedBody["token"] != nil && receivedBody["token"] != "" {
			t.Errorf("expected nil/empty token, got '%v'", receivedBody["token"])
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		client := NewClient("", "test-user")

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
			{
				name:          "empty message",
				options:       &SendOptions{Title: "test", Message: ""},
				errorContains: "message is required",
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
				responseBody:  `{"status": "error", "message": "Invalid request"}`,
				errorType:     &ValidationError{},
				errorContains: "Invalid request",
			},
			{
				name:          "401 unauthorized",
				statusCode:    401,
				responseBody:  `{"status": "error", "message": "Unauthorized"}`,
				errorType:     &AuthError{},
				errorContains: "Unauthorized",
			},
			{
				name:          "403 forbidden",
				statusCode:    403,
				responseBody:  `{"status": "error", "message": "Forbidden"}`,
				errorType:     &AuthError{},
				errorContains: "Forbidden",
			},
			{
				name:          "429 rate limit",
				statusCode:    429,
				responseBody:  `{"status": "error", "message": "Rate limit exceeded"}`,
				errorType:     &RateLimitError{},
				errorContains: "Rate limit exceeded",
			},
			{
				name:          "500 server error",
				statusCode:    500,
				responseBody:  `{"status": "error", "message": "Internal server error"}`,
				errorType:     &Error{},
				errorContains: "Internal server error",
			},
			{
				name:          "non-JSON error response",
				statusCode:    500,
				responseBody:  "Internal Server Error",
				errorType:     &Error{},
				errorContains: "Internal Server Error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
					w.Write([]byte(tt.responseBody))
				}))
				defer server.Close()

				client := NewClient("", "test-user", WithAPIURL(server.URL))

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

		client := NewClient("", "test-user", WithAPIURL(server.URL))

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

		client := NewClient("", "test-user", WithAPIURL(server.URL))

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
		client := NewClient("", "test-user", WithAPIURL("http://localhost:1"))

		err := client.Send(context.Background(), &SendOptions{
			Title:   "Test",
			Message: "Test",
		})

		if err == nil {
			t.Fatal("expected network error, got nil")
		}

		if _, ok := err.(*Error); !ok {
			t.Errorf("expected Error type, got %T", err)
		}
	})

	t.Run("successful send with non-JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		client := NewClient("", "test-user", WithAPIURL(server.URL))

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

func TestErrorTypes(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		err := &Error{Message: "test error", StatusCode: 500}
		expected := "wirepusher: test error (status: 500)"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("Error without status code", func(t *testing.T) {
		err := &Error{Message: "test error", StatusCode: 0}
		expected := "wirepusher: test error"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("AuthError", func(t *testing.T) {
		err := &AuthError{Message: "unauthorized", StatusCode: 401}
		expected := "wirepusher auth error: unauthorized (status: 401)"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := &ValidationError{Message: "invalid input", StatusCode: 400}
		expected := "wirepusher validation error: invalid input (status: 400)"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("RateLimitError", func(t *testing.T) {
		err := &RateLimitError{Message: "rate limit exceeded", StatusCode: 429}
		expected := "wirepusher rate limit error: rate limit exceeded (status: 429)"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})
}
