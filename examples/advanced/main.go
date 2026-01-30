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

	// Send notification with all options
	err := client.Send(context.Background(), &pincho.SendOptions{
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
	err = client.Send(context.Background(), &pincho.SendOptions{
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
