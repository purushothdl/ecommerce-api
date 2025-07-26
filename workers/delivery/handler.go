// workers/delivery/handler.go
package delivery

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/purushothdl/ecommerce-api/events"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	"github.com/purushothdl/ecommerce-api/internal/shared/tasks"
	apiclient "github.com/purushothdl/ecommerce-api/pkg/api-client"
	"github.com/purushothdl/ecommerce-api/pkg/utils/jsonutil"
)

type DeliveryHandler struct {
	logger         *slog.Logger
	taskCreator    *tasks.TaskCreator
	apiClient      *apiclient.Client
	processingTime  time.Duration
}

func NewDeliveryHandler(logger *slog.Logger, taskCreator *tasks.TaskCreator, apiClient *apiclient.Client, processingTime time.Duration) *DeliveryHandler {
	return &DeliveryHandler{
		logger:         logger,
		taskCreator:    taskCreator,
		apiClient:      apiClient,
		processingTime: processingTime,
	}
}

func (h *DeliveryHandler) HandleOrderShipped(w http.ResponseWriter, r *http.Request) {
	var event events.OrderShippedEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.logger.Error("failed to decode order shipped event", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Received order shipped event, processing delivery...", "order_id", event.OrderID)

	time.Sleep(h.processingTime)

	updatePayload := dto.UpdateOrderStatusRequest{
		Status:         models.OrderStatusDelivered,
	}


	// Update the main API with the new status
	if err := h.apiClient.UpdateOrderStatus(r.Context(), event.OrderID, updatePayload); err != nil {
		h.logger.Error("failed to update api status to delivered", "order_id", event.OrderID, "error", err)
		http.Error(w, "failed to update api status", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Order delivered successfully.", "order_id", event.OrderID)
	
	// This is the final step in the fulfillment chain, so we only create a notification task.
	deliveredEvent := events.OrderDeliveredEvent{
		OrderID:     event.OrderID,
		OrderNumber: event.OrderNumber,
		UserID:      event.UserID,
		UserEmail:   event.UserEmail,
		DeliveredAt: time.Now(),
	}

	// Create the notification task to email the user
	notificationEvent := events.NotificationRequestEvent{
		Type:      "ORDER_DELIVERED",
		UserEmail: event.UserEmail,
		Payload:   jsonutil.MustMarshal(deliveredEvent),
	}
	if err := h.taskCreator.CreateFulfillmentTask(r.Context(), "/handle/notification-request", notificationEvent); err != nil {
		h.logger.Error("failed to create delivered notification task", "order_id", event.OrderID, "error", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Delivery task processed successfully."))
}

