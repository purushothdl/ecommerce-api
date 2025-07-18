// internal/auth/service.go
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/utils/crypto"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo domain.UserRepository
	authRepo domain.AuthRepository
}

func NewAuthService(userRepo domain.UserRepository, authRepo domain.AuthRepository) domain.AuthService {
	return &authService{
		userRepo: userRepo,
		authRepo: authRepo,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (*models.User, *models.RefreshToken, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, nil, apperrors.ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("auth service: could not process login: %w", err)
	}

	err = crypto.CheckPasswordHash(password, user.PasswordHash)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, nil, apperrors.ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("auth service: could not process login: %w", err)
	}

	const maxSessionsPerUser = 5
	existingTokens, err := s.authRepo.GetUserRefreshTokens(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("auth service: could not check existing sessions: %w", err)
	}

	if len(existingTokens) >= maxSessionsPerUser {
		oldestToken := existingTokens[len(existingTokens)-1]
		if err := s.authRepo.RevokeRefreshTokenByID(ctx, oldestToken.ID); err != nil {
			return nil, nil, fmt.Errorf("auth service: could not revoke old session: %w", err)
		}
	}

	refreshToken, err := GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("auth service: could not generate refresh token: %w", err)
	}

	if err := s.authRepo.StoreRefreshToken(ctx, refreshToken); err != nil {
		return nil, nil, fmt.Errorf("auth service: could not store refresh token: %w", err)
	}

	return user, refreshToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, tokenPlaintext string) (*models.User, *models.RefreshToken, error) {
	token, err := s.authRepo.GetRefreshToken(ctx, tokenPlaintext)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidToken) {
			return nil, nil, apperrors.ErrInvalidToken
		}
		return nil, nil, fmt.Errorf("auth service: could not get refresh token: %w", err)
	}

	if token.ExpiresAt.Before(time.Now()) {
		return nil, nil, apperrors.ErrTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("auth service: could not fetch user: %w", err)
	}

	return user, token, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if err := s.authRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("auth service: could not revoke refresh token: %w", err)
	}
	return nil
}

func (s *authService) GetUserSessions(ctx context.Context, userID int64) ([]*models.RefreshToken, error) {
	tokens, err := s.authRepo.GetUserRefreshTokens(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth service: could not get user sessions: %w", err)
	}
	return tokens, nil
}

func (s *authService) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	if err := s.authRepo.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		return fmt.Errorf("auth service: could not revoke all user sessions: %w", err)
	}
	return nil
}

func (s *authService) RevokeUserSession(ctx context.Context, userID, sessionID int64) error {
	if err := s.authRepo.RevokeUserSessionByID(ctx, userID, sessionID); err != nil {
		if errors.Is(err, apperrors.ErrSessionNotFound) {
			return apperrors.ErrInvalidToken
		}
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
