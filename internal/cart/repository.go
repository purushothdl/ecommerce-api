// internal/cart/repository.go
package cart

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
)

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) domain.CartRepository {
	return &cartRepository{db: db}
}

// ... (implement all the interface methods)

func (r *cartRepository) GetByUserID(ctx context.Context, userID int64) (*models.Cart, error) {
	query := `SELECT id, user_id, created_at, updated_at FROM carts WHERE user_id = $1`
	var cart models.Cart
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("cart repo: get by user id: %w", err)
	}
	return &cart, nil
}

func (r *cartRepository) GetByID(ctx context.Context, cartID int64) (*models.Cart, error) {
	query := `SELECT id, user_id, created_at, updated_at FROM carts WHERE id = $1`
	var cart models.Cart
	err := r.db.QueryRowContext(ctx, query, cartID).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("cart repo: get by id: %w", err)
	}
	return &cart, nil
}

func (r *cartRepository) Create(ctx context.Context, userID *int64) (*models.Cart, error) {
	query := `INSERT INTO carts (user_id) VALUES ($1) RETURNING id, user_id, created_at, updated_at`
	var cart models.Cart
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("cart repo: create: %w", err)
	}
	return &cart, nil
}

func (r *cartRepository) AddItem(ctx context.Context, cartID int64, productID int64, quantity int) error {
    // This logic is crucial: it tries to insert, and on conflict (item already in cart), it updates the quantity.
	query := `
        INSERT INTO cart_items (cart_id, product_id, quantity)
        VALUES ($1, $2, $3)
        ON CONFLICT (cart_id, product_id)
        DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity`
	_, err := r.db.ExecContext(ctx, query, cartID, productID, quantity)
	if err != nil {
		return fmt.Errorf("cart repo: add item: %w", err)
	}

    // Also update the cart's updated_at timestamp
    _, err = r.db.ExecContext(ctx, "UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)
	return err
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, cartID int64, productID int64, quantity int) error {
	query := `UPDATE cart_items SET quantity = $1 WHERE cart_id = $2 AND product_id = $3`
	_, err := r.db.ExecContext(ctx, query, quantity, cartID, productID)
    if err != nil {
		return fmt.Errorf("cart repo: update item: %w", err)
	}

    _, err = r.db.ExecContext(ctx, "UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)
	return err
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartID int64, productID int64) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`
	_, err := r.db.ExecContext(ctx, query, cartID, productID)
    if err != nil {
		return fmt.Errorf("cart repo: remove item: %w", err)
	}

    _, err = r.db.ExecContext(ctx, "UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)
	return err
}

func (r *cartRepository) GetItemsByCartID(ctx context.Context, cartID int64) ([]models.CartItem, error) {
	query := `
        SELECT
            ci.id, ci.cart_id, ci.quantity,
            p.id, p.name, p.price, p.thumbnail, p.stock_quantity
        FROM cart_items ci
        JOIN products p ON ci.product_id = p.id
        WHERE ci.cart_id = $1`

	rows, err := r.db.QueryContext(ctx, query, cartID)
	if err != nil {
		return nil, fmt.Errorf("cart repo: get items: %w", err)
	}
	defer rows.Close()

	var items []models.CartItem
	for rows.Next() {
		var item models.CartItem
		var product models.Product
		if err := rows.Scan(
			&item.ID, &item.CartID, &item.Quantity,
			&product.ID, &product.Name, &product.Price, &product.Thumbnail, &product.StockQuantity,
		); err != nil {
			return nil, fmt.Errorf("cart repo: scan item: %w", err)
		}
		item.Product = &product
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *cartRepository) MergeCarts(ctx context.Context, fromCartID, toCartID int64) error {
	// This is a complex operation, best done in a transaction at the service level
	// The SQL logic would be to update cart_items where cart_id = fromCartID to toCartID, handling conflicts.
	query := `
        INSERT INTO cart_items (cart_id, product_id, quantity)
        SELECT $1, product_id, quantity FROM cart_items WHERE cart_id = $2
        ON CONFLICT (cart_id, product_id)
        DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity`
	_, err := r.db.ExecContext(ctx, query, toCartID, fromCartID)
	return err
}

func (r *cartRepository) Delete(ctx context.Context, cartID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM carts WHERE id = $1", cartID)
	return err
}