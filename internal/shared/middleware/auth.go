// internal/shared/middleware/auth.go
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/internal/shared/context"
	"github.com/purushothdl/ecommerce-api/pkg/response"
)

// extractAndSetUser is a shared function that extracts user from JWT and sets it in context
func extractAndSetUser(r *http.Request, jwtSecret string) (*http.Request, error) {
    // Extract token from Authorization header
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return r, errors.New("authorization header missing")
    }

    // Parse Bearer token
    headerParts := strings.Split(authHeader, " ")
    if len(headerParts) != 2 || headerParts[0] != "Bearer" {
        return r, errors.New("invalid authorization header format")
    }

    tokenString := headerParts[1]

    // Validate JWT token
    claims, err := auth.ValidateToken(tokenString, jwtSecret)
    if err != nil {
        return r, err
    }

    // Extract user information from claims
    user := context.UserContext{
        ID:    int64(claims["sub"].(float64)),
        Name:  claims["name"].(string),
        Email: claims["email"].(string),
        Role:  claims["role"].(string),
    }

    // Set user in context
    ctx := context.SetUser(r.Context(), user)
    return r.WithContext(ctx), nil
}

// AuthMiddleware creates authentication middleware that REQUIRES authentication
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            newRequest, err := extractAndSetUser(r, jwtSecret)
            if err != nil {
                response.Error(w, http.StatusUnauthorized, err.Error())
                return
            }
            
            next.ServeHTTP(w, newRequest)
        })
    }
}

// OptionalAuthMiddleware extracts user from JWT if present, but continues if missing/invalid
func OptionalAuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            newRequest, err := extractAndSetUser(r, jwtSecret)
            if err != nil {
                // If auth fails, continue as anonymous user with original request
                next.ServeHTTP(w, r)
                return
            }
            
            // If auth succeeds, continue with user context
            next.ServeHTTP(w, newRequest)
        })
    }
}

// AdminMiddleware remains the same
func AdminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, err := context.GetUser(r.Context())
        if err != nil {
            response.Error(w, http.StatusUnauthorized, "unauthorized")
            return
        }

        if user.Role != "admin" {
            response.Error(w, http.StatusForbidden, "access forbidden: admin rights required")
            return
        }

        next.ServeHTTP(w, r)
    })
}
