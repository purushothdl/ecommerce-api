// internal/shared/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/purushothdl/ecommerce-api/internal/auth"
	usercontext "github.com/purushothdl/ecommerce-api/internal/shared/context"
	"github.com/purushothdl/ecommerce-api/pkg/response"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "authorization header missing")
				return
			}

			// Parse Bearer token
			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				response.Error(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			tokenString := headerParts[1]

			// Validate JWT token
			claims, err := auth.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, err.Error())
				return
			}

			// Extract user information from claims
			user := usercontext.UserContext{
				ID:    int64(claims["sub"].(float64)),
				Name:  claims["name"].(string),
				Email: claims["email"].(string),
				Role:  claims["role"].(string),
			}

			// Add user to request context
			ctx := usercontext.SetUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
