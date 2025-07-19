// internal/admin/service.go
package admin

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
	"github.com/purushothdl/ecommerce-api/pkg/utils/ptr"
)

type adminService struct {
	userRepo domain.UserRepository
	logger   *slog.Logger
}

func NewAdminService(userRepo domain.UserRepository, logger *slog.Logger) domain.AdminService {
	return &adminService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *adminService) ListUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to list users", "error", err)
		return nil, fmt.Errorf("could not retrieve users: %w", err)
	}
	return users, nil
}

func (s *adminService) CreateUser(ctx context.Context, name, email, password string, role models.Role) (*models.User, error) {
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		s.logger.Error("failed to hash password during admin user creation", "error", err)
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	if err := s.userRepo.Insert(ctx, user); err != nil {
		s.logger.Error("failed to insert new user via admin", "email", email, "error", err)
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	s.logger.Info("admin successfully created a new user", "email", email, "role", role)
	return user, nil
}

func (s *adminService) UpdateUser(ctx context.Context, userID int64, name, email *string, role *models.Role) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user for admin update", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	ptr.UpdateStringIfProvided(&user.Name, name)
	ptr.UpdateStringIfProvided(&user.Email, email)

	if role != nil {
		user.Role = *role
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user via admin", "user_id", userID, "error", err)
		return nil, fmt.Errorf("could not update user: %w", err)
	}

	s.logger.Info("admin successfully updated user", "user_id", userID)
	return user, nil
}

func (s *adminService) DeleteUser(ctx context.Context, userID int64) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.Error("failed to delete user via admin", "user_id", userID, "error", err)
		return fmt.Errorf("could not delete user: %w", err)
	}
	s.logger.Info("admin successfully deleted user", "user_id", userID)
	return nil
}