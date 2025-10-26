package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/wirepusher/go-sdk"
)

func main() {
	// Get credentials from environment variables
	token := os.Getenv("WIREPUSHER_TOKEN")
	userID := os.Getenv("WIREPUSHER_USER_ID")

	if token == "" || userID == "" {
		log.Fatal("WIREPUSHER_TOKEN and WIREPUSHER_USER_ID environment variables are required")
	}

	// Create client
	client := wirepusher.NewClient(token, userID)

	// Example 1: With timeout
	fmt.Println("Example 1: Sending with 5-second timeout...")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()

	err := client.SendSimple(ctx1, "Test with Timeout", "This request has a 5-second timeout")
	if err != nil {
		if ctx1.Err() == context.DeadlineExceeded {
			log.Fatal("Request timed out after 5 seconds")
		} else {
			log.Fatal("Request failed:", err)
		}
	}
	fmt.Println("✓ Sent successfully with timeout")

	// Example 2: With cancellation
	fmt.Println("\nExample 2: Sending with cancellable context...")
	ctx2, cancel2 := context.WithCancel(context.Background())

	// Cancel after 2 seconds (for demonstration)
	go func() {
		time.Sleep(100 * time.Millisecond) // Give it time to send
		fmt.Println("  (Cancellation is ready but request should complete first)")
	}()

	err = client.SendSimple(ctx2, "Test with Cancellation", "This request can be cancelled")
	if err != nil {
		if ctx2.Err() == context.Canceled {
			log.Fatal("Request was canceled")
		} else {
			log.Fatal("Request failed:", err)
		}
	}
	cancel2() // Clean up
	fmt.Println("✓ Sent successfully (completed before cancellation)")

	// Example 3: Custom timeout on client
	fmt.Println("\nExample 3: Using client with custom timeout...")
	customClient := wirepusher.NewClient(
		token,
		userID,
		wirepusher.WithTimeout(10*time.Second),
	)

	err = customClient.SendSimple(context.Background(), "Test with Custom Client Timeout", "Client configured with 10-second timeout")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Sent successfully with custom client timeout")

	fmt.Println("\nAll examples completed successfully!")
}
