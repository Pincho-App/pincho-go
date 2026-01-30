package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gitlab.com/pincho-app/pincho-go"
)

func main() {
	// Get token from environment variable
	token := os.Getenv("PINCHO_TOKEN")
	if token == "" {
		token = "abc12345" // Fallback for testing
	}

	// Create client
	client := pincho.NewClient(token)

	// Send simple notification
	err := client.SendSimple(context.Background(), "Hello from Go!", "This is a test notification")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Notification sent successfully!")
}
