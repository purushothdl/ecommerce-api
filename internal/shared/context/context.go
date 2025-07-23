// internal/shared/context/context.go
package context

import (
	"context"
	"errors"
)

// contextKey is a type for context keys to prevent key collisions
type contextKey string

// Context keys for storing and retrieving values from context
const (
	UserContextKey contextKey = "user"
	CartContextKey contextKey = "cart"
)

// UserContext represents the authenticated user in request context
type UserContext struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// SetUser adds a user to the context
func SetUser(ctx context.Context, user UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// GetUser retrieves the user from the context
func GetUser(ctx context.Context) (UserContext, error) {
	user, ok := ctx.Value(UserContextKey).(UserContext)
	if !ok {
		return UserContext{}, errors.New("user not found in context")
	}
	return user, nil
}

// GetUserID retrieves just the user ID from the context
func GetUserID(ctx context.Context) (int64, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// CartContext represents the cart in request context
type CartContext struct {
	ID int64 `json:"id"`
}

// SetCart adds a cart to the context
func SetCart(ctx context.Context, cart interface{}) context.Context {
	return context.WithValue(ctx, CartContextKey, cart)
}

// GetCart retrieves the cart from the context
func GetCart(ctx context.Context) (CartContext, error) {
	cart, ok := ctx.Value(CartContextKey).(CartContext)
	if !ok {
		return CartContext{}, errors.New("cart not found in context")
	}
	return cart, nil
}
