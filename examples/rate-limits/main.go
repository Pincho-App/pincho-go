package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	ctx := context.Background()

	// Example 1: Monitor rate limits after each request
	fmt.Println("Example 1: Monitoring rate limits after requests")
	fmt.Println("---------------------------------------------------")

	for i := 1; i <= 3; i++ {
		fmt.Printf("Sending notification %d...\n", i)

		err := client.SendSimple(ctx, "Rate Limit Test", fmt.Sprintf("Test notification %d", i))
		if err != nil {
			log.Printf("  Error: %v\n", err)
			continue
		}

		// Check rate limit information after successful request
		if info := client.LastRateLimit; info != nil {
			fmt.Printf("  Request %d successful!\n", i)
			fmt.Printf("  Rate Limit Info:\n")
			fmt.Printf("    - Limit: %d requests per window\n", info.Limit)
			fmt.Printf("    - Remaining: %d requests\n", info.Remaining)
			if !info.Reset.IsZero() {
				fmt.Printf("    - Resets at: %s\n", info.Reset.Format(time.RFC3339))
				fmt.Printf("    - Time until reset: %s\n", time.Until(info.Reset).Round(time.Second))
			}
		} else {
			fmt.Printf("  Request %d successful (no rate limit headers in response)\n", i)
		}
		fmt.Println()

		// Small delay between requests to be nice to the API
		time.Sleep(500 * time.Millisecond)
	}

	// Example 2: Proactive rate limit checking
	fmt.Println("Example 2: Proactive rate limit checking")
	fmt.Println("------------------------------------------")

	err := client.SendSimple(ctx, "Final Test", "Checking remaining quota")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if info := client.LastRateLimit; info != nil {
		fmt.Printf("Current quota: %d/%d\n\n", info.Remaining, info.Limit)

		// Implement proactive rate limiting
		if info.Remaining == 0 {
			waitTime := time.Until(info.Reset)
			fmt.Printf("WARNING: Rate limit exhausted! Wait %s before next request\n", waitTime.Round(time.Second))
		} else if info.Remaining < 10 {
			fmt.Printf("WARNING: Low quota! Only %d requests remaining\n", info.Remaining)
			fmt.Println("Consider throttling or batching requests")
		} else if info.Remaining < info.Limit/2 {
			fmt.Printf("INFO: Using quota - %d%% remaining\n", (info.Remaining*100)/info.Limit)
		} else {
			fmt.Printf("OK: Plenty of quota available (%d requests)\n", info.Remaining)
		}
	}

	// Example 3: Using rate limit info for scheduling
	fmt.Println("\nExample 3: Smart scheduling based on rate limits")
	fmt.Println("---------------------------------------------------")

	if info := client.LastRateLimit; info != nil && info.Limit > 0 {
		// Calculate safe rate for continuous sending
		windowDuration := time.Until(info.Reset)
		if windowDuration > 0 && info.Remaining > 0 {
			safeInterval := windowDuration / time.Duration(info.Remaining)
			fmt.Printf("For %d remaining requests in %s:\n", info.Remaining, windowDuration.Round(time.Second))
			fmt.Printf("  Safe sending interval: %s per notification\n", safeInterval.Round(time.Millisecond))
			fmt.Printf("  Max sustainable rate: %.2f notifications/minute\n", float64(info.Remaining)/windowDuration.Minutes())
		}
	}

	fmt.Println("\nAll examples completed!")
}
