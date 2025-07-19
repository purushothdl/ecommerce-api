// internal/user/service.go
package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
	"github.com/purushothdl/ecommerce-api/pkg/utils/ptr"
)

type userService struct {
	userRepo domain.UserRepository
	logger   *slog.Logger
}

// NewUserService returns a domain.UserService implementation
func NewUserService(userRepo domain.UserRepository, logger *slog.Logger) domain.UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *userService) Register(ctx context.Context, name, email, password string) (*models.User, error) {
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return nil, fmt.Errorf("user service: failed to hash password: %w", err)
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         "user",
	}

	if err := s.userRepo.Insert(ctx, user); err != nil {
		s.logger.Error("failed to create user", "email", email, "error", err)
		return nil, fmt.Errorf("user service: failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to fetch user profile", "user_id", userID, "error", err)
		return nil, fmt.Errorf("user service: failed to get user profile: %w", err)
	}

	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to fetch user for password change", "user_id", userID, "error", err)
		return fmt.Errorf("user service: failed to get user: %w", err)
	}

	if err := crypto.CheckPasswordHash(currentPassword, user.PasswordHash); err != nil {
		if errors.Is(err, crypto.ErrInvalidCredentials) {
			s.logger.Warn("invalid current password", "user_id", userID)
			return apperrors.ErrInvalidCredentials
		}

		s.logger.Error("failed to verify current password", "user_id", userID, "error", err)
		return fmt.Errorf("user service: failed to verify current password: %w", err)
	}

	hashedNewPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		s.logger.Error("failed to hash new password", "user_id", userID, "error", err)
		return fmt.Errorf("user service: failed to hash new password: %w", err)
	}

	user.PasswordHash = hashedNewPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update password", "user_id", userID, "error", err)
		return fmt.Errorf("user service: failed to update password: %w", err)
	}

	return nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID int64, name, email *string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to fetch user for update", "user_id", userID, "error", err)
		return nil, fmt.Errorf("user service: failed to get user: %w", err)
	}

	ptr.UpdateStringIfProvided(&user.Name, name)
	ptr.UpdateStringIfProvided(&user.Email, email)

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update profile", "user_id", userID, "error", err)
		return nil, fmt.Errorf("user service: failed to update profile: %w", err)
	}

	return user, nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID int64, password string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to fetch user for deletion", "user_id", userID, "error", err)
		return fmt.Errorf("user service: failed to get user: %w", err)
	}

	if err := crypto.CheckPasswordHash(password, user.PasswordHash); err != nil {
		if errors.Is(err, crypto.ErrInvalidCredentials) {
			s.logger.Warn("invalid password for account deletion", "user_id", userID)
			return apperrors.ErrInvalidCredentials
		}

		s.logger.Error("failed to verify password", "user_id", userID, "error", err)
		return fmt.Errorf("user service: failed to verify password: %w", err)
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.Error("failed to delete account", "user_id", userID, "error", err)
		return fmt.Errorf("user service: failed to delete account: %w", err)
	}

	return nil
}
