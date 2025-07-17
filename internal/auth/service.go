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

type UserServiceInterface interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type Service struct {
	userService UserServiceInterface
}

func NewService(userService UserServiceInterface) *Service {
	return &Service{userService: userService}
}

func (s *Service) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.userService.GetByEmail(ctx, email)
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
