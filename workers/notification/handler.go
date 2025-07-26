// workers/notification/handler.go (New File)
package notification

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/purushothdl/ecommerce-api/events"
)

type NotificationHandler struct {
	logger          *slog.Logger
	emailService    *EmailService
	templateService *TemplateService
}


func NewNotificationHandler(logger *slog.Logger, emailService *EmailService, templateService *TemplateService) *NotificationHandler {
	return &NotificationHandler{
		logger:          logger,
		emailService:    emailService,
		templateService: templateService,
	}
}

// HandleNotificationRequest is the single entrypoint for all notification tasks.
func (h *NotificationHandler) HandleNotificationRequest(w http.ResponseWriter, r *http.Request) {
	var event events.NotificationRequestEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.logger.Error("failed to decode notification request event", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Received notification request", "type", event.Type, "email", event.UserEmail)

	var subject, body string
	var err error

	// This acts as a router for different notification types
	switch event.Type {
	case "ORDER_CONFIRMED":
		var payload events.OrderCreatedEvent
		if err = json.Unmarshal(event.Payload, &payload); err == nil {
			subject, body, err = h.templateService.GenerateOrderConfirmedEmail(payload)
		}

	case "ORDER_PACKED":
		var payload events.OrderPackedEvent
		if err = json.Unmarshal(event.Payload, &payload); err == nil {
			subject, body, err = h.templateService.GenerateOrderPackedEmail(payload)
		}

	case "ORDER_SHIPPED":
		var payload events.OrderShippedEvent
		if err = json.Unmarshal(event.Payload, &payload); err == nil {
			subject, body, err = h.templateService.GenerateOrderShippedEmail(payload)
		}

	case "ORDER_DELIVERED": 
		var payload events.OrderDeliveredEvent
		if err = json.Unmarshal(event.Payload, &payload); err == nil {
			subject, body, err = h.templateService.GenerateOrderDeliveredEmail(payload)
		}

	default:
		err = fmt.Errorf("unhandled notification type: %s", event.Type)
	}	

	if err != nil {
		h.logger.Error("failed to process notification payload", "type", event.Type, "error", err)
		http.Error(w, "failed to process payload", http.StatusBadRequest)
		return
	}

	// Send the generated email
	if _, err := h.emailService.SendEmail(event.UserEmail, subject, body); err != nil {
		http.Error(w, "failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Notification processed successfully."))
}


