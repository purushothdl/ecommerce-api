// internal/user/handler.go
package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/auth"
	"github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Post("/users", h.HandleRegister)
	r.Get("/profile", h.HandleGetProfile)
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var input CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	v := validator.New()

	if input.Validate(v); !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}
	user, err := h.service.Register(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicateEmail) {
			response.Error(w, http.StatusConflict, "Email address is already in use")
		} else {
			response.Error(w, http.StatusInternalServerError, "Could not create user")
		}
		return
	}

	response.JSON(w, http.StatusCreated, user)
}

func (h *Handler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	// Add a 6 second delay
	time.Sleep(6 * time.Second)

	// Retrieve user details from the context
	userCtx, ok := r.Context().Value(auth.UserContextKey).(struct {
		ID    int64
		Name  string
		Email string
		Role  string
	})
	if !ok {
		response.Error(w, http.StatusInternalServerError, "error retrieving user from context")
		return
	}

	// Return user details
	response.JSON(w, http.StatusOK, map[string]any{
		"message": "Welcome to your protected profile!",
		"user_id": userCtx.ID,
		"name":    userCtx.Name,
		"email":   userCtx.Email,
		"role":    userCtx.Role,
	})
}
