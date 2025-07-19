// internal/shared/middleware/cors.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/purushothdl/ecommerce-api/configs"

	"slices"
)


// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() configs.CORSConfig {
	return configs.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
			http.MethodPatch,
		},
		AllowHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		ExposeHeaders:    []string{},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware creates CORS middleware with configuration
func CORSMiddleware(config ...configs.CORSConfig) func(http.Handler) http.Handler {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Set CORS headers
			if len(cfg.AllowOrigins) > 0 {
				if cfg.AllowOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					if slices.Contains(cfg.AllowOrigins, origin) {
							w.Header().Set("Access-Control-Allow-Origin", origin)
						}
				}
			}

			if len(cfg.AllowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
			}

			if len(cfg.AllowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
			}

			if len(cfg.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
			}

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if cfg.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", string(rune(cfg.MaxAge)))
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
