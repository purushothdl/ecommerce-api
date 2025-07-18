// internal/user/service.go
package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
)

// Service defines the business logic operations for users
type Service interface {
	Register(ctx context.Context, name, email, password string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetRefreshToken(ctx context.Context, tokenPlaintext string) (*models.RefreshToken, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, name, email, password string) (*models.User, error) {
	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         "user",
	}

	err = s.repo.Insert(ctx, user)
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicateEmail) {
			return nil, apperrors.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("could not register user: %w", err)
	}
	return user, nil
}

func (s *service) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("could not get user by email: %w", err)
	}
	return user, nil
}

func (s *service) GetRefreshToken(ctx context.Context, tokenPlaintext string) (*models.RefreshToken, error) {
	return s.repo.GetRefreshToken(ctx, tokenPlaintext)
}
