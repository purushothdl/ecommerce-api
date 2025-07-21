// internal/product/repository.go
package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
)

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, p *models.Product) error {
	query := `INSERT INTO products (name, description, price, stock_quantity, category_id, brand, sku, images, thumbnail, dimensions, warranty_information)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
              RETURNING id, created_at, updated_at, version`
	args := []any{
		p.Name, p.Description, p.Price, p.StockQuantity, p.CategoryID,
		p.Brand, p.SKU, p.Images, p.Thumbnail, p.Dimensions, p.WarrantyInformation,
	}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt, &p.Version)
	if err != nil {
		return fmt.Errorf("product repository: failed to create product: %w", err)
	}
	return nil
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `
        SELECT p.id, p.name, p.description, p.price, p.stock_quantity, p.category_id, p.brand, p.sku, 
               p.images, p.thumbnail, p.dimensions, p.warranty_information, p.created_at, p.updated_at, p.version,
               c.name as category_name
        FROM products p
        LEFT JOIN categories c ON p.category_id = c.id
        WHERE p.id = $1`

	var p models.Product
	var cat models.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.StockQuantity, &p.CategoryID, &p.Brand, &p.SKU,
		&p.Images, &p.Thumbnail, &p.Dimensions, &p.WarrantyInformation, &p.CreatedAt, &p.UpdatedAt, &p.Version,
		&cat.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("product repository: failed to get product by ID: %w", err)
	}

	cat.ID = p.CategoryID
	p.Category = &cat
	return &p, nil
}

func (r *productRepository) GetAll(ctx context.Context, filters domain.ProductFilters) ([]*models.Product, error) {
	// Query builder for dynamic filtering and pagination
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
        SELECT p.id, p.name, p.description, p.price, p.stock_quantity, p.category_id, p.brand, p.thumbnail,
               c.name as category_name
        FROM products p
        LEFT JOIN categories c ON p.category_id = c.id
    `)

	var args []any
	var conditions []string
	argCount := 1

	if filters.Category != "" {
		conditions = append(conditions, fmt.Sprintf("c.name = $%d", argCount))
		args = append(args, filters.Category)
		argCount++
	}

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY p.id ASC LIMIT $%d OFFSET $%d", argCount, argCount+1))
	args = append(args, filters.PageSize, (filters.Page-1)*filters.PageSize)

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("product repository: query failed: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var p models.Product
		var cat models.Category
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.StockQuantity, &p.CategoryID, &p.Brand, &p.Thumbnail,
			&cat.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("product repository: failed to scan row: %w", err)
		}
		cat.ID = p.CategoryID
		p.Category = &cat
		products = append(products, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("product repository: error iterating rows: %w", err)
	}

	return products, nil
}