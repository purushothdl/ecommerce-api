// internal/auth/repository.go
package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
)

type authRepository struct {
	db domain.DBTX
}

func NewAuthRepository(db domain.DBTX) domain.AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) StoreRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)
        RETURNING id, created_at`

	args := []any{token.UserID, token.TokenHash, token.ExpiresAt}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return fmt.Errorf("auth repository: failed to store refresh token: %w", err)
	}
	return nil
}

func (r *authRepository) GetRefreshToken(ctx context.Context, tokenPlaintext string) (*models.RefreshToken, error) {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenPlaintext)))

	query := `
        SELECT id, user_id, token_hash, expires_at, created_at
        FROM refresh_tokens
        WHERE token_hash = $1 AND expires_at > NOW()`

	var token models.RefreshToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrInvalidToken
		}
		return nil, fmt.Errorf("auth repository: failed to get refresh token: %w", err)
	}

	token.Token = tokenPlaintext
	return &token, nil
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, tokenPlaintext string) error {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenPlaintext)))
	query := "DELETE FROM refresh_tokens WHERE token_hash = $1"

	result, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("auth repository: failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("auth repository: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrInvalidToken
	}

	return nil
}

func (r *authRepository) GetUserRefreshTokens(ctx context.Context, userID int64) ([]*models.RefreshToken, error) {
	query := `
        SELECT id, user_id, token_hash, expires_at, created_at
        FROM refresh_tokens 
        WHERE user_id = $1 AND expires_at > NOW()
        ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("auth repository: failed to get user refresh tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*models.RefreshToken
	for rows.Next() {
		var token models.RefreshToken
		if err := rows.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt); err != nil {
			return nil, fmt.Errorf("auth repository: failed to scan refresh token: %w", err)
		}
		tokens = append(tokens, &token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("auth repository: error iterating rows: %w", err)
	}

	return tokens, nil
}

func (r *authRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	query := "DELETE FROM refresh_tokens WHERE user_id = $1"
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("auth repository: failed to revoke all user refresh tokens: %w", err)
	}
	return nil
}

func (r *authRepository) RevokeRefreshTokenByID(ctx context.Context, tokenID int64) error {
	query := "DELETE FROM refresh_tokens WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("auth repository: failed to revoke refresh token by ID: %w", err)
	}
	return nil
}

func (r *authRepository) RevokeUserSessionByID(ctx context.Context, userID, sessionID int64) error {
	query := "DELETE FROM refresh_tokens WHERE id = $1 AND user_id = $2"
	result, err := r.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("auth repository: failed to revoke user session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("auth repository: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrSessionNotFound
	}

	return nil
}

func (r *authRepository) CleanupExpiredTokens(ctx context.Context) error {
	query := "DELETE FROM refresh_tokens WHERE expires_at < NOW()"
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("auth repository: failed to cleanup expired tokens: %w", err)
	}

	if rowsAffected, err := result.RowsAffected(); err == nil {
		fmt.Printf("Cleaned up %d expired refresh tokens\n", rowsAffected)
	}

	return nil
}
