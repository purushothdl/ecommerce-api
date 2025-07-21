// internal/user/handler.go
package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	usercontext "github.com/purushothdl/ecommerce-api/internal/shared/context"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
	"github.com/purushothdl/ecommerce-api/pkg/web"
)

type Handler struct {
	userService  domain.UserService
	authService  domain.AuthService
	cartService  domain.CartService
	jwtSecret    string
	isProduction bool
	logger       *slog.Logger
}

func NewHandler(userService domain.UserService, authService domain.AuthService, cartService domain.CartService, jwtSecret string, isProduction bool, logger *slog.Logger) *Handler {
	return &Handler{
		userService:  userService,
		authService:  authService,
		cartService:  cartService,
		jwtSecret:    jwtSecret,
		isProduction: isProduction,
		logger:       logger,
	}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("register user request received")
	var input CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request payload", "error", err)
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		h.logger.Warn("validation failed", "errors", v.Errors)
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	// Create user account
	user, err := h.userService.Register(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicateEmail) {
			h.logger.Warn("duplicate email", "email", input.Email)
			response.Error(w, http.StatusConflict, "Email address is already in use")
		} else {
			h.logger.Error("failed to register user", "error", err)
			response.Error(w, http.StatusInternalServerError, "Could not create user")
		}
		return
	}

	// Merge anonymous cart if exists
	var anonymousCartID int64
	if cookie, err := r.Cookie(web.CartIDCookieName); err == nil {
		if id, parseErr := strconv.ParseInt(cookie.Value, 10, 64); parseErr == nil {
			anonymousCartID = id
		} else {
			h.logger.Warn("failed to parse cart_id cookie", "cookie_value", cookie.Value, "error", parseErr)
		}
	}

	if anonymousCartID != 0 {
		if err := h.cartService.HandleLogin(r.Context(), user.ID, anonymousCartID); err != nil {
			h.logger.Error("failed to merge cart", "user_id", user.ID, "anonymous_cart_id", anonymousCartID, "error", err)
		} else {
			web.ClearCookie(w, web.CartIDCookieName, h.isProduction)
			h.logger.Info("successfully merged cart", "user_id", user.ID, "anonymous_cart_id", anonymousCartID)
		}
	}

	// Auto-login after registration
	refreshToken, err := h.authService.GenerateRefreshToken(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("failed to create refresh token", "user_id", user.ID, "error", err)
		response.JSON(w, http.StatusCreated, dto.NewUserResponse(user))
		return
	}

	accessToken, err := h.authService.GenerateAccessToken(r.Context(), user)
	if err != nil {
		h.logger.Error("failed to generate access token", "user_id", user.ID, "error", err)
		response.JSON(w, http.StatusCreated, dto.NewUserResponse(user))
		return
	}

	web.SetRefreshTokenCookie(w, refreshToken.Token, h.isProduction)

	payload := LoginResponse{
		User:        dto.NewUserResponse(user),
		AccessToken: accessToken,
	}

	response.JSON(w, http.StatusCreated, payload)
	h.logger.Info("user registered and logged in", "user_id", user.ID, "email", user.Email)
}

func (h *Handler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get profile request received")
	userCtx, err := usercontext.GetUser(r.Context())
	if err != nil {
		h.logger.Error("failed to retrieve user from context", "error", err)
		response.Error(w, http.StatusInternalServerError, "error retrieving user from context")
		return
	}

	// Fetch the full user model
	user, err := h.userService.GetProfile(r.Context(), userCtx.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			h.logger.Warn("user not found", "user_id", userCtx.ID)
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("failed to retrieve profile", "user_id", userCtx.ID, "error", err)
		response.Error(w, http.StatusInternalServerError, "could not retrieve profile")
		return
	}

	// Create the structured response
	response.JSON(w, http.StatusOK, dto.NewUserResponse(user))
	h.logger.Info("profile retrieved successfully", "user_id", user.ID)
}

func (h *Handler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update profile request received")
	user, err := usercontext.GetUser(r.Context())
	if err != nil {
		h.logger.Warn("unauthorized access", "error", err)
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request payload", "error", err)
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		h.logger.Warn("validation failed", "errors", v.Errors)
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	updatedUser, err := h.userService.UpdateProfile(r.Context(), user.ID, input.Name, input.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			h.logger.Warn("user not found", "user_id", user.ID)
			response.Error(w, http.StatusNotFound, "user not found")
		} else if errors.Is(err, apperrors.ErrDuplicateEmail) {
			h.logger.Warn("duplicate email", "email", input.Email)
			response.Error(w, http.StatusConflict, "email address is already in use")
		} else {
			h.logger.Error("failed to update profile", "user_id", user.ID, "error", err)
			response.Error(w, http.StatusInternalServerError, "could not update profile")
		}
		return
	}

	response.JSON(w, http.StatusOK, dto.NewUserResponse(updatedUser))
	h.logger.Info("profile updated successfully", "user_id", user.ID)
}

func (h *Handler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("change password request received")
	user, err := usercontext.GetUser(r.Context())
	if err != nil {
		h.logger.Warn("unauthorized access", "error", err)
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request payload", "error", err)
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		h.logger.Warn("validation failed", "errors", v.Errors)
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	if err := h.userService.ChangePassword(r.Context(), user.ID, input.CurrentPassword, input.NewPassword); err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "current password is incorrect")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not change password")
		return
	}

	// Log out from all devices after password change for security
	if err := h.authService.RevokeAllUserSessions(r.Context(), user.ID); err != nil {
		// Log the error but don't fail the request
		h.logger.Error("Failed to revoke all user sessions", "error", err, "userID", user.ID)
	}

	resp := response.MessageResponse{Message: "Password changed successfully. Please log in again."}
	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("delete account request received")
	user, err := usercontext.GetUser(r.Context())
	if err != nil {
		h.logger.Warn("unauthorized access", "error", err)
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input DeleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request payload", "error", err)
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		h.logger.Warn("validation failed", "errors", v.Errors)
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	if err := h.userService.DeleteAccount(r.Context(), user.ID, input.Password); err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			h.logger.Warn("invalid password for account deletion", "user_id", user.ID)
			response.Error(w, http.StatusUnauthorized, "password is incorrect")
			return
		}
		h.logger.Error("failed to delete account", "user_id", user.ID, "error", err)
		response.Error(w, http.StatusInternalServerError, "could not delete account")
		return
	}

	h.logger.Info("account deleted successfully", "user_id", user.ID)

	resp := response.MessageResponse{Message: "Account deleted successfully"}
	response.JSON(w, http.StatusOK, resp)
}
