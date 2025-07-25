// internal/shared/middleware/cors.go
package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/cors"
	"github.com/purushothdl/ecommerce-api/configs"

	"slices"
)

// This is the recommended, production-ready middleware.
func ChiCors(cfg configs.CORSConfig) func(http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowOrigins,
		AllowedMethods:   cfg.AllowMethods,
		AllowedHeaders:   cfg.AllowHeaders,
		ExposedHeaders:   cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}).Handler
}

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
		MaxAge:           86400, 
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
            
            // Always set Access-Control-Allow-Origin
            if len(cfg.AllowOrigins) > 0 {
                if cfg.AllowOrigins[0] == "*" {
                    w.Header().Set("Access-Control-Allow-Origin", "*")
                } else if slices.Contains(cfg.AllowOrigins, origin) {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                } else {
                    w.Header().Set("Access-Control-Allow-Origin", cfg.AllowOrigins[0])
                }
            }

            // Set other CORS headers
            if len(cfg.AllowMethods) > 0 {
                w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
            }

            if len(cfg.AllowHeaders) > 0 {
                w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
            }

            if len(cfg.ExposeHeaders) > 0 && cfg.ExposeHeaders[0] != "" {
                w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
            }

            if cfg.AllowCredentials {
                w.Header().Set("Access-Control-Allow-Credentials", "true")
            }

            if cfg.MaxAge > 0 {
                w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))
            }

			fmt.Printf("CORS Headers Set:\n")
			fmt.Printf("  Origin: %s\n", w.Header().Get("Access-Control-Allow-Origin"))
			fmt.Printf("  Methods: %s\n", w.Header().Get("Access-Control-Allow-Methods"))
			fmt.Printf("  Headers: %s\n", w.Header().Get("Access-Control-Allow-Headers"))
			fmt.Printf("  Credentials: %s\n", w.Header().Get("Access-Control-Allow-Credentials"))

            // Handle preflight requests
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent) 
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

func DevCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-CSRF-Token")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Max-Age", "86400")
        w.Header().Set("Vary", "Origin")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}
