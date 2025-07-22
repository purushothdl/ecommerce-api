// internal/auth/service.go
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo    domain.UserRepository
	authRepo    domain.AuthRepository
	cartService domain.CartService
	jwtSecret   string
	logger      *slog.Logger
}

func NewAuthService(authRepo domain.AuthRepository, userRepo domain.UserRepository, cartService domain.CartService, jwtSecret string, logger *slog.Logger) domain.AuthService {
	return &authService{
		authRepo:    authRepo,
		userRepo:    userRepo,
		cartService: cartService,
		jwtSecret:   jwtSecret,
		logger:      logger,
	}
}

func (s *authService) LoginWithCartMerge(ctx context.Context, store domain.Store, email, password string, anonymousCartID *int64) (*models.User, *models.RefreshToken, error) {
    // First, authenticate user (outside transaction)
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, apperrors.ErrUserNotFound) {
            s.logger.Error("failed to get user by email during login", "email", email, "error", err)
            return nil, nil, apperrors.ErrInvalidCredentials
        }
        return nil, nil, fmt.Errorf("auth service: could not process login: %w", err)
    }

    err = crypto.CheckPasswordHash(password, user.PasswordHash)
    if err != nil {
        if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
            s.logger.Warn("invalid password attempt", "user_id", user.ID, "email", email)
            return nil, nil, apperrors.ErrInvalidCredentials
        }
        return nil, nil, fmt.Errorf("auth service: could not process login: %w", err)
    }

    // Now do session management and cart merge in transaction
    var refreshToken *models.RefreshToken
    
    err = store.ExecTx(ctx, func(q *domain.Queries) error {
        // 1. Manage existing sessions
        const maxSessionsPerUser = 5
        existingTokens, err := q.AuthRepo.GetUserRefreshTokens(ctx, user.ID)
        if err != nil {
            s.logger.Error("failed to get user refresh tokens", "user_id", user.ID, "error", err)
            return fmt.Errorf("auth service: could not check existing sessions: %w", err)
        }

        if len(existingTokens) >= maxSessionsPerUser {
            oldestToken := existingTokens[len(existingTokens)-1]
            s.logger.Info(
                "max sessions reached, revoking oldest token",
                "user_id", user.ID,
                "revoked_token_id", oldestToken.ID,
            )

            if err := q.AuthRepo.RevokeRefreshTokenByID(ctx, oldestToken.ID); err != nil {
                s.logger.Error("failed to revoke old session", "user_id", user.ID, "error", err)
                return fmt.Errorf("auth service: could not revoke old session: %w", err)
            }
        }

        // 2. Generate and store new refresh token
        refreshToken, err = s.GenerateRefreshToken(ctx, user.ID)
        if err != nil {
            s.logger.Error("failed to generate refresh token", "user_id", user.ID, "error", err)
            return fmt.Errorf("auth service: could not generate refresh token: %w", err)
        }

        if err := q.AuthRepo.StoreRefreshToken(ctx, refreshToken); err != nil {
            s.logger.Error("failed to store refresh token", "user_id", user.ID, "error", err)
            return fmt.Errorf("auth service: could not store refresh token: %w", err)
        }

        // 3. Handle cart merge if needed
        if anonymousCartID != nil && *anonymousCartID != 0 {
            if err := s.cartService.HandleLoginWithTransaction(ctx, q, user.ID, *anonymousCartID); err != nil {
                return fmt.Errorf("failed to merge cart: %w", err)
            }
        }

        return nil
    })

    if err != nil {
        return nil, nil, err
    }

    return user, refreshToken, nil
}


func (s *authService) RefreshToken(ctx context.Context, tokenPlaintext string) (*models.User, *models.RefreshToken, error) {
	s.logger.Info("processing refresh token request")

	token, err := s.authRepo.GetRefreshToken(ctx, tokenPlaintext)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidToken) {
			s.logger.Warn("invalid refresh token provided", "error", err)
			return nil, nil, apperrors.ErrInvalidToken
		}
		s.logger.Error("failed to get refresh token", "error", err)
		return nil, nil, fmt.Errorf("auth service: could not get refresh token: %w", err)
	}

	if token.ExpiresAt.Before(time.Now()) {
		s.logger.Warn("expired refresh token provided", "token_id", token.ID)
		return nil, nil, apperrors.ErrTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		s.logger.Error("failed to fetch user for refresh token", "user_id", token.UserID, "error", err)
		return nil, nil, fmt.Errorf("auth service: could not fetch user: %w", err)
	}

	return user, token, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	s.logger.Info("processing logout request")

	if err := s.authRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		s.logger.Error("failed to revoke refresh token", "error", err)
		return fmt.Errorf("auth service: could not revoke refresh token: %w", err)
	}

	return nil
}

func (s *authService) GetUserSessions(ctx context.Context, userID int64) ([]*models.RefreshToken, error) {
	s.logger.Info("fetching user sessions", "user_id", userID)

	tokens, err := s.authRepo.GetUserRefreshTokens(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user sessions", "user_id", userID, "error", err)
		return nil, fmt.Errorf("auth service: could not get user sessions: %w", err)
	}

	return tokens, nil
}

func (s *authService) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	s.logger.Info("revoking all user sessions", "user_id", userID)

	if err := s.authRepo.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		s.logger.Error("failed to revoke all user sessions", "user_id", userID, "error", err)
		return fmt.Errorf("auth service: could not revoke all user sessions: %w", err)
	}

	return nil
}

func (s *authService) RevokeUserSession(ctx context.Context, userID, sessionID int64) error {
	s.logger.Info("revoking user session", "user_id", userID, "session_id", sessionID)

	if err := s.authRepo.RevokeUserSessionByID(ctx, userID, sessionID); err != nil {
		if errors.Is(err, apperrors.ErrSessionNotFound) {
			s.logger.Warn("session not found", "user_id", userID, "session_id", sessionID, "error", err)
			return apperrors.ErrInvalidToken
		}
		s.logger.Error("failed to revoke user session", "user_id", userID, "session_id", sessionID, "error", err)
		return fmt.Errorf("auth service: could not revoke user session: %w", err)
	}

	return nil
}

func (s *authService) CleanupExpiredTokens(ctx context.Context) error {
	if err := s.authRepo.CleanupExpiredTokens(ctx); err != nil {
		return fmt.Errorf("auth service: could not cleanup expired tokens: %w", err)
	}
	return nil
}

func (s *authService) GenerateAccessToken(ctx context.Context, user *models.User) (string, error) {
	return GenerateAccessToken(user, s.jwtSecret)
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	return ValidateToken(tokenString, s.jwtSecret)
}

func (s *authService) GenerateRefreshToken(ctx context.Context, userID int64) (*models.RefreshToken, error) {
	token, err := GenerateRefreshToken(userID)
	if err != nil {
		s.logger.Error(
			"failed to generate refresh token",
			"user_id", userID,
			"error", err,
		)
		return nil, fmt.Errorf("auth service: %w", err)
	}
	return token, nil
}
