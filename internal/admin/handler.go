// internal/admin/handler.go
package admin

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

type Handler struct {
	adminService domain.AdminService
	logger       *slog.Logger
}

func NewHandler(adminService domain.AdminService, logger *slog.Logger) *Handler {
	return &Handler{
		adminService: adminService,
		logger:       logger,
	}
}

func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("list users request received")
	users, err := h.adminService.ListUsers(r.Context())

	if err != nil {
		h.logger.Error("failed to list users", "error", err)
		response.Error(w, http.StatusInternalServerError, "could not retrieve users")
		return
	}

	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = dto.NewUserResponse(user)
	}

	resp := UserListResponse{
		Users: userResponses,
		Count: len(userResponses),
	}

	response.JSON(w, http.StatusOK, resp)

	h.logger.Info("users listed successfully", "count", len(userResponses))
}

func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("create user request received")
	var input CreateUserRequest
	
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

	user, err := h.adminService.CreateUser(r.Context(), input.Name, input.Email, input.Password, input.Role)
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicateEmail) {
			h.logger.Warn("duplicate email", "email", input.Email)
			response.Error(w, http.StatusConflict, "email address is already in use")
			return
		}
		h.logger.Error("failed to create user", "error", err)
		response.Error(w, http.StatusInternalServerError, "could not create user")
		return
	}

	response.JSON(w, http.StatusCreated, dto.NewUserResponse(user))

	h.logger.Info("user created successfully", "user_id", user.ID, "email", user.Email)
}

func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update user request received")
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)

	if err != nil {
		h.logger.Warn("invalid user ID", "user_id", userIDStr, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var input UpdateUserRequest
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

	updatedUser, err := h.adminService.UpdateUser(r.Context(), userID, input.Name, input.Email, input.Role)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			h.logger.Warn("user not found", "user_id", userID)
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("failed to update user", "user_id", userID, "error", err)
		response.Error(w, http.StatusInternalServerError, "could not update user")
		return
	}

	response.JSON(w, http.StatusOK, dto.NewUserResponse(updatedUser))

	h.logger.Info("user updated successfully", "user_id", userID)
}

func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("delete user request received")
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)

	if err != nil {
		h.logger.Warn("invalid user ID", "user_id", userIDStr, "error", err)
		response.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.adminService.DeleteUser(r.Context(), userID); err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			h.logger.Warn("user not found", "user_id", userID)
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("failed to delete user", "user_id", userID, "error", err)
		response.Error(w, http.StatusInternalServerError, "could not delete user")
		return
	}

	resp := response.MessageResponse{Message: "user deleted successfully"}
	response.JSON(w, http.StatusOK, resp)

	h.logger.Info("user deleted successfully", "user_id", userID)
}
