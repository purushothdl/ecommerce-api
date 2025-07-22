// internal/cart/responses.go
package cart

import (
    "time"
    "github.com/purushothdl/ecommerce-api/internal/models"
)

// CartResponse represents a cart with its items and computed totals
type CartResponse struct {
    ID        int64            `json:"id"`
    UserID    *int64           `json:"user_id,omitempty"`
    Items     []CartItemResponse `json:"items"`
    Total     float64          `json:"total"`
    ItemCount int              `json:"item_count"`
    CreatedAt time.Time        `json:"created_at"`
    UpdatedAt time.Time        `json:"updated_at"`
}

// CartItemResponse represents a single item in the cart
type CartItemResponse struct {
    ID        int64              `json:"id"`
    Product   CartProductResponse `json:"product"`  // Clean product DTO
    Quantity  int                `json:"quantity"`
    Subtotal  float64            `json:"subtotal"`
    CreatedAt time.Time          `json:"created_at"`
    UpdatedAt time.Time          `json:"updated_at"`
}

// CartProductResponse represents only the product fields needed in cart
type CartProductResponse struct {
    ID            int64   `json:"id"`
    Name          string  `json:"name"`
    Price         float64 `json:"price"`
    Thumbnail     string  `json:"thumbnail"`
    StockQuantity int     `json:"stock_quantity"`
}

// NewCartResponse creates a CartResponse from models
func NewCartResponse(cart *models.Cart, items []models.CartItem) *CartResponse {
    cartItems := make([]CartItemResponse, len(items))
    total := 0.0
    
    for i, item := range items {
        subtotal := float64(item.Quantity) * item.Product.Price
        total += subtotal
        
        cartItems[i] = CartItemResponse{
            ID: item.ID,
            Product: CartProductResponse{
                ID:            item.Product.ID,
                Name:          item.Product.Name,
                Price:         item.Product.Price,
                Thumbnail:     item.Product.Thumbnail,
                StockQuantity: item.Product.StockQuantity,
            },
            Quantity:  item.Quantity,
            Subtotal:  subtotal,
            CreatedAt: item.CreatedAt,
            UpdatedAt: item.UpdatedAt,
        }
    }
    
    return &CartResponse{
        ID:        cart.ID,
        UserID:    cart.UserID,
        Items:     cartItems,
        Total:     total,
        ItemCount: len(items),
        CreatedAt: cart.CreatedAt,
        UpdatedAt: cart.UpdatedAt,
    }
}
