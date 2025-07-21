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

// Handler holds dependencies for cart-related HTTP handlers.
type Handler struct {
	cartSvc domain.CartService
	logger  *slog.Logger
}

// NewHandler creates a new cart handler.
func NewHandler(cartSvc domain.CartService, logger *slog.Logger) *Handler {
	return &Handler{cartSvc: cartSvc, logger: logger}
}


// HandleGetCart retrieves the current user's or session's cart.
func (h *Handler) HandleGetCart(w http.ResponseWriter, r *http.Request) {
	// The CartMiddleware has already done the hard work of finding or creating a cart.
	cart, err := context.GetCart(r.Context())
	if err != nil {
		h.logger.Error("could not get cart from context in handler", "error", err)
		response.Error(w, http.StatusInternalServerError, "could not retrieve cart from context")
		return
	}

	// The cart in the context is just the base cart object.
	// We need to call the service to populate it with items and totals.
	fullCart, err := h.cartSvc.GetCartContents(r.Context(), cart.ID)
	if err != nil {
		h.logger.Error("failed to get full cart contents", "cart_id", cart.ID, "error", err)
		response.Error(w, http.StatusInternalServerError, "could not retrieve cart contents")
		return
	}

	h.logger.Info("successfully retrieved cart", "cart_id", fullCart.ID, "item_count", len(fullCart.Items))
	response.JSON(w, http.StatusOK, fullCart)
}

// HandleAddItem adds a new product or increases the quantity of an existing one in the cart.
func (h *Handler) HandleAddItem(w http.ResponseWriter, r *http.Request) {
	cart, err := context.GetCart(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not retrieve cart from context")
		return
	}

	var input AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	h.logger.Info("request to add item to cart", "cart_id", cart.ID, "product_id", input.ProductID, "quantity", input.Quantity)

	updatedCart, err := h.cartSvc.AddProductToCart(r.Context(), cart.ID, input.ProductID, input.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			response.Error(w, http.StatusNotFound, "product not found")
		case errors.Is(err, apperrors.ErrInsufficientStock):
			response.Error(w, http.StatusConflict, "insufficient stock for the requested quantity")
		default:
			h.logger.Error("unhandled error adding item to cart", "error", err)
			response.Error(w, http.StatusInternalServerError, "could not add item to cart")
		}
		return
	}

	response.JSON(w, http.StatusOK, updatedCart)
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