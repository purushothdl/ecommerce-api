// internal/user/service.go
package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/utils/ptr"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
)

type userService struct {
	userRepo domain.UserRepository
}

// NewUserService returns a domain.UserService implementation
func NewUserService(userRepo domain.UserRepository) domain.UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Register(ctx context.Context, name, email, password string) (*models.User, error) {
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("user service: failed to hash password: %w", err)
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "user",
	}

	if err := s.userRepo.Insert(ctx, user); err != nil {
		return nil, fmt.Errorf("user service: failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user service: failed to get user profile: %w", err)
	}
	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	// First, get the user to verify current password
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user service: failed to get user: %w", err)
	}

	// Verify current password
	if err := crypto.CheckPasswordHash(currentPassword, user.PasswordHash); err != nil {
		if errors.Is(err, crypto.ErrInvalidCredentials) {
			return apperrors.ErrInvalidCredentials
		}
		return fmt.Errorf("user service: failed to verify current password: %w", err)
	}

	// Hash the new password
	hashedNewPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("user service: failed to hash new password: %w", err)
	}

	// Update the user's password
	user.PasswordHash = hashedNewPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("user service: failed to update password: %w", err)
	}

	return nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID int64, name, email *string) (*models.User, error) {
	// Get the current user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user service: failed to get user: %w", err)
	}

	// Update only the fields that are provided
	ptr.UpdateStringIfProvided(&user.Name, name)
	ptr.UpdateStringIfProvided(&user.Email, email)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("user service: failed to update profile: %w", err)
	}

	return user, nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID int64, password string) error {
	// First, get the user to verify password
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user service: failed to get user: %w", err)
	}

	// Verify password before deletion
	if err := crypto.CheckPasswordHash(password, user.PasswordHash); err != nil {
		if errors.Is(err, crypto.ErrInvalidCredentials) {
			return apperrors.ErrInvalidCredentials
		}
		return fmt.Errorf("user service: failed to verify password: %w", err)
	}

	// Delete the user
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("user service: failed to delete account: %w", err)
	}

	return nil
}
