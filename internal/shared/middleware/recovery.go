// internal/shared/middleware/recovery.go
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/purushothdl/ecommerce-api/pkg/response"
)

// RecoveryMiddleware creates panic recovery middleware
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic
					logger.Error("panic recovered",
						"error", err,
						"method", r.Method,
						"path", r.URL.Path,
						"stack", string(debug.Stack()),
					)

					// Return internal server error
					response.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddlewareWithCustomHandler creates recovery middleware with custom panic handler
func RecoveryMiddlewareWithCustomHandler(logger *slog.Logger, handler func(any, http.ResponseWriter, *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						"error", err,
						"method", r.Method,
						"path", r.URL.Path,
						"stack", string(debug.Stack()),
					)

					handler(err, w, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
