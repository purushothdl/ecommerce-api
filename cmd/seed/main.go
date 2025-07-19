package main

import (
	"context"
	"log"
	"os"

	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/database"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/internal/user"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
)

func main() {
	log.Println("Starting database seeding...")

	// 1. Load configuration
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// 2. Connect to the database
	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// 3. Read admin credentials from environment variables
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" || adminPassword == "" {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD environment variables must be set")
	}

	// 4. Check if admin already exists
	userRepo := user.NewUserRepository(db)
	existingAdmin, err := userRepo.GetByEmail(context.Background(), adminEmail)
	if err == nil && existingAdmin != nil {
		log.Printf("Admin user with email %s already exists. Seeding skipped.", adminEmail)
		return
	}

	// 5. Hash the password
	hashedPassword, err := crypto.HashPassword(adminPassword)
	if err != nil {
		log.Fatalf("failed to hash admin password: %v", err)
	}

	// 6. Create the admin user model
	adminUser := &models.User{
		Name:         "Admin",
		Email:        adminEmail,
		PasswordHash: hashedPassword,
		Role:         "admin", 
	}

	// 7. Insert the admin user into the database
	if err := userRepo.Insert(context.Background(), adminUser); err != nil {
		log.Fatalf("failed to insert admin user: %v", err)
	}

	log.Printf("Successfully seeded database with admin user: %s", adminEmail)
}