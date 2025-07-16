// internal/server/middleware.go
package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/pkg/response"
)

// authMiddleware creates a new authentication middleware.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the Authorization header.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "authorization header missing")
			return
		}

		// 2. Validate the header format: "Bearer <token>".
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			response.Error(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		tokenString := headerParts[1]

		// 3. Validate the token.
		claims, err := auth.ValidateToken(tokenString, s.config.JWT.Secret)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, err.Error())
			return
		}

		// 4. Extract user info and add it to the request context.
		// We'll create a simplified User struct for the context.
		user := struct {
			ID   int64
			Role string
		}{
			ID:   int64(claims["sub"].(float64)), // JWT numbers are decoded as float64
			Role: claims["role"].(string),
		}

		// Create a new context with the user value.
		ctx := context.WithValue(r.Context(), auth.UserContextKey, user)

		// 5. Call the next handler in the chain with the new context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}