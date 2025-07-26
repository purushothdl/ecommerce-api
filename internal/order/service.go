package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors" 
	"github.com/purushothdl/ecommerce-api/pkg/utils/orders"
)

type orderService struct {
	store          domain.Store
	paymentService domain.PaymentService 
	logger         *slog.Logger
}

// NewOrderService creates a new OrderService
func NewOrderService(store domain.Store, paymentService domain.PaymentService, logger *slog.Logger) domain.OrderService {
	return &orderService{
		store:          store,
		paymentService: paymentService,
		logger:         logger,
	}
}


// CreateOrder handles the entire process of creating an order.
func (s *orderService) CreateOrder(ctx context.Context, userID int64, cartID int64, req *dto.CreateOrderRequest) (*models.PaymentIntent, error) {
	var paymentIntent *models.PaymentIntent

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

		// TODO: Add tax and shipping calculation logic here
		totalAmount := subtotal

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

		order := &models.Order{
			UserID:          userID,
			OrderNumber:     orders.Generate(),
			Status:          models.OrderStatusPendingPayment,
			PaymentStatus:   models.PaymentStatusPending,
			PaymentMethod:   req.PaymentMethod,
			PaymentIntentID: stripePI.ID,
			Subtotal:        subtotal,
			TotalAmount:     totalAmount,
			ShippingAddress: json.RawMessage(shippingJSON),
			BillingAddress:  json.RawMessage(billingJSON),
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
		paymentIntent = stripePI
		return nil
	})

	return paymentIntent, err
}


func (s *orderService) HandlePaymentSucceeded(ctx context.Context, paymentIntentID string) error {
	return s.store.ExecTx(ctx, func(q *domain.Queries) error {
		order, err := q.OrderRepo.GetByPaymentIntentID(ctx, paymentIntentID)
		if err != nil {
			s.logger.Error("webhook cannot find order for payment intent", "pi_id", paymentIntentID)
			return err
		}

		if order.PaymentStatus == models.PaymentStatusPaid {
			s.logger.Info("webhook received for already-paid order, ignoring", "order_id", order.ID)
			return nil
		}
		
		s.logger.Info("updating order status to confirmed/paid", "order_id", order.ID, "pi_id", paymentIntentID)
		return q.OrderRepo.UpdateStatus(ctx, order.ID, models.OrderStatusConfirmed, models.PaymentStatusPaid)
	})
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
		return q.OrderRepo.UpdateStatus(ctx, order.ID, models.OrderStatusCancelled, models.PaymentStatusRefunded)
	})
}