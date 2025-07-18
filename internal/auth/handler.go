// auth/handler.go (The corrected version)
package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	usercontext "github.com/purushothdl/ecommerce-api/internal/shared/context"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

type Handler struct {
	authService domain.AuthService
	jwtSecret   string
}

func NewHandler(authService domain.AuthService, jwtSecret string) *Handler {
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

	// Set refresh token as HTTP-only cookie for security
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken.Token,
		HttpOnly: true,
		Secure:   false, // true in prod
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
	})

	response.JSON(w, http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}

func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "No refresh token provided")
		return
	}

	user, _, err := h.authService.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

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
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "No refresh token provided")
		return
	}

	if err := h.authService.Logout(r.Context(), cookie.Value); err != nil {
		response.Error(w, http.StatusInternalServerError, "Could not revoke token")
		return
	}

	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	response.JSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// HandleGetSessions returns all active sessions for the authenticated user
func (h *Handler) HandleGetSessions(w http.ResponseWriter, r *http.Request) {
	user, err := usercontext.GetUser(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessions, err := h.authService.GetUserSessions(r.Context(), user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not retrieve sessions")
		return
	}

	// Convert to safe response format (no sensitive data)
	sessionInfos := make([]map[string]any, len(sessions))
	for i, session := range sessions {
		sessionInfos[i] = map[string]any{
			"id":         session.ID,
			"created_at": session.CreatedAt,
			"expires_at": session.ExpiresAt,
			"is_active":  session.ExpiresAt.After(time.Now()),
		}
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"sessions": sessionInfos,
		"count":    len(sessionInfos),
	})
}

// HandleLogoutAllDevices logs out the user from all devices
func (h *Handler) HandleLogoutAllDevices(w http.ResponseWriter, r *http.Request) {
	user, err := usercontext.GetUser(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.authService.RevokeAllUserSessions(r.Context(), user.ID); err != nil {
		response.Error(w, http.StatusInternalServerError, "could not revoke all sessions")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "logged out from all devices successfully",
	})
}

// HandleLogoutSpecificDevice logs out from a specific device/session
func (h *Handler) HandleLogoutSpecificDevice(w http.ResponseWriter, r *http.Request) {
	user, err := usercontext.GetUser(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "sessionId")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	if err := h.authService.RevokeUserSession(r.Context(), user.ID, sessionID); err != nil {
		if errors.Is(err, apperrors.ErrInvalidToken) {
			response.Error(w, http.StatusNotFound, "session not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not revoke session")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "session revoked successfully",
	})
}