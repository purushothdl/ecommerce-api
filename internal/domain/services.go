// internal/domain/services.go
package domain

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/purushothdl/ecommerce-api/internal/models"
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
}

