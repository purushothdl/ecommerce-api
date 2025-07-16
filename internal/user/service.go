// internal/user/service.go
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Repository is an interface that our data layer must satisfy.
// This is for decoupling and easier testing.
type RepositoryInterface interface {
	Insert(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

// Service provides user-related business logic.
type Service struct {
	repo RepositoryInterface
}

// NewService creates a new user service.
func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// Register handles the business logic for creating a new user.
func (s *Service) Register(ctx context.Context, name, email, password string) (*models.User, error) {
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
		if errors.Is(err, ErrDuplicateEmail) {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("could not register user: %w", err)
	}
	return user, nil
}

// Login handles business logic for user authentication.
func (s *Service) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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