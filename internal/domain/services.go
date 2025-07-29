// internal/domain/services.go
package domain

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
)

// UserService handles user business logic
type UserService interface {
	Register(ctx context.Context, name, email, password string) (*models.User, error)
	RegisterWithCartMerge(ctx context.Context, store Store, name, email, password string, anonymousCartID *int64) (*models.User, *models.RefreshToken, error)
	GetProfile(ctx context.Context, userID int64) (*models.User, error)
	UpdateProfile(ctx context.Context, userID int64, name, email *string) (*models.User, error) 
	ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error
	DeleteAccount(ctx context.Context, userID int64, password string) error
}

// AuthService handles authentication business logic
type AuthService interface {
	LoginWithCartMerge(ctx context.Context, store Store, email, password string, anonymousCartID *int64) (*models.User, *models.RefreshToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.User, *models.RefreshToken, error)
	Logout(ctx context.Context, refreshToken string) error
	GetUserSessions(ctx context.Context, userID int64) ([]*models.RefreshToken, error)
	RevokeAllUserSessions(ctx context.Context, userID int64) error
	RevokeUserSession(ctx context.Context, userID, sessionID int64) error
	CleanupExpiredTokens(ctx context.Context) error
	GenerateAccessToken(ctx context.Context, user *models.User) (string, error)
	GenerateRefreshToken(ctx context.Context, userID int64) (*models.RefreshToken, error)
    ValidateToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) 
}

// AdminService handles admin-specific business logic
type AdminService interface {
	ListUsers(ctx context.Context) ([]*models.User, error)
	CreateUser(ctx context.Context, name, email, password string, role models.Role) (*models.User, error)
	UpdateUser(ctx context.Context, userID int64, name, email *string, role *models.Role) (*models.User, error)
	DeleteUser(ctx context.Context, userID int64) error
}

// ProductService handles product business logic
type ProductService interface {
	ListProducts(ctx context.Context, filters ProductFilters) ([]*models.Product, error)
	GetProduct(ctx context.Context, id int64) (*models.Product, error)
}

// CategoryService handles category business logic
type CategoryService interface {
	ListCategories(ctx context.Context) ([]*models.Category, error)
	GetOrCreate(ctx context.Context, name string) (*models.Category, error)
}

// CartService handles shopping cart operations
type CartService interface {
    GetOrCreateCart(ctx context.Context, userID *int64, anonymousCartID *int64) (*models.Cart, error)
    AddProductToCart(ctx context.Context, cartID int64, productID int64, quantity int) (*models.Cart, error)
    UpdateProductInCart(ctx context.Context, cartID int64, productID int64, quantity int) (*models.Cart, error)
    RemoveProductFromCart(ctx context.Context, cartID int64, productID int64) (*models.Cart, error)
    GetCartContents(ctx context.Context, cartID int64) (*models.Cart, error)
	HandleLoginWithTransaction(ctx context.Context, q *Queries, userID int64, anonymousCartID int64) error
	CleanupOldAnonymousCarts(ctx context.Context, olderThan time.Duration) (int64, error)
}

// AddressService handles user address operations
type AddressService interface {
    Create(ctx context.Context, userID int64, req *dto.CreateAddressRequest) (*models.UserAddress, error)
    GetByID(ctx context.Context, id int64, userID int64) (*models.UserAddress, error)
    ListByUserID(ctx context.Context, userID int64) ([]*models.UserAddress, error)
    Update(ctx context.Context, userID int64, id int64, req *dto.UpdateAddressRequest) (*models.UserAddress, error)
    Delete(ctx context.Context, userID int64, id int64) error
    SetDefault(ctx context.Context, userID int64, id int64, req string) error
}

// OrderService handles order business logic
type OrderService interface {
	CreateOrder(ctx context.Context, userID int64, cartID int64, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error)
	HandlePaymentSucceeded(ctx context.Context, paymentIntentID string) error
	ListUserOrders(ctx context.Context, userID int64) ([]*dto.OrderResponse, error) 
	GetUserOrder(ctx context.Context, userID, orderID int64) (*dto.OrderWithItemsResponse, error) 
	CleanupPendingOrders(ctx context.Context, olderThan time.Duration) (int, error)
	CancelOrder(ctx context.Context, userID, orderID int64) error 
	UpdateOrderStatus(ctx context.Context, orderID int64, status models.OrderStatus, paymentStatus *models.PaymentStatus, trackingNumber *string, estimatedDeliveryDate *time.Time,) error
}

// PaymentService defines the interface for a payment provider like Stripe.
type PaymentService interface {
	CreatePaymentIntent(ctx context.Context, amount float64) (*dto.PaymentIntent, error)
	RefundPaymentIntent(ctx context.Context, paymentIntentID string) error
}