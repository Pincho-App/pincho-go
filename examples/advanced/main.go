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

	// Send notification with all options
	err := client.Send(context.Background(), &wirepusher.SendOptions{
		Title:     "Deployment Complete",
		Message:   "Version 2.1.0 has been successfully deployed to production",
		Type:      "deployment",
		Tags:      []string{"production", "backend", "api"},
		ImageURL:  "https://example.com/success-icon.png",
		ActionURL: "https://dashboard.example.com/deployments/123",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Advanced notification sent successfully!")

	// Send another notification with different type
	err = client.Send(context.Background(), &wirepusher.SendOptions{
		Title:   "Server Alert",
		Message: "CPU usage is above 90% on web-server-01",
		Type:    "alert",
		Tags:    []string{"monitoring", "critical"},
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Alert notification sent successfully!")
}
