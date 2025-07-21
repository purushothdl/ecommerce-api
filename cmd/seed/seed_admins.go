// cmd/seed/seed_admins.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"

	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
)

var SeedAdminsTask = SeederTask{
	Name: "Seed Admins",
	Run: func(ctx context.Context, deps SeederDeps) error {
		adminEmail := os.Getenv("ADMIN_EMAIL")
		adminPassword := os.Getenv("ADMIN_PASSWORD")

		if adminEmail == "" || adminPassword == "" {
			return fmt.Errorf("ADMIN_EMAIL and ADMIN_PASSWORD must be set")
		}

		// Check if admin already exists using the repository
		_, err := deps.UserRepo.GetByEmail(ctx, adminEmail)
		if err == nil {
			log.Printf("Admin user with email %s already exists. Skipping.", adminEmail)
			return nil // Not an error, successfully skipped.
		}
		if !errors.Is(err, apperrors.ErrUserNotFound) {
			// A different error occurred
			return fmt.Errorf("failed to check for existing admin: %w", err)
		}

		// Admin does not exist, so create them
		hashedPassword, err := crypto.HashPassword(adminPassword)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		adminUser := &models.User{
			Name:         "Admin",
			Email:        adminEmail,
			PasswordHash: hashedPassword,
			Role:         models.RoleAdmin,
		}

		if err := deps.UserRepo.Insert(ctx, adminUser); err != nil {
			return fmt.Errorf("failed to insert admin user: %w", err)
		}

		log.Printf("Successfully seeded admin user: %s", adminEmail)
		return nil
	},
}