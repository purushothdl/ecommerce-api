// auth/handler.go (The corrected version)
package auth

import (
	"encoding/json"
	"net/http"

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

	user, refreshToken, err := h.authService.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	accessToken, err := GenerateAccessToken(user, h.jwtSecret)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Could not generate authentication token")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken.Token,
	})
}

// New handler for refreshing tokens
func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var input RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	user, _, err := h.authService.RefreshToken(r.Context(), input.RefreshToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// If refresh token is valid, issue a new access token
	accessToken, err := GenerateAccessToken(user, h.jwtSecret)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Could not generate new access token")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var input RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.authService.Logout(r.Context(), input.RefreshToken); err != nil {
		response.Error(w, http.StatusInternalServerError, "Could not revoke token")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
