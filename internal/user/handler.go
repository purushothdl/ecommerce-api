// internal/user/handler.go
package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	usercontext "github.com/purushothdl/ecommerce-api/internal/shared/context"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

type Handler struct {
	userService domain.UserService
	authService domain.AuthService 
	logger      *slog.Logger
}

func NewHandler(userService domain.UserService, authService domain.AuthService, logger *slog.Logger) *Handler {
	return &Handler{
		userService: userService,
		authService: authService,
		logger:      logger,
	}
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

	user, err := h.userService.Register(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicateEmail) {
			response.Error(w, http.StatusConflict, "Email address is already in use")
		} else {
			response.Error(w, http.StatusInternalServerError, "Could not create user")
		}
		return
	}

	response.JSON(w, http.StatusCreated, NewUserResponse(user))
}

func (h *Handler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	userCtx, err := usercontext.GetUser(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error retrieving user from context")
		return
	}

	// Fetch the full user model
	user, err := h.userService.GetProfile(r.Context(), userCtx.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not retrieve profile")
		return
	}

	// Create the structured response
	resp := ProfileResponse{
		UserResponse: NewUserResponse(user),
		Message:      "Welcome to your protected profile!",
	}

	response.JSON(w, http.StatusOK, resp)
}

// HandleUpdateProfile allows users to update their profile information
func (h *Handler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, err := usercontext.GetUser(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	updatedUser, err := h.userService.UpdateProfile(r.Context(), user.ID, input.Name, input.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, apperrors.ErrDuplicateEmail) {
			response.Error(w, http.StatusConflict, "email address is already in use")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not update profile")
		return
	}

	resp := UpdateProfileResponse{
		UserResponse: NewUserResponse(updatedUser),
		Message:      "Profile updated successfully",
	}

	response.JSON(w, http.StatusOK, resp)
}

// HandleChangePassword allows users to change their password
func (h *Handler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	user, err := usercontext.GetUser(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
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
		// Password was changed successfully, session revocation is secondary
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "password changed successfully. please log in again.",
	})
}

// HandleDeleteAccount allows users to delete their account
func (h *Handler) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	user, err := usercontext.GetUser(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input DeleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	if err := h.userService.DeleteAccount(r.Context(), user.ID, input.Password); err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "password is incorrect")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not delete account")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "account deleted successfully",
	})
}
