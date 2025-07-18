// internal/user/repository.go
package user

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
)

// Repository defines the data access methods for users
type Repository interface {
	Insert(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	StoreRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenPlaintext string) (*models.RefreshToken, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	RevokeRefreshToken(ctx context.Context, tokenPlaintext string) error
}

// repository implements the Repository interface
type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Insert(ctx context.Context, user *models.User) error {
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

func (r *repository) GetByID(ctx context.Context, id int64) (*models.User, error) {
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

func (r *repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
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

// --- New Refresh Token Methods ---

func (r *repository) StoreRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)`

	args := []any{token.UserID, token.TokenHash, token.ExpiresAt}

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("refresh token repository: failed to store token: %w", err)
	}
	return nil
}

func (r *repository) GetRefreshToken(ctx context.Context, tokenPlaintext string) (*models.RefreshToken, error) {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenPlaintext)))

	query := `
        SELECT rt.id, rt.user_id, rt.token_hash, rt.expires_at
        FROM refresh_tokens rt
        WHERE rt.token_hash = $1`

	var token models.RefreshToken

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("refresh token repository: failed to get token: %w", err)
	}

	// Delete the token after use (optional, but recommended for security)
	// go r.deleteRefreshToken(context.Background(), token.ID)

	return &token, nil
}

func (r *repository) RevokeRefreshToken(ctx context.Context, tokenPlaintext string) error {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenPlaintext)))
	query := "DELETE FROM refresh_tokens WHERE token_hash = $1"

	_, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}


func (r *repository) GetUserRefreshTokens(ctx context.Context, userID int64) ([]*models.RefreshToken, error) {
    query := `
        SELECT id, user_id, token_hash, expires_at, created_at
        FROM refresh_tokens 
        WHERE user_id = $1 AND expires_at > NOW()
        ORDER BY created_at DESC`

    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user refresh tokens: %w", err)
    }
    defer rows.Close()

    var tokens []*models.RefreshToken
    for rows.Next() {
        var token models.RefreshToken
        if err := rows.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt); err != nil {
            return nil, fmt.Errorf("failed to scan refresh token: %w", err)
        }
        tokens = append(tokens, &token)
    }

    return tokens, nil
}

func (r *repository) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
    query := "DELETE FROM refresh_tokens WHERE user_id = $1"
    _, err := r.db.ExecContext(ctx, query, userID)
    if err != nil {
        return fmt.Errorf("failed to revoke all user refresh tokens: %w", err)
    }
    return nil
}

func (r *repository) CleanupExpiredTokens(ctx context.Context) error {
    query := "DELETE FROM refresh_tokens WHERE expires_at < NOW()"
    _, err := r.db.ExecContext(ctx, query)
    return err
}
