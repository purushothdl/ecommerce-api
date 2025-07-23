package address

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	context "github.com/purushothdl/ecommerce-api/internal/shared/context"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

type Handler struct {
    service domain.AddressService
    logger  *slog.Logger
}

// NewHandler creates a new address handler
func NewHandler(service domain.AddressService, logger *slog.Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

// HandleCreate creates a new address
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
    userID, err := context.GetUserID(r.Context())
    if err != nil {
        h.logger.Warn("user ID not found in context", "error", err)
        response.Error(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    var req dto.CreateAddressRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Warn("invalid create address payload", "error", err)
        response.Error(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    v := validator.New()
    ValidateCreateAddressRequest(req, v)
    if !v.Valid() {
        h.logger.Warn("create address validation failed", "errors", v.Errors)
        response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
        return
    }

    addr, err := h.service.Create(r.Context(), userID, &req)
    if err != nil {
        h.logger.Error("failed to create address", "user_id", userID, "error", err)
        response.Error(w, http.StatusInternalServerError, "Could not create address")
        return
    }

    response.JSON(w, http.StatusCreated, addr)
}

// HandleGetByID gets an address by ID
func (h *Handler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
    userID, err := context.GetUserID(r.Context())
    if err != nil {
        h.logger.Warn("user ID not found in context", "error", err)
        response.Error(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid address ID")
        return
    }

    addr, err := h.service.GetByID(r.Context(), id, userID)
    if err != nil {
        if errors.Is(err, apperrors.ErrNotFound) {
            response.Error(w, http.StatusNotFound, "Address not found")
        } else if errors.Is(err, apperrors.ErrUnauthorized) {
            response.Error(w, http.StatusUnauthorized, "Not authorized")
        } else {
            h.logger.Error("failed to get address", "id", id, "user_id", userID, "error", err)
            response.Error(w, http.StatusInternalServerError, "Could not get address")
        }
        return
    }

    response.JSON(w, http.StatusOK, addr)
}

// HandleList lists all addresses for the user
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
    userID, err := context.GetUserID(r.Context())
    if err != nil {
        h.logger.Warn("user ID not found in context", "error", err)
        response.Error(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    addresses, err := h.service.ListByUserID(r.Context(), userID)
    if err != nil {
        h.logger.Error("failed to list addresses", "user_id", userID, "error", err)
        response.Error(w, http.StatusInternalServerError, "Could not list addresses")
        return
    }

    response.JSON(w, http.StatusOK, dto.AddressListResponse{Addresses: addresses})
}

// HandleUpdate updates an existing address
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
    userID, err := context.GetUserID(r.Context())
    if err != nil {
        h.logger.Warn("user ID not found in context", "error", err)
        response.Error(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid address ID")
        return
    }

    var req dto.UpdateAddressRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Warn("invalid update address payload", "error", err)
        response.Error(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    v := validator.New()
    ValidateUpdateAddressRequest(req, v)
    if !v.Valid() {
        h.logger.Warn("update address validation failed", "errors", v.Errors)
        response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
        return
    }

    addr, err := h.service.Update(r.Context(), userID, id, &req)
    if err != nil {
        if errors.Is(err, apperrors.ErrNotFound) {
            response.Error(w, http.StatusNotFound, "Address not found")
        } else {
            h.logger.Error("failed to update address", "id", id, "user_id", userID, "error", err)
            response.Error(w, http.StatusInternalServerError, "Could not update address")
        }
        return
    }

    response.JSON(w, http.StatusOK, addr)
}

// HandleDelete deletes an address
func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
    userID, err := context.GetUserID(r.Context())
    if err != nil {
        h.logger.Warn("user ID not found in context", "error", err)
        response.Error(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid address ID")
        return
    }

    if err := h.service.Delete(r.Context(), userID, id); err != nil {
        if errors.Is(err, apperrors.ErrNotFound) {
            response.Error(w, http.StatusNotFound, "Address not found")
        } else {
            h.logger.Error("failed to delete address", "id", id, "user_id", userID, "error", err)
            response.Error(w, http.StatusInternalServerError, "Could not delete address")
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// HandleSetDefault sets an address as default
func (h *Handler) HandleSetDefault(w http.ResponseWriter, r *http.Request) {
    userID, err := context.GetUserID(r.Context())
    if err != nil {
        h.logger.Warn("user ID not found in context", "error", err)
        response.Error(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid address ID")
        return
    }

    var req dto.SetDefaultAddressRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Warn("invalid set default payload", "error", err)
        response.Error(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    v := validator.New()
    ValidateSetDefaultAddressRequest(req, v)
    if !v.Valid() {
        h.logger.Warn("set default validation failed", "errors", v.Errors)
        response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
        return
    }

    if err := h.service.SetDefault(r.Context(), userID, id, req.Type); err != nil {
        if errors.Is(err, apperrors.ErrNotFound) {
            response.Error(w, http.StatusNotFound, "Address not found")
        } else {
            h.logger.Error("failed to set default address", "id", id, "user_id", userID, "error", err)
            response.Error(w, http.StatusInternalServerError, "Could not set default address")
        }
        return
    }

    response.JSON(w, http.StatusOK, response.MessageResponse{Message: "Default address set successfully"})
}
