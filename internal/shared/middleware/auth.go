// internal/shared/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/shared/context"
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
			user := context.UserContext{
				ID:    int64(claims["sub"].(float64)),
				Name:  claims["name"].(string),
				Email: claims["email"].(string),
				Role:  claims["role"].(string),
			}
			ctx := context.SetUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (populated by AuthMiddleware)
		user, err := context.GetUser(r.Context())
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Check if the user role is 'admin'
		if user.Role != "admin" {
			response.Error(w, http.StatusForbidden, "access forbidden: admin rights required")
			return
		}

		// If user is an admin, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// OptionalAuthMiddleware extracts user from JWT if present, but doesn't require authentication
func OptionalAuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Try to get token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                // No auth header, continue as anonymous user
                next.ServeHTTP(w, r)
                return
            }

            // Extract token from "Bearer <token>"
            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                // Invalid auth header format, continue as anonymous
                next.ServeHTTP(w, r)
                return
            }

            tokenString := parts[1]
            
            // Validate token (use your existing validation logic)
            user, err := auth.ValidateToken(tokenString, jwtSecret)
            if err != nil {
                // Invalid token, continue as anonymous
                next.ServeHTTP(w, r)
                return
            }

            // Valid token - set user in context
            ctx := context.SetUser(r.Context(), context.UserContext{
				ID:    int64(user["sub"].(float64)),
				Name:  user["name"].(string),
				Email: user["email"].(string),
				Role:  user["role"].(string),
            })

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
