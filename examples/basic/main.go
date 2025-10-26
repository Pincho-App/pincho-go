package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	// Send simple notification
	err := client.SendSimple(context.Background(), "Hello from Go!", "This is a test notification")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Notification sent successfully!")
}
