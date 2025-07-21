// internal/category/service.go
package category

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
)

type categoryService struct {
	repo   domain.CategoryRepository
	logger *slog.Logger
}

func NewCategoryService(repo domain.CategoryRepository, logger *slog.Logger) domain.CategoryService {
	return &categoryService{repo: repo, logger: logger}
}

// GetOrCreate is a perfect utility for our future seeder.
func (s *categoryService) GetOrCreate(ctx context.Context, name string) (*models.Category, error) {
	cat, err := s.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			s.logger.Info("category not found, creating new one", "name", name)
			newCat := &models.Category{Name: name}
			if createErr := s.repo.Create(ctx, newCat); createErr != nil {
				return nil, fmt.Errorf("category service: failed to create category: %w", createErr)
			}
			return newCat, nil
		}
		return nil, fmt.Errorf("category service: failed to get category: %w", err)
	}
	return cat, nil
}

func (s *categoryService) ListCategories(ctx context.Context) ([]*models.Category, error) {
	categories, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to list categories", "error", err)
		return nil, fmt.Errorf("category service: could not retrieve categories: %w", err)
	}
	return categories, nil
}