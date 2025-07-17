// auth/handler.go (The corrected version)
package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

type Handler struct {
	authService Service
	jwtSecret   string
}

func NewHandler(authService Service, jwtSecret string) *Handler {
	return &Handler{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Post("/login", h.HandleLogin)
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var input CreateTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	user, err := h.authService.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := GenerateToken(user, h.jwtSecret, 24*time.Hour)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Could not generate authentication token")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"token_type":   "Bearer",
		"access_token": token,
	})
}