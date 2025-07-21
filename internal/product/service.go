// internal/product/service.go
package product

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
)

type productService struct {
	repo   domain.ProductRepository
	logger *slog.Logger
}

func NewProductService(repo domain.ProductRepository, logger *slog.Logger) domain.ProductService {
	return &productService{repo: repo, logger: logger}
}

func (s *productService) ListProducts(ctx context.Context, filters domain.ProductFilters) ([]*models.Product, error) {
	// Set default pagination if not provided
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 10
	}

	products, err := s.repo.GetAll(ctx, filters)
	if err != nil {
		s.logger.Error("failed to list products from repository", "error", err)
		return nil, fmt.Errorf("product service: could not retrieve products: %w", err)
	}
	return products, nil
}

func (s *productService) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to get product from repository", "product_id", id, "error", err)
		return nil, fmt.Errorf("product service: could not retrieve product: %w", err)
	}
	return product, nil
}