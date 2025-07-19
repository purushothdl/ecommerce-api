// internal/domain/services.go
package domain

import (
	"context"

	"github.com/purushothdl/ecommerce-api/internal/models"
)

// UserService handles user business logic
type UserService interface {
	Register(ctx context.Context, name, email, password string) (*models.User, error)
	GetProfile(ctx context.Context, userID int64) (*models.User, error)
	UpdateProfile(ctx context.Context, userID int64, name, email *string) (*models.User, error) 
	ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error
	DeleteAccount(ctx context.Context, userID int64, password string) error
}

// AuthService handles authentication business logic
type AuthService interface {
	Login(ctx context.Context, email, password string) (*models.User, *models.RefreshToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.User, *models.RefreshToken, error)
	Logout(ctx context.Context, refreshToken string) error
	GetUserSessions(ctx context.Context, userID int64) ([]*models.RefreshToken, error)
	RevokeAllUserSessions(ctx context.Context, userID int64) error
	RevokeUserSession(ctx context.Context, userID, sessionID int64) error
	CleanupExpiredTokens(ctx context.Context) error
}
