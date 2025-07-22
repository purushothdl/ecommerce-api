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
	db domain.DBTX
}

func NewCartRepository(db domain.DBTX) domain.CartRepository {
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

func (r *cartRepository) Delete(ctx context.Context, cartID int64) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM carts WHERE id = $1", cartID)
    return err
}

func (r *cartRepository) AddItem(ctx context.Context, cartID int64, productID int64, quantity int) error {
    // Updated to handle timestamps properly
    query := `
        INSERT INTO cart_items (cart_id, product_id, quantity, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        ON CONFLICT (cart_id, product_id)
        DO UPDATE SET 
            quantity = cart_items.quantity + EXCLUDED.quantity,
            updated_at = NOW()`  
    
    _, err := r.db.ExecContext(ctx, query, cartID, productID, quantity)
    if err != nil {
        return fmt.Errorf("cart repo: add item: %w", err)
    }

    // Update cart's updated_at timestamp
    _, err = r.db.ExecContext(ctx, "UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)
    return err
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, cartID int64, productID int64, quantity int) error {
    query := `
        UPDATE cart_items 
        SET quantity = $1, updated_at = NOW() 
        WHERE cart_id = $2 AND product_id = $3`
    
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
            ci.id, ci.cart_id, ci.quantity, ci.created_at, ci.updated_at,
            p.id, p.name, p.price, p.thumbnail, p.stock_quantity
        FROM cart_items ci
        JOIN products p ON ci.product_id = p.id
        WHERE ci.cart_id = $1
        ORDER BY ci.created_at DESC`  

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
            &item.ID, &item.CartID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt,
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
    query := `
        INSERT INTO cart_items (cart_id, product_id, quantity, created_at, updated_at)
        SELECT $1, product_id, quantity, created_at, NOW() FROM cart_items WHERE cart_id = $2
        ON CONFLICT (cart_id, product_id)
        DO UPDATE SET 
            quantity = cart_items.quantity + EXCLUDED.quantity,
            updated_at = NOW()`  
    
    _, err := r.db.ExecContext(ctx, query, toCartID, fromCartID)
    return err
}

func (r *cartRepository) CleanupOldAnonymousCartItems(ctx context.Context, olderThanDays int) error {
    query := `
        DELETE FROM cart_items 
        WHERE cart_id IN (
            SELECT id FROM carts WHERE user_id IS NULL
        ) 
        AND created_at < NOW() - INTERVAL '%d days'`
    
    _, err := r.db.ExecContext(ctx, fmt.Sprintf(query, olderThanDays))
    return err
}