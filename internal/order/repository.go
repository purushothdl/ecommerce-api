package order

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/purushothdl/ecommerce-api/internal/domain"
    "github.com/purushothdl/ecommerce-api/internal/models"
    apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
)

type orderRepository struct {
    db domain.DBTX
}

// NewOrderRepository creates a new OrderRepository
func NewOrderRepository(db domain.DBTX) domain.OrderRepository {
    return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
    query := `
        INSERT INTO orders (
            user_id, order_number, status, payment_status, payment_method, payment_intent_id,
            shipping_address, billing_address, subtotal, tax_amount, shipping_cost, discount_amount, total_amount,
            notes, tracking_number, estimated_delivery_date
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
        RETURNING id, created_at, updated_at
    `
    err := r.db.QueryRowContext(ctx, query,
        order.UserID, order.OrderNumber, order.Status, order.PaymentStatus, order.PaymentMethod, order.PaymentIntentID,
        order.ShippingAddress, order.BillingAddress, order.Subtotal, order.TaxAmount, order.ShippingCost, order.DiscountAmount, order.TotalAmount,
        order.Notes, order.TrackingNumber, order.EstimatedDeliveryDate,
    ).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to create order: %w", err)
    }
    return nil
}

func (r *orderRepository) CreateItems(ctx context.Context, items []*models.OrderItem) error {
    for _, item := range items {
        query := `
            INSERT INTO order_items (
                order_id, product_id, product_name, product_sku, product_image,
                unit_price, quantity, total_price
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING id, created_at
        `
        err := r.db.QueryRowContext(ctx, query,
            item.OrderID, item.ProductID, item.ProductName, item.ProductSKU, item.ProductImage,
            item.UnitPrice, item.Quantity, item.TotalPrice,
        ).Scan(&item.ID, &item.CreatedAt)
        if err != nil {
            return fmt.Errorf("failed to create order item: %w", err)
        }
    }
    return nil
}

func (r *orderRepository) GetByID(ctx context.Context, id int64, userID int64) (*models.Order, error) {
    query := `
        SELECT id, user_id, order_number, status, payment_status, payment_method, payment_intent_id,
               shipping_address, billing_address, subtotal, tax_amount, shipping_cost, discount_amount, total_amount,
               notes, tracking_number, estimated_delivery_date, created_at, updated_at
        FROM orders WHERE id = $1 AND user_id = $2
    `
    order := &models.Order{}
    err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
        &order.ID, &order.UserID, &order.OrderNumber, &order.Status, &order.PaymentStatus, &order.PaymentMethod, &order.PaymentIntentID,
        &order.ShippingAddress, &order.BillingAddress, &order.Subtotal, &order.TaxAmount, &order.ShippingCost, &order.DiscountAmount, &order.TotalAmount,
        &order.Notes, &order.TrackingNumber, &order.EstimatedDeliveryDate, &order.CreatedAt, &order.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, apperrors.ErrNotFound
    } else if err != nil {
        return nil, fmt.Errorf("failed to get order by ID: %w", err)
    }
    return order, nil
}

func (r *orderRepository) GetItemsByOrderID(ctx context.Context, orderID int64) ([]*models.OrderItem, error) {
    query := `
        SELECT id, order_id, product_id, product_name, product_sku, product_image,
               unit_price, quantity, total_price, created_at
        FROM order_items WHERE order_id = $1
    `
    rows, err := r.db.QueryContext(ctx, query, orderID)
    if err != nil {
        return nil, fmt.Errorf("failed to get order items: %w", err)
    }
    defer rows.Close()

    var items []*models.OrderItem
    for rows.Next() {
        item := &models.OrderItem{}
        if err := rows.Scan(
            &item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.ProductSKU, &item.ProductImage,
            &item.UnitPrice, &item.Quantity, &item.TotalPrice, &item.CreatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan order item: %w", err)
        }
        items = append(items, item)
    }
    return items, nil
}

func (r *orderRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Order, error) {
    query := `
        SELECT id, user_id, order_number, status, payment_status, payment_method, payment_intent_id,
               shipping_address, billing_address, subtotal, tax_amount, shipping_cost, discount_amount, total_amount,
               notes, tracking_number, estimated_delivery_date, created_at, updated_at
        FROM orders WHERE user_id = $1 ORDER BY created_at DESC
    `
    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get orders by user ID: %w", err)
    }
    defer rows.Close()

    var orders []*models.Order
    for rows.Next() {
        order := &models.Order{}
        if err := rows.Scan(
            &order.ID, &order.UserID, &order.OrderNumber, &order.Status, &order.PaymentStatus, &order.PaymentMethod, &order.PaymentIntentID,
            &order.ShippingAddress, &order.BillingAddress, &order.Subtotal, &order.TaxAmount, &order.ShippingCost, &order.DiscountAmount, &order.TotalAmount,
            &order.Notes, &order.TrackingNumber, &order.EstimatedDeliveryDate, &order.CreatedAt, &order.UpdatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan order: %w", err)
        }
        orders = append(orders, order)
    }
    return orders, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, status models.OrderStatus) error {
    query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
    res, err := r.db.ExecContext(ctx, query, status, id)
    if err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return apperrors.ErrNotFound
    }
    return nil
}

func (r *orderRepository) UpdatePaymentStatus(ctx context.Context, id int64, status models.PaymentStatus) error {
    query := `UPDATE orders SET payment_status = $1, updated_at = NOW() WHERE id = $2`
    res, err := r.db.ExecContext(ctx, query, status, id)
    if err != nil {
        return fmt.Errorf("failed to update payment status: %w", err)
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return apperrors.ErrNotFound
    }
    return nil
}
