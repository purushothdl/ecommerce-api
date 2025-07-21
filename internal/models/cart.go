// internal/models/cart.go
package models


type Cart struct {
	BaseModel
    UserID    *int64     `json:"user_id"` // Pointer to handle NULL for anonymous users
    Items     []CartItem `json:"items,omitempty"` // For eager loading items
    Total     float64    `json:"total,omitempty"` // Calculated field
}

type CartItem struct {
    ID        int64    `json:"-"` // Internal ID, not exposed
    CartID    int64    `json:"-"`
    Product   *Product `json:"product"` // Eager load product details
    Quantity  int      `json:"quantity"`
}