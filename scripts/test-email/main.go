// scripts/test-email/main.go
package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/purushothdl/ecommerce-api/workers/notification"
)

func main() {
	log.Println("--- Starting Email Service Test Script ---")

	err := godotenv.Load("worker.env")
	if err != nil {
		log.Fatalf("Error loading worker.env file. Make sure it exists at the project root. Error: %v", err)
	}

	// Get required configuration from environment variables
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	fromName := os.Getenv("RESEND_FROM_NAME")
	testRecipient := os.Getenv("TEST_EMAIL_RECIPIENT")

	// Validate that the required configuration is present
	if apiKey == "" || fromEmail == "" || testRecipient == "" {
		log.Fatalf("FATAL: Please make sure RESEND_API_KEY, RESEND_FROM_EMAIL, and TEST_EMAIL_RECIPIENT are set in your worker.env file.")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	log.Printf("Attempting to send email from '%s <%s>' to '%s'", fromName, fromEmail, testRecipient)

	// Instantiate the EmailService using the exact same code as our real worker
	emailService := notification.NewEmailService(apiKey, fromEmail, fromName, logger)

	subject := "Test Email from GoKart E-Commerce"
	htmlBody := "<h1>Hello!</h1><p>This is a test message from your Go e-commerce application's EmailService. If you are seeing this, it means your Resend configuration and code are working correctly!</p>"

	messageID, err := emailService.SendEmail(testRecipient, subject, htmlBody)
	if err != nil {
		log.Fatalf("--- TEST FAILED --- \nFailed to send email: %v", err)
	}

	log.Printf("--- TEST SUCCESSFUL --- \nEmail sent successfully! \nResend Message ID: %s", messageID)
}