// internal/cart/handler.go
package cart

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/shared/context"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

type Handler struct {
	cartSvc domain.CartService
	logger  *slog.Logger
}

func NewHandler(cartSvc domain.CartService, logger *slog.Logger) *Handler {
	return &Handler{cartSvc: cartSvc, logger: logger}
}

// HandleGetCart returns the cart with items and totals
func (h *Handler) HandleGetCart(w http.ResponseWriter, r *http.Request) {
	cartCtx, err := context.GetCart(r.Context())
	if err != nil {
		h.logger.Error("cart context missing", "error", err)
		response.Error(w, http.StatusInternalServerError, "cart unavailable")
		return
	}

	fullCart, err := h.cartSvc.GetCartContents(r.Context(), cartCtx.ID)
	if err != nil {
		h.logger.Error("failed to load cart", "cart_id", cartCtx.ID, "error", err)
		response.Error(w, http.StatusInternalServerError, "could not load cart")
		return
	}

	// Use NewCartResponse to format the response
	resp := NewCartResponse(fullCart, fullCart.Items)
	response.JSON(w, http.StatusOK, resp)
}

// HandleAddItem adds a product to the cart
func (h *Handler) HandleAddItem(w http.ResponseWriter, r *http.Request) {
	cartCtx, err := context.GetCart(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "cart unavailable")
		return
	}

	var input AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request format")
		return
	}

	// Validate input
	v := validator.New()
	if input.Validate(v); !v.Valid() {
		response.ValidationError(w, v.Errors)
		return
	}

	// Business logic
	updatedCart, err := h.cartSvc.AddProductToCart(r.Context(), cartCtx.ID, input.ProductID, input.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			response.Error(w, http.StatusNotFound, "product not found")
		case errors.Is(err, apperrors.ErrInsufficientStock):
			response.Error(w, http.StatusConflict, "insufficient stock")
		default:
			h.logger.Error("cart update failed", "error", err)
			response.Error(w, http.StatusInternalServerError, "could not update cart")
		}
		return
	}

	// Format response
	resp := NewCartResponse(updatedCart, updatedCart.Items)
	response.JSON(w, http.StatusOK, resp)
}

// HandleUpdateItem changes the quantity of a specific item in the cart.
func (h *Handler) HandleUpdateItem(w http.ResponseWriter, r *http.Request) {
	cart, err := context.GetCart(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not retrieve cart from context")
		return
	}

	productIDStr := chi.URLParam(r, "productId")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid product ID in URL")
		return
	}

	var input UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	h.logger.Info("request to update item quantity", "cart_id", cart.ID, "product_id", productID, "new_quantity", input.Quantity)

	updatedCart, err := h.cartSvc.UpdateProductInCart(r.Context(), cart.ID, productID, input.Quantity)
	if err != nil {
		// The service layer handles the case where quantity is 0 by calling Remove.
		// We only need to handle generic errors here.
		h.logger.Error("unhandled error updating item in cart", "error", err)
		response.Error(w, http.StatusInternalServerError, "could not update item in cart")
		return
	}

	response.JSON(w, http.StatusOK, updatedCart)
}

// HandleRemoveItem completely removes a product from the cart.
func (h *Handler) HandleRemoveItem(w http.ResponseWriter, r *http.Request) {
	cart, err := context.GetCart(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not retrieve cart from context")
		return
	}

	productIDStr := chi.URLParam(r, "productId")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid product ID in URL")
		return
	}

	h.logger.Info("request to remove item from cart", "cart_id", cart.ID, "product_id", productID)

	_, err = h.cartSvc.RemoveProductFromCart(r.Context(), cart.ID, productID)
	if err != nil {
		h.logger.Error("unhandled error removing item from cart", "error", err)
		response.Error(w, http.StatusInternalServerError, "could not remove item from cart")
		return
	}

	resp := response.MessageResponse{Message: "item removed successfully"}
	response.JSON(w, http.StatusOK, resp)
}
