// internal/domain/repositories.go
package domain

import (
	"context"
	"github.com/purushothdl/ecommerce-api/internal/models"
)

// UserRepository handles user data operations
type UserRepository interface {
	Insert(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int64) error
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
