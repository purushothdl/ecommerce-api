// internal/user/repository.go
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/purushothdl/ecommerce-api/internal/models" 
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// Repository provides access to the user storage.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new user repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Insert inserts a new user record into the database.
func (r *Repository) Insert(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (name, email, password_hash, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.PasswordHash, user.Role}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("user repository: failed to insert user: %w", err)
	}
	return nil
}

// GetByEmail retrieves a user by their email address.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, created_at, name, email, password_hash, role, version
        FROM users
        WHERE email = $1`

	var user models.User
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("user repository: failed to get user by email: %w", err)
	}
	return &user, nil
}