// internal/domain/repositories.go
package domain

import (
	"context"
	"database/sql"
	"time"

	"github.com/purushothdl/ecommerce-api/internal/models"
)

// UserRepository handles user data operations
type UserRepository interface {
	Insert(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int64) error
	GetAll(ctx context.Context) ([]*models.User, error) 
}

// AuthRepository handles authentication data operations
type AuthRepository interface {
	StoreRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenPlaintext string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenPlaintext string) error
	GetUserRefreshTokens(ctx context.Context, userID int64) ([]*models.RefreshToken, error)
	RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error
	RevokeRefreshTokenByID(ctx context.Context, tokenID int64) error
	RevokeUserSessionByID(ctx context.Context, userID, sessionID int64) error
	CleanupExpiredTokens(ctx context.Context) error
}

// ProductRepository handles product data operations
type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetAll(ctx context.Context, filters ProductFilters) ([]*models.Product, error)
	GetByID(ctx context.Context, id int64) (*models.Product, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*models.Product, error) 
	UpdateStock(ctx context.Context, productID int64, quantityChange int) error
}

// CategoryRepository handles category data operations
type CategoryRepository interface {
	GetByName(ctx context.Context, name string) (*models.Category, error)
	GetAll(ctx context.Context) ([]*models.Category, error)
	Create(ctx context.Context, category *models.Category) error
}

// CartRepository handles shopping cart data operations
type CartRepository interface {
    // Cart methods
    GetByUserID(ctx context.Context, userID int64) (*models.Cart, error)
    GetByID(ctx context.Context, cartID int64) (*models.Cart, error)
    Create(ctx context.Context, userID *int64) (*models.Cart, error)
    MergeCarts(ctx context.Context, fromCartID, toCartID int64) error
	Delete(ctx context.Context, cartID int64) error
	ClearCart(ctx context.Context, cartID int64) error

    // CartItem methods
    AddItem(ctx context.Context, cartID int64, productID int64, quantity int) error
    UpdateItemQuantity(ctx context.Context, cartID int64, productID int64, quantity int) error
    RemoveItem(ctx context.Context, cartID int64, productID int64) error
	GetItemsByCartID(ctx context.Context, cartID int64) ([]models.CartItem, error)
	CleanupOldAnonymousCarts(ctx context.Context, olderThan time.Time) (int64, error)

}

// AddressRepository handles user address data operations
type AddressRepository interface {
    Create(ctx context.Context, addr *models.UserAddress) error
    GetByID(ctx context.Context, id int64) (*models.UserAddress, error)
    GetByUserID(ctx context.Context, userID int64) ([]*models.UserAddress, error)
    Update(ctx context.Context, addr *models.UserAddress) error
    Delete(ctx context.Context, id int64, userID int64) error
    UnsetDefaultShipping(ctx context.Context, userID int64) error
    UnsetDefaultBilling(ctx context.Context, userID int64) error
}

type OrderRepository interface {
    Create(ctx context.Context, order *models.Order) error
    CreateItems(ctx context.Context, items []*models.OrderItem) error
    GetByID(ctx context.Context, id int64, userID int64) (*models.Order, error)
	GetByIDForUpdate(ctx context.Context, id int64, userID int64) (*models.Order, error)
    GetItemsByOrderID(ctx context.Context, orderID int64) ([]*models.OrderItem, error)
    GetByUserID(ctx context.Context, userID int64) ([]*models.Order, error)
	GetOrderByID(ctx context.Context, id int64) (*models.Order, error)   

	FindPendingOrdersOlderThan(ctx context.Context, olderThan time.Time) ([]*models.Order, error) 
	GetByPaymentIntentID(ctx context.Context, paymentIntentID string) (*models.Order, error)
	UpdateStatus(ctx context.Context,id int64,status models.OrderStatus,paymentStatus models.PaymentStatus,trackingNumber *string, estimatedDeliveryDate *time.Time) error
}

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Queries is a container for all your repository types. This is the key change.
type Queries struct {
	UserRepo     UserRepository
	CartRepo     CartRepository
	ProductRepo  ProductRepository
	AuthRepo     AuthRepository
	AddressRepo  AddressRepository
	OrderRepo    OrderRepository

}