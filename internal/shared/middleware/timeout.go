// internal/shared/middleware/timeout.go
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/purushothdl/ecommerce-api/pkg/response"
)

// TimeoutMiddleware creates a timeout middleware with specified duration
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
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
