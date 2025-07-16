// internal/auth/handler.go
package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/pkg/response"
)

// UserServiceInterface defines the methods our handler needs from the user service.
// This is the key to breaking the dependency cycle.
type UserServiceInterface interface {
	Login(ctx context.Context, email, password string) (*models.User, error)
}

// Handler provides the HTTP handler for creating authentication tokens.
type Handler struct {
	userService UserServiceInterface
	jwtSecret   string
}

// NewHandler creates a new auth handler.
func NewHandler(userService UserServiceInterface, jwtSecret string) *Handler {
	return &Handler{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

// RegisterRoutes registers the auth routes.
func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Post("/login", h.HandleCreateToken)
}

// handleCreateToken is the HTTP handler for the login endpoint.
func (h *Handler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// 1. Call the user service to validate credentials.
	user, err := h.userService.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		// The user service already returns a generic error for security.
		response.Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// 2. Credentials are valid. Generate a JWT.
	token, err := GenerateToken(user, h.jwtSecret, 24*time.Hour)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Could not generate authentication token")
		return
	}

	// 3. Send the token back to the client with the token type.
	response.JSON(w, http.StatusOK, map[string]string{
		"token_type":   "Bearer",
		"access_token": token,
	})
}