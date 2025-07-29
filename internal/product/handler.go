// internal/product/handler.go
package product

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/response"
)

type Handler struct {
	productSvc  domain.ProductService
	categorySvc domain.CategoryService
	logger      *slog.Logger
}

func NewHandler(productSvc domain.ProductService, categorySvc domain.CategoryService, logger *slog.Logger) *Handler {
	return &Handler{
		productSvc:  productSvc,
		categorySvc: categorySvc,
		logger:      logger,
	}
}

func (h *Handler) HandleListProducts(w http.ResponseWriter, r *http.Request) {
	filters := domain.ProductFilters{
		Category: r.URL.Query().Get("category"),
		SearchQuery: r.URL.Query().Get("q"),
		Page:     1,
		PageSize: 10, 
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}
	if pageSizeStr := r.URL.Query().Get("limit"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			filters.PageSize = pageSize
		}
	}

	products, err := h.productSvc.ListProducts(r.Context(), filters)
	if err != nil {
		h.logger.Error("failed to list products", "error", err)
		response.Error(w, http.StatusInternalServerError, "could not retrieve products")
		return
	}

	response.JSON(w, http.StatusOK, products)
}

func (h *Handler) HandleGetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "productId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	product, err := h.productSvc.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "product not found")
			return
		}
		h.logger.Error("failed to get product", "product_id", id, "error", err)
		response.Error(w, http.StatusInternalServerError, "could not retrieve product")
		return
	}

	response.JSON(w, http.StatusOK, product)
}

func (h *Handler) HandleListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categorySvc.ListCategories(r.Context())
	if err != nil {
		h.logger.Error("failed to list categories", "error", err)
		response.Error(w, http.StatusInternalServerError, "could not retrieve categories")
		return
	}
	response.JSON(w, http.StatusOK, categories)
}