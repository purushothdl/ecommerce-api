package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
	"golang.org/x/crypto/bcrypt"
)

// Extending the interface for our user repository
type UserRepo interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	StoreRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenPlaintext string) (*models.RefreshToken, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	RevokeRefreshToken(ctx context.Context, tokenPlaintext string) error
}

// Service defines the business logic for authentication.
type Service interface {
	Login(ctx context.Context, email, password string) (*models.User, *models.RefreshToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.User, *models.RefreshToken, error)
	Logout(ctx context.Context, refreshToken string) error
}

type service struct {
	userRepo UserRepo
}

func NewService(userRepo UserRepo) *service {
	return &service{userRepo: userRepo}
}

func (s *service) Login(ctx context.Context, email, password string) (*models.User, *models.RefreshToken, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, nil, apperrors.ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("could not process login: %w", err)
	}

	err = crypto.CheckPasswordHash(password, user.PasswordHash)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, nil, apperrors.ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("could not process login: %w", err)
	}

	// Generate and store the refresh token
	refreshToken, err := GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate refresh token: %w", err)
	}

	if err := s.userRepo.StoreRefreshToken(ctx, refreshToken); err != nil {
		return nil, nil, fmt.Errorf("could not store refresh token: %w", err)
	}

	return user, refreshToken, nil
}

func (s *service) RefreshToken(ctx context.Context, tokenPlaintext string) (*models.User, *models.RefreshToken, error) {
	token, err := s.userRepo.GetRefreshToken(ctx, tokenPlaintext)
	if err != nil {
		return nil, nil, apperrors.ErrInvalidToken
	}

	if token.ExpiresAt.Before(time.Now()) {
		return nil, nil, apperrors.ErrTokenExpired
	}

	// Fetch the user by ID (not email)
	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("could not fetch user: %w", err)
	}

	return user, token, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	return s.userRepo.RevokeRefreshToken(ctx, refreshToken)
}
