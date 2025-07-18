package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/pkg/errors"
)

type contextKey string

const UserContextKey = contextKey("user")

func GenerateAccessToken(user *models.User, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func GenerateRefreshToken(userID int64) (*models.RefreshToken, error) {
	// Generate secure 32-byte random string
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes for token: %w", err)
	}

	// The actual token sent to the user
	tokenPlaintext := base64.URLEncoding.EncodeToString(randomBytes)

	// The hash (hex string) stored in the database
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	return &models.RefreshToken{
		UserID:    userID,
		TokenHash: fmt.Sprintf("%x", tokenHash),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Token:     tokenPlaintext,
	}, nil
}

func ValidateToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.ErrUnexpectedMethod
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrTokenExpired
		}
		return nil, apperrors.ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, apperrors.ErrInvalidToken
}
