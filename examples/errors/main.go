package main

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/wirepusher/wirepusher-go"
)

func main() {
	// Get token from environment variable
	token := os.Getenv("WIREPUSHER_TOKEN")
	if token == "" {
		token = "abc12345" // Fallback for testing
	}

	// Create client
	client := wirepusher.NewClient(token)

	// Example 1: Handle validation error
	fmt.Println("Example 1: Validation error (empty title)")
	err := client.Send(context.Background(), &wirepusher.SendOptions{
		Title:   "", // Empty title will cause validation error
		Message: "Test message",
	})
	handleError(err)

	// Example 2: Handle validation error (empty message)
	fmt.Println("\nExample 2: Validation error (empty message)")
	err = client.Send(context.Background(), &wirepusher.SendOptions{
		Title:   "Test title",
		Message: "", // Empty message will cause validation error
	})
	handleError(err)

	// Example 3: Handle auth error (invalid token)
	fmt.Println("\nExample 3: Auth error (trying with invalid token)")
	invalidClient := wirepusher.NewClient("invalid-token")
	err = invalidClient.SendSimple(context.Background(), "Test", "This will fail")
	handleError(err)

	// Example 4: Successful send
	fmt.Println("\nExample 4: Successful send")
	err = client.SendSimple(context.Background(), "Success!", "This notification should send successfully")
	if err != nil {
		handleError(err)
	} else {
		fmt.Println("✓ Notification sent successfully!")
	}
}

func handleError(err error) {
	if err == nil {
		fmt.Println("✓ No error")
		return
	}

	// Type switch to handle different error types
	switch e := err.(type) {
	case *wirepusher.ValidationError:
		fmt.Printf("✗ Validation Error: %s (status: %d)\n", e.Message, e.StatusCode)
		fmt.Println("  → Check your input parameters (title, message, etc.)")

	case *wirepusher.AuthError:
		fmt.Printf("✗ Authentication Error: %s (status: %d)\n", e.Message, e.StatusCode)
		fmt.Println("  → Check your token")

	case *wirepusher.RateLimitError:
		fmt.Printf("✗ Rate Limit Error: %s (status: %d)\n", e.Message, e.StatusCode)
		fmt.Println("  → You're sending too many requests. Wait and try again.")

	case *wirepusher.Error:
		fmt.Printf("✗ API Error: %s", e.Message)
		if e.StatusCode > 0 {
			fmt.Printf(" (status: %d)", e.StatusCode)
		}
		fmt.Println()
		fmt.Println("  → Check the error message for details")

	default:
		fmt.Printf("✗ Unknown Error: %v\n", err)
	}
}
