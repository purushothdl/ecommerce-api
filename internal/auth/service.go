package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
	"github.com/purushothdl/ecommerce-api/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// UserFinder describes any type that can find a user by email.
// This interface lives INSIDE the auth package.
type UserFinder interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

// Service defines the business logic for authentication.
type Service interface {
	Login(ctx context.Context, email, password string) (*models.User, error)
}

type service struct {
	userFinder UserFinder
}

func NewService(userFinder UserFinder) *service {
	return &service{userFinder: userFinder}
}

func (s *service) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.userFinder.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("could not process login: %w", err)
	}

	err = crypto.CheckPasswordHash(password, user.PasswordHash)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("could not process login: %w", err)
	}

	return user, nil
}
