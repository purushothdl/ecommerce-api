// internal/category/repository.go
package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
)

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	query := `INSERT INTO categories (name) VALUES ($1) RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, category.Name).Scan(&category.ID, &category.CreatedAt)
	if err != nil {
		// Note: A more robust implementation would check for specific DB errors like unique violation
		return fmt.Errorf("category repository: failed to create category: %w", err)
	}
	return nil
}

func (r *categoryRepository) GetByName(ctx context.Context, name string) (*models.Category, error) {
	query := `SELECT id, name, created_at FROM categories WHERE name = $1`
	var cat models.Category
	err := r.db.QueryRowContext(ctx, query, name).Scan(&cat.ID, &cat.Name, &cat.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound 
		}
		return nil, fmt.Errorf("category repository: failed to get category by name: %w", err)
	}
	return &cat, nil
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]*models.Category, error) {
	query := `SELECT id, name, created_at, updated_at FROM categories ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("category repository: failed to get all categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, fmt.Errorf("category repository: failed to scan category row: %w", err)
		}
		categories = append(categories, &cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("category repository: error iterating rows: %w", err)
	}
	return categories, nil
}