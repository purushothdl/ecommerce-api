// workers/shipping/handler.go
package shipping

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
	"github.com/purushothdl/ecommerce-api/pkg/utils/orders"
	"github.com/purushothdl/ecommerce-api/pkg/utils/timeutil"
)

type ShippingHandler struct {
	logger         *slog.Logger
	taskCreator    *tasks.TaskCreator
	apiClient      *apiclient.Client
	processingTime  time.Duration
}

func NewShippingHandler(logger *slog.Logger, taskCreator *tasks.TaskCreator, apiClient *apiclient.Client, processingTime time.Duration) *ShippingHandler {
	return &ShippingHandler{
		logger:         logger,
		taskCreator:    taskCreator,
		apiClient:      apiClient,
		processingTime: processingTime,
	}
}

func (h *ShippingHandler) HandleOrderPacked(w http.ResponseWriter, r *http.Request) {
	var event events.OrderPackedEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.logger.Error("failed to decode order packed event", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Received order packed event, processing shipping...", "order_id", event.OrderID)

	time.Sleep(h.processingTime)
	trackingNumber := orders.GenerateTrackingID()
	estimatedDeliveryDate := timeutil.CalculateEDD(time.Now(), 2)

	updatePayload := dto.UpdateOrderStatusRequest{
		Status:         models.OrderStatusShipped,
		TrackingNumber: &trackingNumber, 
		EstimatedDeliveryDate: &estimatedDeliveryDate,
	}

	// Update the main API with the new status
	if err := h.apiClient.UpdateOrderStatus(r.Context(), event.OrderID, updatePayload); err != nil {
		h.logger.Error("failed to update api status to shipped", "order_id", event.OrderID, "error", err)
		http.Error(w, "failed to update api status", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Order shipped successfully.", "order_id", event.OrderID, "edd", estimatedDeliveryDate.Format("2006-01-02"))
	
	// Create the event for the next step (delivery)
	shippedEvent := events.OrderShippedEvent{
		OrderID:               event.OrderID,
		OrderNumber:           event.OrderNumber,
		UserID:                event.UserID,
		UserEmail:             event.UserEmail,
		TrackingNumber:        trackingNumber,
		ShippedAt:             time.Now(),
		EstimatedDeliveryDate: estimatedDeliveryDate,
	}

	// Create the fulfillment task for the delivery handler
	if err := h.taskCreator.CreateFulfillmentTask(r.Context(), "/handle/order-shipped", shippedEvent); err != nil {
		h.logger.Error("failed to create delivery task", "order_id", event.OrderID, "error", err)
		http.Error(w, "failed to enqueue next task", http.StatusInternalServerError)
		return
	}

	// Create the notification task to email the user
	notificationEvent := events.NotificationRequestEvent{
		Type:      "ORDER_SHIPPED",
		UserEmail: event.UserEmail,
		Payload:   jsonutil.MustMarshal(shippedEvent),
	}
	if err := h.taskCreator.CreateFulfillmentTask(r.Context(), "/handle/notification-request", notificationEvent); err != nil {
		h.logger.Error("failed to create shipped notification task", "order_id", event.OrderID, "error", err)
		// Don't fail the main task for this, just log it.
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Shipping task processed successfully."))
}

