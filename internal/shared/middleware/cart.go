// internal/shared/middleware/cart.go
package middleware

import (
	"net/http"
	"strconv"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/shared/context"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/web"
)

// CartMiddleware manages cart creation for authenticated or anonymous users.
func CartMiddleware(cartSvc domain.CartService, isProduction bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID if authenticated
			var userID *int64
			if user, err := context.GetUser(r.Context()); err == nil {
				userID = &user.ID
			}

			// Check for existing anonymous cart cookie
			var anonymousCartID *int64
			if cookie, err := r.Cookie(web.CartIDCookieName); err == nil {
				if id, err := strconv.ParseInt(cookie.Value, 10, 64); err == nil {
					anonymousCartID = &id
				}
			}

			// Fetch or create cart
			cart, err := cartSvc.GetOrCreateCart(r.Context(), userID, anonymousCartID)
			if err != nil {
				response.Error(w, http.StatusInternalServerError, "failed to load cart")
				return
			}

			// Set cookie for new anonymous carts
			if userID == nil && anonymousCartID == nil {
				web.SetCartCookie(w, strconv.FormatInt(cart.ID, 10), isProduction)
			}

			// Inject cart into context
			ctx := context.SetCart(r.Context(), context.CartContext{
				ID:     cart.ID,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
