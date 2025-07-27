// events/types.go
package events

import (
	"encoding/json"
	"time"
)

// OrderCreatedEvent is the payload for the first task in the fulfillment pipeline.
type OrderCreatedEvent struct {
	OrderID         int64            `json:"order_id"`
	UserID          int64            `json:"user_id"`
	OrderNumber     string           `json:"order_number"`
	UserEmail       string           `json:"user_email"`
	TotalAmount     float64          `json:"total_amount"`
	OrderDate       time.Time        `json:"order_date"`
	Items           []OrderItemInfo  `json:"items"`
}

// OrderPackedEvent is triggered by the warehouse.
type OrderPackedEvent struct {
	OrderID     int64     `json:"order_id"`
	OrderNumber string    `json:"order_number"`
	UserID      int64     `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	PackedAt    time.Time `json:"packed_at"`
}

// OrderShippedEvent is triggered by the shipping service.
type OrderShippedEvent struct {
	OrderID               int64     `json:"order_id"`
	OrderNumber           string    `json:"order_number"`
	UserID                int64     `json:"user_id"`
	UserEmail             string    `json:"user_email"`
	TrackingNumber        string    `json:"tracking_number"`
	ShippedAt             time.Time `json:"shipped_at"`
	EstimatedDeliveryDate time.Time `json:"estimated_delivery_date"`
}

// OrderDeliveredEvent is triggered by the delivery service.
type OrderDeliveredEvent struct {
	OrderID     int64     `json:"order_id"`
	OrderNumber string    `json:"order_number"`
	UserID      int64     `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	DeliveredAt time.Time `json:"delivered_at"`
}

// NotificationEvent is a generic event for the notification service.
type NotificationEvent struct {
	UserEmail string `json:"user_email"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

// NotificationRequestEvent is the payload sent to the notification worker.
type NotificationRequestEvent struct {
	Type      string          `json:"type"` // e.g., "ORDER_CONFIRMED", "ORDER_SHIPPED"
	UserEmail string          `json:"user_email"`
	Payload   json.RawMessage `json:"payload"` 
}


// Add Item and Address structs for email templates
type OrderItemInfo struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type OrderAddressInfo struct {
	Name       string `json:"name"`
	Street1    string `json:"street1"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}