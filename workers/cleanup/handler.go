// workers/cleanup/handler.go
package cleanup

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/purushothdl/ecommerce-api/internal/domain"
)

// CleanupHandler handles scheduled cleanup operations for the application.
type CleanupHandler struct {
	logger                       *slog.Logger
	orderService                 domain.OrderService
	cartService                  domain.CartService
	pendingOrderCleanupThreshold time.Duration
	anonymousCartCleanupThreshold time.Duration
}

// Update the constructor to accept the CartService
func NewCleanupHandler(
	logger *slog.Logger,
	orderService domain.OrderService,
	cartService domain.CartService,
	pendingOrderCleanupThreshold time.Duration,
	anonymousCartCleanupThreshold time.Duration,
) *CleanupHandler {
	return &CleanupHandler{
		logger:                       logger,
		orderService:                 orderService,
		cartService:                  cartService,
		pendingOrderCleanupThreshold: pendingOrderCleanupThreshold,
		anonymousCartCleanupThreshold: anonymousCartCleanupThreshold,
	}
}

// HandleCleanupPendingOrders is the HTTP endpoint for the scheduled order cleanup job.
func (h *CleanupHandler) HandleCleanupPendingOrders(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Received scheduled request to clean up pending orders.")

	cleanedCount, err := h.orderService.CleanupPendingOrders(r.Context(), h.pendingOrderCleanupThreshold)
	if err != nil {
		h.logger.Error("Scheduled job failed to clean up pending orders", "error", err)
		http.Error(w, "Failed to clean up pending orders", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Pending order cleanup job completed successfully.", "cleaned_count", cleanedCount)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pending order cleanup job finished."))
}

// HandleCleanupAnonymousCarts is the HTTP endpoint for the scheduled cart cleanup job.
func (h *CleanupHandler) HandleCleanupAnonymousCarts(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Received scheduled request to clean up anonymous carts.")

	// Use the configured duration for cart cleanup (e.g., 30 days).
	cleanedCount, err := h.cartService.CleanupOldAnonymousCarts(r.Context(), h.anonymousCartCleanupThreshold)
	if err != nil {
		h.logger.Error("Scheduled job failed to clean up anonymous carts", "error", err)
		http.Error(w, "Failed to clean up anonymous carts", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Anonymous cart cleanup job completed successfully.", "cleaned_cart_count", cleanedCount)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Anonymous cart cleanup job finished."))
}

func (h *CleanupHandler) HandleScheduledMaintenance(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("--- Received request to run all scheduled maintenance tasks ---")

	// --- Run Order Cleanup ---
	orderCleanedCount, orderErr := h.orderService.CleanupPendingOrders(r.Context(), h.pendingOrderCleanupThreshold)
	if orderErr != nil {
		h.logger.Error("Maintenance sub-task failed: CleanupPendingOrders", "error", orderErr)
		// Log the error but don't stop. We still want to try the cart cleanup.
	} else {
		h.logger.Info("Maintenance sub-task successful: CleanupPendingOrders", "cleaned_count", orderCleanedCount)
	}

	// --- Run Cart Cleanup ---
	cartCleanedCount, cartErr := h.cartService.CleanupOldAnonymousCarts(r.Context(), h.anonymousCartCleanupThreshold)
	if cartErr != nil {
		h.logger.Error("Maintenance sub-task failed: CleanupOldAnonymousCarts", "error", cartErr)
	} else {
		h.logger.Info("Maintenance sub-task successful: CleanupOldAnonymousCarts", "cleaned_cart_count", cartCleanedCount)
	}
	
	h.logger.Info("--- All scheduled maintenance tasks have been run ---")
	
	// Always return a 200 OK so Cloud Scheduler doesn't retry unless there's a total crash.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scheduled maintenance finished."))
}