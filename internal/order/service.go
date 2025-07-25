package order

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	"github.com/purushothdl/ecommerce-api/pkg/utils/order_number"

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
		// 1. Get cart items. A new method on CartRepository might be needed to get items without the full cart object.
		cartItems, err := q.CartRepo.GetItemsByCartID(ctx, cartID)
		if err != nil {
			s.logger.Error("failed to get cart items for order creation", "cart_id", cartID, "error", err)
			return fmt.Errorf("could not retrieve cart for order: %w", err)
		}
		if len(cartItems) == 0 {
			return errors.New("cannot create an order from an empty cart")
		}

		// 2. Lock products, validate stock, and calculate totals.
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

		// 3. Create Stripe Payment Intent. This call happens within the DB transaction.
		stripePI, err := s.paymentService.CreatePaymentIntent(ctx, totalAmount)
		if err != nil {
			s.logger.Error("failed to create stripe payment intent", "error", err)
			return fmt.Errorf("payment provider error: %w", err)
		}

		// 4. Create the main Order record.
		order := &models.Order{
			UserID:          userID,
			OrderNumber:     order_number.Generate(),
			Status:          models.OrderStatusPendingPayment,
			PaymentStatus:   models.PaymentStatusPending,
			PaymentMethod:   req.PaymentMethod,
			PaymentIntentID: stripePI.ID,
			Subtotal:        subtotal,
			TotalAmount:     totalAmount,
			// TODO: Populate address from req.ShippingAddressID or req.ShippingAddress
		}
		if err := q.OrderRepo.Create(ctx, order); err != nil {
			s.logger.Error("failed to save order", "error", err)
			return fmt.Errorf("could not save order: %w", err)
		}

		// 5. Create Order Items and update stock.
		for _, item := range cartItems {
			product := productSnapshots[item.Product.ID]
			orderItem := &models.OrderItem{
				OrderID:      order.ID,
				ProductID:    product.ID,
				ProductName:  product.Name, 
				ProductSKU:   product.SKU,
				UnitPrice:    product.Price,
				Quantity:     item.Quantity,
				TotalPrice:   product.Price * float64(item.Quantity),
			}
			if err := q.OrderRepo.CreateItems(ctx, []*models.OrderItem{orderItem}); err != nil {
				s.logger.Error("failed to save order item", "error", err)
				return fmt.Errorf("could not save order item: %w", err)
			}
			
			// Decrement stock
			if err := q.ProductRepo.UpdateStock(ctx, product.ID, -item.Quantity); err != nil {
				return fmt.Errorf("failed to update stock for product %d: %w", product.ID, err)
			}
		}

		// 6. Clear the cart.
		if err := q.CartRepo.ClearCart(ctx, cartID); err != nil {
		 return fmt.Errorf("failed to clear cart: %w", err)
		}

		// 7. Set the response object to be returned by the outer function.
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