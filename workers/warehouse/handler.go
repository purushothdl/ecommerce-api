// workers/warehouse/handler.go
package warehouse

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

type WarehouseHandler struct {
	logger        *slog.Logger
	taskCreator   *tasks.TaskCreator
	apiClient     *apiclient.Client 
	processingTime time.Duration
}

// The constructor now accepts the API client
func NewWarehouseHandler(logger *slog.Logger, taskCreator *tasks.TaskCreator, apiClient *apiclient.Client, processingTime time.Duration) *WarehouseHandler {
	return &WarehouseHandler{
		logger:         logger,
		taskCreator:    taskCreator,
		apiClient:      apiClient,
		processingTime: processingTime,
	}
}


func (h *WarehouseHandler) HandleOrderCreated(w http.ResponseWriter, r *http.Request) {
	var event events.OrderCreatedEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.logger.Error("failed to decode order created event", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Received order created event, processing warehouse fulfillment...", "order_id", event.OrderID)

	// Simulate work
	time.Sleep(h.processingTime)
	updatePayload := dto.UpdateOrderStatusRequest{
		Status:         models.OrderStatusProcessing,
	}

	if err := h.apiClient.UpdateOrderStatus(r.Context(), event.OrderID, updatePayload); err != nil {
		h.logger.Error("failed to update api status via client", "order_id", event.OrderID, "error", err)
		http.Error(w, "failed to update api status", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Order packed successfully.", "order_id", event.OrderID)
	
	// FULFILLMENT TASK
	packedEvent := events.OrderPackedEvent{
		OrderID:     event.OrderID,
		OrderNumber: event.OrderNumber, 
		UserID:      event.UserID,
		UserEmail:   event.UserEmail,
		PackedAt:    time.Now(),
	}

	if err := h.taskCreator.CreateFulfillmentTask(r.Context(), "/handle/order-packed", packedEvent); err != nil {
		h.logger.Error("failed to create shipping task", "order_id", event.OrderID, "error", err)
		http.Error(w, "failed to enqueue next task", http.StatusInternalServerError)
		return
	}

	// NOTIFICATION TASK
	notificationEvent := events.NotificationRequestEvent{
		Type:      "ORDER_PACKED", 
		UserEmail: event.UserEmail,
		Payload:   jsonutil.MustMarshal(packedEvent),
	}
	if err := h.taskCreator.CreateFulfillmentTask(r.Context(), "/handle/notification-request", notificationEvent); err != nil {
		h.logger.Error("failed to create packed notification task", "order_id", event.OrderID, "error", err)
		// We don't fail the main task for this. Just log it.
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Warehouse task processed successfully."))
}