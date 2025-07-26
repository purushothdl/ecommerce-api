package dto

import (
	"encoding/json"
	"time"

	"github.com/purushothdl/ecommerce-api/internal/models"
)

// CreateOrderRequest represents the input for creating an order
type CreateOrderRequest struct {
    ShippingAddressID  int64  `json:"shipping_address_id"`
    BillingAddressID   int64  `json:"billing_address_id"`
    PaymentMethod      string `json:"payment_method" example:"stripe"`
}

// CreateOrderResponse is the specific data returned after successfully creating an order.
type CreateOrderResponse struct {
	OrderID      int64  `json:"order_id"`
	OrderNumber  string `json:"order_number"`
	ClientSecret string `json:"client_secret"` 
}

// ConfirmPaymentRequest represents the input for confirming payment
type ConfirmPaymentRequest struct {
    PaymentIntentID string `json:"payment_intent_id"`
}

// CancelOrderRequest represents the input for cancelling an order (if needed, e.g., with reason)
type CancelOrderRequest struct {
    Reason string `json:"reason,omitempty"`
}

// UpdateOrderStatusRequest is the payload for the internal status update endpoint.
type UpdateOrderStatusRequest struct {
	Status                models.OrderStatus    `json:"status"`
	PaymentStatus         *models.PaymentStatus `json:"payment_status,omitempty"`
	TrackingNumber        *string               `json:"tracking_number,omitempty"`
	EstimatedDeliveryDate *time.Time           ` json:"estimated_delivery_date,omitempty"`
}

// OrderResponse represents a single order output, used in lists
type OrderResponse struct {
	ID            int64                `json:"id"`
	OrderNumber   string               `json:"order_number"`
	Status        models.OrderStatus   `json:"status"`
	PaymentStatus models.PaymentStatus `json:"payment_status"`
	TotalAmount   float64              `json:"total_amount"`
	CreatedAt     time.Time            `json:"created_at"`
}

// OrderWithItemsResponse represents a detailed single order with its items
type OrderWithItemsResponse struct {
	ID                    int64                `json:"id"`
	UserID                int64                `json:"user_id"`
	OrderNumber           string               `json:"order_number"`
	Status                models.OrderStatus   `json:"status"`
	PaymentStatus         models.PaymentStatus `json:"payment_status"`
	PaymentMethod         string               `json:"payment_method"`
	ShippingAddress       json.RawMessage      `json:"shipping_address"`
	BillingAddress        json.RawMessage      `json:"billing_address"`
	Subtotal              float64              `json:"subtotal"`
	TaxAmount             float64              `json:"tax_amount"`
	ShippingCost          float64              `json:"shipping_cost"`
	DiscountAmount        float64              `json:"discount_amount"`
	TotalAmount           float64              `json:"total_amount"`
	TrackingNumber        string               `json:"tracking_number,omitempty"`
	EstimatedDeliveryDate time.Time            `json:"estimated_delivery_date,omitempty"`
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
	Items                 []*OrderItemResponse `json:"items"`
}

// OrderItemResponse represents a single order item output
type OrderItemResponse struct {
	ID           int64   `json:"id"`
	ProductID    int64   `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductImage string  `json:"product_image,omitempty"`
	UnitPrice    float64 `json:"unit_price"`
	Quantity     int     `json:"quantity"`
	TotalPrice   float64 `json:"total_price"`
}

// MapModelsToOrderWithItemsResponse is a helper to convert DB models to a DTO
func MapModelsToOrderWithItemsResponse(order *models.Order, items []*models.OrderItem) *OrderWithItemsResponse {
	orderItems := make([]*OrderItemResponse, len(items))
	for i, item := range items {
		orderItems[i] = &OrderItemResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			UnitPrice:    item.UnitPrice,
			Quantity:     item.Quantity,
			TotalPrice:   item.TotalPrice,
		}
	}

	return &OrderWithItemsResponse{
		ID:                    order.ID,
		UserID:                order.UserID,
		OrderNumber:           order.OrderNumber,
		Status:                order.Status,
		PaymentStatus:         order.PaymentStatus,
		PaymentMethod:         order.PaymentMethod,
		ShippingAddress:       order.ShippingAddress,
		BillingAddress:        order.BillingAddress,
		Subtotal:              order.Subtotal,
		TaxAmount:             order.TaxAmount,
		ShippingCost:          order.ShippingCost,
		DiscountAmount:        order.DiscountAmount,
		TotalAmount:           order.TotalAmount,
		TrackingNumber:        order.TrackingNumber,
		EstimatedDeliveryDate: order.EstimatedDeliveryDate,
		CreatedAt:             order.CreatedAt,
		UpdatedAt:             order.UpdatedAt,
		Items:                 orderItems,
	}
}