package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/events"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	"github.com/purushothdl/ecommerce-api/internal/shared/tasks"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/utils/jsonutil"
	"github.com/purushothdl/ecommerce-api/pkg/utils/orders"
	"github.com/purushothdl/ecommerce-api/pkg/utils/timeutil"
)

type orderService struct {
	store          domain.Store
	paymentService domain.PaymentService 
	taskCreator    *tasks.TaskCreator
	logger         *slog.Logger
	config         configs.OrderFinancialsConfig
}

// NewOrderService creates a new OrderService
func NewOrderService(store domain.Store, paymentService domain.PaymentService, taskCreator *tasks.TaskCreator, logger *slog.Logger, config configs.OrderFinancialsConfig) domain.OrderService {
	return &orderService{
		store:          store,
		paymentService: paymentService,
		taskCreator:    taskCreator, 
		logger:         logger,
		config:         config,
	}
}


// CreateOrder handles the entire process of creating an order.
func (s *orderService) CreateOrder(ctx context.Context, userID int64, cartID int64, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
	var response *dto.CreateOrderResponse

	err := s.store.ExecTx(ctx, func(q *domain.Queries) error {
		// 1. Get cart items from the user's cart in context.
		cartItems, err := q.CartRepo.GetItemsByCartID(ctx, cartID)
		if err != nil {
			s.logger.Error("failed to get cart items for order creation", "cart_id", cartID, "error", err)
			return fmt.Errorf("could not retrieve cart for order: %w", err)
		}
		if len(cartItems) == 0 {
			return errors.New("cannot create an order from an empty cart")
		}

		// 2. Fetch and validate addresses.
		shippingAddr, err := q.AddressRepo.GetByID(ctx, req.ShippingAddressID)
		if err != nil {
			return fmt.Errorf("shipping address not found: %w", err)
		}
		if shippingAddr.UserID != userID {
			return apperrors.ErrUnauthorized 
		}

		billingAddr, err := q.AddressRepo.GetByID(ctx, req.BillingAddressID)
		if err != nil {
			return fmt.Errorf("billing address not found: %w", err)
		}
		if billingAddr.UserID != userID {
			return apperrors.ErrUnauthorized 
		}
		
		// 3. Lock products, validate stock, and calculate totals.
		var subtotal float64
		productSnapshots := make(map[int64]*models.Product)

		for _, item := range cartItems {
			product, err := q.ProductRepo.GetByIDForUpdate(ctx, item.Product.ID)
			if err != nil {
				return fmt.Errorf("product with ID %d not found: %w", item.Product.ID, err)
			}
			if product.StockQuantity < item.Quantity {
				return fmt.Errorf("insufficient stock for %s. available: %d, requested: %d", product.Name, product.StockQuantity, item.Quantity)
			}
			subtotal += product.Price * float64(item.Quantity)
			productSnapshots[item.Product.ID] = product
		}

		// Calculate tax, shipping, and discount amounts
		taxAmount := subtotal * s.config.OrderTaxRate
		shippingCost := s.config.OrderShippingCost
		discountAmount := s.config.OrderDiscountAmount
        
        // Calculate total amount
        totalAmount := max(subtotal + taxAmount + shippingCost - discountAmount, 0)

		// 4. Create Stripe Payment Intent.
		stripePI, err := s.paymentService.CreatePaymentIntent(ctx, totalAmount)
		if err != nil {
			s.logger.Error("failed to create stripe payment intent", "error", err)
			return fmt.Errorf("payment provider error: %w", err)
		}

		// 5. Create the main Order record.
        // Marshal address structs to JSONB
        shippingJSON, _ := json.Marshal(orders.ToOrderAddress(shippingAddr))
        billingJSON, _ := json.Marshal(orders.ToOrderAddress(billingAddr))

		defaultEDD := timeutil.CalculateEDD(time.Now(), 4)
		
		order := &models.Order{
			UserID:                userID,
			OrderNumber:           orders.Generate(),
			Status:                models.OrderStatusPendingPayment,
			PaymentStatus:         models.PaymentStatusPending,
			PaymentMethod:         req.PaymentMethod,
			PaymentIntentID:       stripePI.ID,
			Subtotal:              subtotal,
			TaxAmount:             taxAmount,
			ShippingCost:          shippingCost,
			DiscountAmount:        discountAmount,
			TotalAmount:           totalAmount,
			ShippingAddress:       json.RawMessage(shippingJSON),
			BillingAddress:        json.RawMessage(billingJSON),
			EstimatedDeliveryDate: defaultEDD,
		}
		if err := q.OrderRepo.Create(ctx, order); err != nil {
			s.logger.Error("failed to save order", "error", err)
			return fmt.Errorf("could not save order: %w", err)
		}

		// 6. Create Order Items and update stock.
		var orderItemsToCreate []*models.OrderItem
		for _, item := range cartItems {
			product := productSnapshots[item.Product.ID]
			orderItem := &models.OrderItem{
				OrderID:     order.ID,
				ProductID:   product.ID,
				ProductName: product.Name,
				ProductSKU:  product.SKU,
				UnitPrice:   product.Price,
				Quantity:    item.Quantity,
				TotalPrice:  product.Price * float64(item.Quantity),
			}
            orderItemsToCreate = append(orderItemsToCreate, orderItem)

			// Decrement stock
			if err := q.ProductRepo.UpdateStock(ctx, product.ID, -item.Quantity); err != nil {
				return fmt.Errorf("failed to update stock for product %d: %w", product.ID, err)
			}
		}

        if err := q.OrderRepo.CreateItems(ctx, orderItemsToCreate); err != nil {
            s.logger.Error("failed to save order items", "error", err)
            return fmt.Errorf("could not save order items: %w", err)
        }

		// 7. Clear the cart.
		if err := q.CartRepo.ClearCart(ctx, cartID); err != nil {
			return fmt.Errorf("failed to clear cart: %w", err)
		}

		// 8. Set the response object to be returned by the outer function.
		response = &dto.CreateOrderResponse{
			OrderID:      order.ID,
			OrderNumber:  order.OrderNumber,
			ClientSecret: stripePI.ClientSecret,
		}
		return nil
	})

	return response, err
}


func (s *orderService) HandlePaymentSucceeded(ctx context.Context, paymentIntentID string) error {
	var order *models.Order
	var user *models.User
	var orderItems []*models.OrderItem

	// The transaction ensures we only create the task if the DB update succeeds.
	err := s.store.ExecTx(ctx, func(q *domain.Queries) error {
		var txErr error
		order, txErr = q.OrderRepo.GetByPaymentIntentID(ctx, paymentIntentID)
		if txErr != nil {
			s.logger.Error("webhook cannot find order for payment intent", "pi_id", paymentIntentID)
			return txErr
		}

		if order.PaymentStatus == models.PaymentStatusPaid {
			s.logger.Info("webhook received for already-paid order, ignoring", "order_id", order.ID)
			return nil
		}

		// Fetch the user to get their email for the notification.
		user, txErr = q.UserRepo.GetByID(ctx, order.UserID)
		if txErr != nil {
			s.logger.Error("failed to get user for task creation", "user_id", order.UserID)
			return txErr
		}

		// Fetch the order items associated with this order.
		orderItems, txErr = q.OrderRepo.GetItemsByOrderID(ctx, order.ID)
		if txErr != nil {
			s.logger.Error("failed to get order items for event", "order_id", order.ID)
			return txErr
		}

		s.logger.Info("updating order status to confirmed/paid", "order_id", order.ID, "pi_id", paymentIntentID)
		// We pass nil for tracking and EDD as they are not available yet.
		return q.OrderRepo.UpdateStatus(ctx, order.ID, models.OrderStatusConfirmed, models.PaymentStatusPaid, nil, nil)
	})

	if err != nil {
		return err
	}

	// This happens *after* the transaction has successfully committed.
	if order != nil && user != nil {

		// Map database order items to the event's OrderItemInfo struct.
		// This now works because 'orderItems' was populated inside the transaction.
		eventItems := make([]events.OrderItemInfo, len(orderItems))
		for i, item := range orderItems {
			eventItems[i] = events.OrderItemInfo{
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
			}
		}

		// FULFILLMENT TASK
		fulfillmentEvent := events.OrderCreatedEvent{
			OrderID:         order.ID,
			OrderNumber:     order.OrderNumber,
			UserID:          order.UserID,
			UserEmail:       user.Email,
			TotalAmount:     order.TotalAmount,
			OrderDate:       order.CreatedAt,
			Items:           eventItems,
		}

		if err := s.taskCreator.CreateFulfillmentTask(ctx, "/handle/order-created", fulfillmentEvent); err != nil {
			s.logger.Error("CRITICAL: failed to enqueue order fulfillment task", "order_id", order.ID, "error", err)
		}

		// NOTIFICATION TASK
		notificationEvent := events.NotificationRequestEvent{
			Type:      "ORDER_CONFIRMED",
			UserEmail: user.Email,
			Payload:   jsonutil.MustMarshal(fulfillmentEvent),
		}
		if err := s.taskCreator.CreateFulfillmentTask(ctx, "/handle/notification-request", notificationEvent); err != nil {
			s.logger.Error("CRITICAL: failed to enqueue order confirmed notification task", "order_id", order.ID, "error", err)
		}
	}

	return nil
}

// ListUserOrders retrieves all orders for a given user.
func (s *orderService) ListUserOrders(ctx context.Context, userID int64) ([]*dto.OrderResponse, error) {
	var orders []*models.Order
	var orderDTOs []*dto.OrderResponse

	err := s.store.ExecTx(ctx, func(q *domain.Queries) error {
		var txErr error
		orders, txErr = q.OrderRepo.GetByUserID(ctx, userID)
		return txErr
	})

	if err != nil {
		s.logger.Error("failed to list user orders", "user_id", userID, "error", err)
		return nil, err
	}

	// Map the database models to the response DTOs
	for _, order := range orders {
		orderDTOs = append(orderDTOs, &dto.OrderResponse{
			ID:            order.ID,
			OrderNumber:   order.OrderNumber,
			Status:        order.Status,
			PaymentStatus: order.PaymentStatus,
			TotalAmount:   order.TotalAmount,
			CreatedAt:     order.CreatedAt,
		})
	}

	return orderDTOs, nil
}

// GetUserOrder retrieves a single detailed order for a user.
func (s *orderService) GetUserOrder(ctx context.Context, userID, orderID int64) (*dto.OrderWithItemsResponse, error) {
	var order *models.Order
	var items []*models.OrderItem

	err := s.store.ExecTx(ctx, func(q *domain.Queries) error {
		var txErr error
		// The repo method enforces that the userID matches the order.
		order, txErr = q.OrderRepo.GetByID(ctx, orderID, userID)
		if txErr != nil {
			return txErr // Propagates ErrNotFound if order doesn't exist or doesn't belong to user
		}

		items, txErr = q.OrderRepo.GetItemsByOrderID(ctx, orderID)
		return txErr
	})

	if err != nil {
		s.logger.Error("failed to get user order details", "user_id", userID, "order_id", orderID, "error", err)
		return nil, err
	}

	// Map the database models to our detailed DTO.
	return dto.MapModelsToOrderWithItemsResponse(order, items), nil
}

// CancelOrder cancels an order, refunds the payment, and restocks items.
func (s *orderService) CancelOrder(ctx context.Context, userID, orderID int64) error {
	return s.store.ExecTx(ctx, func(q *domain.Queries) error {
		// 1. Get the order and lock the row for update.
		// This also implicitly checks if the order belongs to the user.
		order, err := q.OrderRepo.GetByIDForUpdate(ctx, orderID, userID)
		if err != nil {
			return err // Will be ErrNotFound if not found or no permission
		}

		// 2. Business Rule: Check if the order is in a cancellable state.
		if order.Status != models.OrderStatusConfirmed && order.Status != models.OrderStatusProcessing {
			s.logger.Warn("attempt to cancel order with non-cancellable status", "order_id", order.ID, "status", order.Status)
			return fmt.Errorf("order cannot be cancelled. status: %s", order.Status)
		}

		// 3. Issue the refund via the payment service.
		if err := s.paymentService.RefundPaymentIntent(ctx, order.PaymentIntentID); err != nil {
			s.logger.Error("failed to refund payment intent during cancellation", "order_id", order.ID, "pi_id", order.PaymentIntentID, "error", err)
			return fmt.Errorf("payment refund failed: %w", err)
		}

		// 4. Get order items to restock them.
		items, err := q.OrderRepo.GetItemsByOrderID(ctx, order.ID)
		if err != nil {
			return fmt.Errorf("failed to get order items for restocking: %w", err)
		}

		// 5. Restock each product.
		for _, item := range items {
			if err := q.ProductRepo.UpdateStock(ctx, item.ProductID, +item.Quantity); err != nil {
				return fmt.Errorf("failed to restock product %d: %w", item.ProductID, err)
			}
		}

		// 6. Update the order status to cancelled and payment status to refunded.
		s.logger.Info("order cancelled and refunded successfully", "order_id", order.ID)
		return q.OrderRepo.UpdateStatus(ctx, order.ID, models.OrderStatusCancelled, models.PaymentStatusRefunded, nil, nil)
	})
}

func (s *orderService) UpdateOrderStatus(
    ctx context.Context,
    orderID int64,
    status models.OrderStatus,
    paymentStatus *models.PaymentStatus,
    trackingNumber *string, 
	estimatedDeliveryDate *time.Time,
) error {
	return s.store.ExecTx(ctx, func(q *domain.Queries) error {
		order, err := q.OrderRepo.GetOrderByID(ctx, orderID)
		if err != nil {
			return err
		}
		
		finalPaymentStatus := order.PaymentStatus
		if paymentStatus != nil {
			finalPaymentStatus = *paymentStatus
		}
		
		s.logger.Info("Updating order status via internal call", "order_id", orderID, "new_status", status)
		
		return q.OrderRepo.UpdateStatus(ctx, order.ID, status, finalPaymentStatus, trackingNumber, estimatedDeliveryDate)
	})
}