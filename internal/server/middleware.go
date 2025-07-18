// internal/server/middleware.go
package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/purushothdl/ecommerce-api/internal/auth"
	usercontext "github.com/purushothdl/ecommerce-api/internal/context"
	"github.com/purushothdl/ecommerce-api/pkg/response"
)

// Timeout middleware factory
func (s *Server) timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Channel to signal completion
			done := make(chan struct{})
			
			go func() {
				defer close(done)
				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			select {
			case <-done:
				// Request completed successfully
				return
			case <-ctx.Done():
				// Timeout occurred
				if ctx.Err() == context.DeadlineExceeded {
					response.Error(w, http.StatusGatewayTimeout, "request timeout")
				} else {
					response.Error(w, http.StatusRequestTimeout, "request cancelled")
				}
				return
			}
		})
	}
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "authorization header missing")
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			response.Error(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		tokenString := headerParts[1]

		claims, err := auth.ValidateToken(tokenString, s.config.JWT.Secret)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, err.Error())
			return
		}

		// Use the shared context package
		user := usercontext.UserContext{
			ID:    int64(claims["sub"].(float64)),
			Name:  claims["name"].(string),
			Email: claims["email"].(string),
			Role:  claims["role"].(string),
		}

		ctx := usercontext.SetUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
