// internal/shared/middleware/auth.go
package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/purushothdl/ecommerce-api/internal/auth"
	serverContext "github.com/purushothdl/ecommerce-api/internal/shared/context"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"google.golang.org/api/idtoken"
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
    user := serverContext.UserContext{
        ID:    int64(claims["sub"].(float64)),
        Name:  claims["name"].(string),
        Email: claims["email"].(string),
        Role:  claims["role"].(string),
    }

    // Set user in context
    ctx := serverContext.SetUser(r.Context(), user)
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
        user, err := serverContext.GetUser(r.Context())
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

type idTokenPayloadKey struct{}

// OIDCAuthMiddleware is a constructor that returns a middleware for validating Google-issued OIDC tokens.
// It takes the expected audience (the public URL of the service being protected) as configuration.
func OIDCAuthMiddleware(audience string) func(http.Handler) http.Handler {
	// This is the outer function that accepts the configuration.
	// It returns the actual middleware handler.
	return func(next http.Handler) http.Handler {
		// This is the http.HandlerFunc that will be executed for each request.
		// It has access to the 'audience' variable from the outer scope (this is a "closure").
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if audience == "" {
				response.Error(w, http.StatusInternalServerError, "Internal auth audience not configured")
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "Missing Authorization header")
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				response.Error(w, http.StatusUnauthorized, "Invalid Authorization header format")
				return
			}
			idToken := tokenParts[1]

			// Validate the token against the configured audience
			payload, err := idtoken.Validate(r.Context(), idToken, audience)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, fmt.Sprintf("Invalid token: %v", err))
				return
			}
			
			// Optional: you could add further checks here on the payload claims if needed.
			// For example, ensuring the token was issued by the worker's service account.

			// If validation succeeds, store the validated payload in the request context
			// in case downstream handlers need information from it (like the issuer email).
			ctx := context.WithValue(r.Context(), idTokenPayloadKey{}, payload)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}