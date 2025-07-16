// internal/user/handler.go
package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/internal/models"
)

// ServiceInterface defines the methods our handler needs from the user service.
type ServiceInterface interface {
	Register(ctx context.Context, name, email, password string) (*models.User, error)
}

// Handler provides HTTP handlers for user-related routes.
type Handler struct {
	service ServiceInterface
}

// NewHandler creates a new user handler.
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the user routes with a chi router.
func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Post("/users", h.HandleRegister)
}

// handleRegister handles the user registration request.
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Call the service layer to perform the business logic.
	user, err := h.service.Register(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			response.Error(w, http.StatusConflict, "Email address is already in use")
		} else {
			response.Error(w, http.StatusInternalServerError, "Could not create user")
		}
		return
	}

	response.JSON(w, http.StatusCreated, user)
}