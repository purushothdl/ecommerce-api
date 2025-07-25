package dto

// CreateOrderRequest represents the input for creating an order
type CreateOrderRequest struct {
    ShippingAddressID  int64  `json:"shipping_address_id"`
    BillingAddressID   int64  `json:"billing_address_id"`
    PaymentMethod      string `json:"payment_method" example:"stripe"`
}



// ConfirmPaymentRequest represents the input for confirming payment
type ConfirmPaymentRequest struct {
    PaymentIntentID string `json:"payment_intent_id"`
}

// CancelOrderRequest represents the input for cancelling an order (if needed, e.g., with reason)
type CancelOrderRequest struct {
    Reason string `json:"reason,omitempty"`
}

// OrderResponse represents a single order output
type OrderResponse struct {
    ID                    int64          `json:"id"`
    OrderNumber           string         `json:"order_number"`
    Status                string         `json:"status"` // Use string for enum in responses
    PaymentStatus         string         `json:"payment_status"`
    PaymentMethod         string         `json:"payment_method"`
    ShippingAddress       OrderAddress   `json:"shipping_address"`
    BillingAddress        OrderAddress   `json:"billing_address"`
    Subtotal              float64        `json:"subtotal"`
    TaxAmount             float64        `json:"tax_amount"`
    ShippingCost          float64        `json:"shipping_cost"`
    DiscountAmount        float64        `json:"discount_amount"`
    TotalAmount           float64        `json:"total_amount"`
    Notes                 string         `json:"notes,omitempty"`
    TrackingNumber        string         `json:"tracking_number,omitempty"`
    EstimatedDeliveryDate string         `json:"estimated_delivery_date,omitempty"` // String for JSON
    CreatedAt             string         `json:"created_at"` // String for JSON
    UpdatedAt             string         `json:"updated_at"`
}

// OrderWithItemsResponse represents an order with its items
type OrderWithItemsResponse struct {
    Order OrderResponse `json:"order"`
    Items []OrderItemResponse `json:"items"`
}

// OrderItemResponse represents a single order item output
type OrderItemResponse struct {
    ID           int64   `json:"id"`
    ProductID    int64   `json:"product_id"`
    ProductName  string  `json:"product_name"`
    ProductSKU   string  `json:"product_sku"`
    ProductImage string  `json:"product_image,omitempty"`
    UnitPrice    float64 `json:"unit_price"`
    Quantity     int     `json:"quantity"`
    TotalPrice   float64 `json:"total_price"`
}

// OrderListResponse represents a list of orders
type OrderListResponse struct {
    Orders []*OrderResponse `json:"orders"`
}

// OrderAddress is a shared DTO for addresses in orders
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
