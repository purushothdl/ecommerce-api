package models

import (
	"encoding/json"
	"time"
)

// OrderStatus represents the possible states of an order
type OrderStatus string

const (
    OrderStatusPendingPayment   OrderStatus = "pending_payment"
    OrderStatusConfirmed        OrderStatus = "confirmed"
    OrderStatusProcessing       OrderStatus = "processing"
    OrderStatusShipped          OrderStatus = "shipped"
    OrderStatusOutForDelivery   OrderStatus = "out_for_delivery"
    OrderStatusDelivered        OrderStatus = "delivered"
    OrderStatusCancelled        OrderStatus = "cancelled"
)

// PaymentStatus represents the possible payment states
type PaymentStatus string

const (
    PaymentStatusPending  PaymentStatus = "pending"
    PaymentStatusPaid     PaymentStatus = "paid"
    PaymentStatusFailed   PaymentStatus = "failed"
    PaymentStatusRefunded PaymentStatus = "refunded"
)

// OrderAddress represents a snapshot of an address for orders (stored as JSONB)
type OrderAddress struct {
    Name       string `json:"name"`
    Phone      string `json:"phone"`
    Street1    string `json:"street1"`
    Street2    string `json:"street2,omitempty"`
    City       string `json:"city"`
    State      string `json:"state"`
    PostalCode string `json:"postal_code"`
    Country    string `json:"country"`
}

// Order represents an order in the database
type Order struct {
    ID                    int64          `json:"id"`
    UserID                int64          `json:"user_id"`
    OrderNumber           string         `json:"order_number"`
    Status                OrderStatus    `json:"status"`
    PaymentStatus         PaymentStatus  `json:"payment_status"`
    PaymentMethod         string         `json:"payment_method"`
    PaymentIntentID       string         `json:"payment_intent_id,omitempty"`
    ShippingAddress       json.RawMessage `json:"shipping_address"` 
    BillingAddress        json.RawMessage `json:"billing_address"`
    Subtotal              float64        `json:"subtotal"`
    TaxAmount             float64        `json:"tax_amount"`
    ShippingCost          float64        `json:"shipping_cost"`
    DiscountAmount        float64        `json:"discount_amount"`
    TotalAmount           float64        `json:"total_amount"`
    Notes                 string         `json:"notes,omitempty"`
    TrackingNumber        string         `json:"tracking_number,omitempty"`
    EstimatedDeliveryDate time.Time      `json:"estimated_delivery_date,omitempty"`
    CreatedAt             time.Time      `json:"created_at"`
    UpdatedAt             time.Time      `json:"updated_at"`
}
