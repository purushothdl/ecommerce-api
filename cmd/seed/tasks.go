// cmd/seed/tasks.go
package main

import (
	"context"
	"log"

	"github.com/purushothdl/ecommerce-api/internal/domain"
)

// SeederDeps holds all the dependencies that our seeder tasks might need.
type SeederDeps struct {
	UserRepo     domain.UserRepository
	ProductRepo  domain.ProductRepository
	CategoryRepo domain.CategoryRepository
}

// SeederTask defines the interface for a single seeding operation.
type SeederTask struct {
	Name string
	Run  func(ctx context.Context, deps SeederDeps) error
}

// RunSeeders executes a list of seeder tasks in order.
func RunSeeders(ctx context.Context, deps SeederDeps, tasks []SeederTask) {
	log.Println("Starting database seeding...")

	for _, task := range tasks {
		log.Printf("Running seeder task: %s...", task.Name)
		if err := task.Run(ctx, deps); err != nil {
			log.Fatalf("Task '%s' failed: %v", task.Name, err)
		}
		log.Printf("Task '%s' completed successfully.", task.Name)
	}

	log.Println("All seeder tasks completed successfully.")
}