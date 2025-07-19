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
	users, err := h.adminService.ListUsers(r.Context())
	if err != nil {
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
}

func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	user, err := h.adminService.CreateUser(r.Context(), input.Name, input.Email, input.Password, input.Role)
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicateEmail) {
			response.Error(w, http.StatusConflict, "email address is already in use")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not create user")
		return
	}
	response.JSON(w, http.StatusCreated, dto.NewUserResponse(user))
}

func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var input UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	updatedUser, err := h.adminService.UpdateUser(r.Context(), userID, input.Name, input.Email, input.Role)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not update user")
		return
	}
	response.JSON(w, http.StatusOK, dto.NewUserResponse(updatedUser))
}

func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.adminService.DeleteUser(r.Context(), userID); err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "could not delete user")
		return
	}

	resp := response.MessageResponse{Message: "user deleted successfully"}
	response.JSON(w, http.StatusOK, resp)
}