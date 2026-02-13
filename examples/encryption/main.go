// Package main demonstrates encrypted notifications with Pincho Go SDK.
//
// This example shows how to send encrypted notifications where sensitive content
// is encrypted. The encrypted fields are: title, message, imageURL, actionURL.
// Type and tags remain unencrypted for filtering and routing purposes.
//
// Requirements:
//   - Pincho app installed with configured notification type
//   - Encryption password set for the notification type in app settings
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Pincho-App/pincho-go"
)

func main() {
	// Get credentials from environment variables (recommended)
	token := os.Getenv("PINCHO_TOKEN")
	encryptionPassword := os.Getenv("PINCHO_ENCRYPTION_PASSWORD")

	if token == "" {
		token = "abc12345" // Fallback for testing
	}

	if encryptionPassword == "" {
		log.Fatal("Error: PINCHO_ENCRYPTION_PASSWORD environment variable not set\n" +
			"Set it with: export PINCHO_ENCRYPTION_PASSWORD='your_password'")
	}

	// Create client
	client := pincho.NewClient(token)

	ctx := context.Background()

	// Example 1: Basic encrypted notification
	fmt.Println("Example 1: Sending encrypted notification...")
	err := client.Send(ctx, &pincho.SendOptions{
		Title:              "Secure Alert",                        // Encrypted
		Message:            "Your credit card was charged $49.99", // Encrypted
		Type:               "secure",                              // NOT encrypted (needed for password lookup)
		EncryptionPassword: encryptionPassword,                    // Must match app configuration
	})
	if err != nil {
		log.Fatalf("  Error: %v\n", err)
	}
	fmt.Println("  ✓ Sent successfully")

	// Example 2: Encrypted notification with all optional parameters
	fmt.Println("Example 2: Encrypted notification with tags...")
	err = client.Send(ctx, &pincho.SendOptions{
		Title:              "Security Alert",                                            // Encrypted
		Message:            "Unauthorized login attempt detected from IP 192.168.1.100", // Encrypted
		Type:               "security",                                                  // NOT encrypted (for filtering)
		Tags:               []string{"critical", "security", "login"},                   // NOT encrypted (for filtering)
		EncryptionPassword: encryptionPassword,
	})
	if err != nil {
		log.Fatalf("  Error: %v\n", err)
	}
	fmt.Println("  ✓ Sent successfully")

	// Example 3: Encrypted notification with image and action URL
	fmt.Println("Example 3: Encrypted notification with URLs...")
	err = client.Send(ctx, &pincho.SendOptions{
		Title:              "Payment Alert",                                // Encrypted
		Message:            "Your subscription has been renewed for $9.99", // Encrypted
		Type:               "billing",                                      // NOT encrypted
		Tags:               []string{"payment", "subscription"},            // NOT encrypted
		EncryptionPassword: encryptionPassword,
		ImageURL:           "https://example.com/payment-icon.png", // Encrypted
		ActionURL:          "https://example.com/billing/history",  // Encrypted
	})
	if err != nil {
		log.Fatalf("  Error: %v\n", err)
	}
	fmt.Println("  ✓ Sent successfully")

	// Example 4: Mixed encrypted and unencrypted notifications
	fmt.Println("Example 4: Comparing encrypted vs unencrypted...")

	// Unencrypted notification (no password provided)
	err = client.Send(ctx, &pincho.SendOptions{
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
	err = client.Send(ctx, &pincho.SendOptions{
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
	fmt.Println("  - Encrypted fields: title, message, imageURL, actionURL")
	fmt.Println("  - NOT encrypted: type, tags (needed for filtering/routing)")
	fmt.Println("  - Encryption password must match type configuration in app")
	fmt.Println("  - Each encrypted notification gets a unique random IV")
	fmt.Println("  - Unencrypted notifications work normally (backward compatible)")
	fmt.Println("  - Uses only Go standard library (crypto/aes, crypto/cipher, crypto/sha1)")
}
