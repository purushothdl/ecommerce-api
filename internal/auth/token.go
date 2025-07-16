// internal/auth/token.go
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/purushothdl/ecommerce-api/internal/models"
)

// Define a custom type for our context key to avoid collisions.
type contextKey string
const UserContextKey = contextKey("user")

var (
	ErrInvalidToken     = errors.New("token is invalid")
	ErrTokenExpired     = errors.New("token has expired")
	ErrUnexpectedMethod = errors.New("unexpected signing method")
)


// GenerateToken generates a new JWT for a given user.
func GenerateToken(user *models.User, secret string, duration time.Duration) (string, error) {
	// Create the token claims
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"name": user.Name,
		"role": user.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(duration).Unix(),
	}

	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT string and returns the claims as a map.
func ValidateToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Ensure the signing method is what we expect.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedMethod
		}
		return []byte(secret), nil
	})

	if err != nil {
		// Handle specific JWT errors.
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, ErrInvalidToken
}