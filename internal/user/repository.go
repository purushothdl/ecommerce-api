// internal/user/repository.go
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository returns a domain.UserRepository implementation
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Insert(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (name, email, password_hash, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.PasswordHash, user.Role}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrDuplicateEmail
		}
		return fmt.Errorf("user repository: failed to insert user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
        SELECT id, created_at, name, email, password_hash, role, version
        FROM users
        WHERE id = $1`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("user repository: failed to get user by ID: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, created_at, name, email, password_hash, role, version
        FROM users
        WHERE email = $1`

	var user models.User
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
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("user repository: failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
        UPDATE users 
        SET name = $1, email = $2, password_hash = $3, role = $4, version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING version`

	args := []any{user.Name, user.Email, user.PasswordHash, user.Role, user.ID, user.Version}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrEditConflict
		}
		return fmt.Errorf("user repository: failed to update user: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id = $1"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("user repository: failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("user repository: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	query := `
        SELECT id, created_at, name, email, password_hash, role, version
        FROM users`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("user repository: failed to get all users: %w", err)
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.Name,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("user repository: failed to scan user row: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user repository: error iterating over rows: %w", err)
	}

	return users, nil
}
