package models

import "time"

// OrderItem represents a line item in an order
type OrderItem struct {
    ID            int64   `json:"id"`
    OrderID       int64   `json:"order_id"`
    ProductID     int64   `json:"product_id"`
    ProductName   string  `json:"product_name"`
    ProductSKU    string  `json:"product_sku"`
    ProductImage  string  `json:"product_image,omitempty"`
    UnitPrice     float64 `json:"unit_price"`
    Quantity      int     `json:"quantity"`
    TotalPrice    float64 `json:"total_price"`
    CreatedAt     time.Time `json:"created_at"`
}
