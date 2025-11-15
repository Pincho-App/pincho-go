// Package main demonstrates encrypted notifications with WirePusher Go SDK.
//
// This example shows how to send encrypted notifications where only the message
// content is encrypted. Title, type, tags, and other metadata remain unencrypted
// for filtering and display purposes.
//
// Requirements:
//   - WirePusher app installed with configured notification type
//   - Encryption password set for the notification type in app settings
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gitlab.com/wirepusher/wirepusher-go"
)

func main() {
	// Get credentials from environment variables (recommended)
	token := os.Getenv("WIREPUSHER_TOKEN")
	encryptionPassword := os.Getenv("WIREPUSHER_ENCRYPTION_PASSWORD")

	if token == "" {
		token = "abc12345" // Fallback for testing
	}

	if encryptionPassword == "" {
		log.Fatal("Error: WIREPUSHER_ENCRYPTION_PASSWORD environment variable not set\n" +
			"Set it with: export WIREPUSHER_ENCRYPTION_PASSWORD='your_password'")
	}

	// Create client
	client := wirepusher.NewClient(token)

	ctx := context.Background()

	// Example 1: Basic encrypted notification (personal)
	fmt.Println("Example 1: Sending encrypted notification to personal device...")
	err := client.Send(ctx, &wirepusher.SendOptions{
		Title:              "Secure Alert",                        // NOT encrypted (for display)
		Message:            "Your credit card was charged $49.99", // Encrypted
		Type:               "secure",                              // NOT encrypted (needed for password lookup)
		EncryptionPassword: encryptionPassword,                    // Must match app configuration
	})
	if err != nil {
		log.Fatalf("  Error: %v\n", err)
	}
	fmt.Println("  ✓ Sent successfully")

	// Example 2: Encrypted notification with all optional parameters
	fmt.Println("Example 2: Encrypted notification with tags and metadata...")
	err = client.Send(ctx, &wirepusher.SendOptions{
		Title:              "Security Alert",
		Message:            "Unauthorized login attempt detected from IP 192.168.1.100",
		Type:               "security",
		Tags:               []string{"critical", "security", "login"}, // NOT encrypted
		EncryptionPassword: encryptionPassword,
	})
	if err != nil {
		log.Fatalf("  Error: %v\n", err)
	}
	fmt.Println("  ✓ Sent successfully")

	// Example 3: Encrypted notification with image and action URL
	fmt.Println("Example 3: Encrypted notification with additional metadata...")
	err = client.Send(ctx, &wirepusher.SendOptions{
		Title:              "Payment Alert",
		Message:            "Your subscription has been renewed for $9.99",
		Type:               "billing",
		Tags:               []string{"payment", "subscription"},
		EncryptionPassword: encryptionPassword,
		ImageURL:           "https://example.com/payment-icon.png", // NOT encrypted
		ActionURL:          "https://example.com/billing/history",  // NOT encrypted
	})
	if err != nil {
		log.Fatalf("  Error: %v\n", err)
	}
	fmt.Println("  ✓ Sent successfully")

	// Example 4: Error handling with encryption
	fmt.Println("Example 4: Error handling...")

	// Unencrypted notification (no password provided)
	err = client.Send(ctx, &wirepusher.SendOptions{
		Title:   "Unencrypted Message",
		Message: "This message is sent in plain text",
		Type:    "info",
		// No EncryptionPassword provided
	})
	if err != nil {
		log.Fatalf("  Error: %v\n", err)
	}
	fmt.Println("  ✓ Unencrypted notification sent successfully")

	// Encrypted notification (with password)
	err = client.Send(ctx, &wirepusher.SendOptions{
		Title:              "Encrypted Message",
		Message:            "This message is encrypted",
		Type:               "secure",
		EncryptionPassword: encryptionPassword,
	})
	if err != nil {
		log.Fatalf("  Error: %v\n", err)
	}
	fmt.Println("  ✓ Encrypted notification sent successfully")

	fmt.Println("✓ All examples completed!")
	fmt.Println("\nKey points:")
	fmt.Println("  - Only the 'message' field is encrypted")
	fmt.Println("  - Title, type, tags, imageURL, actionURL remain unencrypted")
	fmt.Println("  - Encryption password must match type configuration in app")
	fmt.Println("  - Each encrypted message gets a unique random IV")
	fmt.Println("  - Unencrypted messages work normally (backward compatible)")
	fmt.Println("  - Uses only Go standard library (crypto/aes, crypto/cipher, crypto/sha1)")
}
