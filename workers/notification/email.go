// workers/notification/email.go 
package notification

import (
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client *resend.Client
	fromEmail string // e.g., "onboarding@resend.dev"
	fromName  string // e.g., "GoKart Shopping"
	logger *slog.Logger
}

func NewEmailService(apiKey string, fromEmail string, fromName string, logger *slog.Logger) *EmailService {
	client := resend.NewClient(apiKey)
	return &EmailService{
		client:    client,
		fromEmail: fromEmail,
		fromName:  fromName,
		logger:    logger,
	}
}


// SendEmail is a generic method to send an email.
func (s *EmailService) SendEmail(to, subject, htmlBody string) (string, error) {
	fromHeader := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)

	params := &resend.SendEmailRequest{
		From:    fromHeader, 
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		s.logger.Error("Failed to send email via Resend", "error", err)
		return "", fmt.Errorf("resend: failed to send email: %w", err)
	}

	s.logger.Info("Email sent successfully", "message_id", sent.Id)
	return sent.Id, nil
}