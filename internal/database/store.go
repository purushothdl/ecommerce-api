package database

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/purushothdl/ecommerce-api/internal/auth"
    "github.com/purushothdl/ecommerce-api/internal/cart"
    "github.com/purushothdl/ecommerce-api/internal/domain"
    "github.com/purushothdl/ecommerce-api/internal/product"
    "github.com/purushothdl/ecommerce-api/internal/user"
)

// sqlStore provides all functions to execute SQL queries and transactions.
type sqlStore struct {
    db *sql.DB
}

// NewStore creates a new Store
func NewStore(db *sql.DB) domain.Store {
    return &sqlStore{
        db: db,
    }
}

// ExecTx executes a function within a database transaction.
func (s *sqlStore) ExecTx(ctx context.Context, fn func(q *domain.Queries) error) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("store: failed to begin transaction: %w", err)
    }
    defer tx.Rollback() // Rollback is a no-op if Commit succeeds.

    // Create a single Queries object, initializing all repositories with the transaction `tx`.
    q := &domain.Queries{
        UserRepo:    user.NewUserRepository(tx),
        CartRepo:    cart.NewCartRepository(tx),
        ProductRepo: product.NewProductRepository(tx),
        AuthRepo:    auth.NewAuthRepository(tx),
    }

    // Execute the callback, passing our single Queries object.
    err = fn(q)
    if err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("transaction rollback failed: %v (original error: %w)", rbErr, err)
        }
        return err 
    }

    return tx.Commit()
}
