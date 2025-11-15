package main

import (
	"context"
	"fmt"
	"log"
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

	// Send simple notification
	err := client.SendSimple(context.Background(), "Hello from Go!", "This is a test notification")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Notification sent successfully!")
}
