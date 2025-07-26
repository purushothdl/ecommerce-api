// cmd/seed/main.go
package main

import (
	"context"
	"log"

	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/category"
	"github.com/purushothdl/ecommerce-api/internal/database"
	"github.com/purushothdl/ecommerce-api/internal/product"
	"github.com/purushothdl/ecommerce-api/internal/user"
)

func main() {
	// 1. Load configuration
	cfg, err := configs.LoadConfig("api.env")
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// 2. Connect to the database
	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// 3. Create repository dependencies
	deps := SeederDeps{
		UserRepo:     user.NewUserRepository(db),
		CategoryRepo: category.NewCategoryRepository(db),
		ProductRepo:  product.NewProductRepository(db),
	}

	// 4. Define the list of tasks to run
	// The order here matters!
	tasks := []SeederTask{
		SeedAdminsTask,
		SeedProductsTask,
	}

	// 5. Run the seeder
	RunSeeders(context.Background(), deps, tasks)
}