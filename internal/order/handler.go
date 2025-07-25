package order

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/purushothdl/ecommerce-api/configs"
	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/shared/context"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/response"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type Handler struct {
	orderService  domain.OrderService
	config        configs.StripeConfig
	logger        *slog.Logger
}

func NewHandler(orderService domain.OrderService, cfg configs.StripeConfig, logger *slog.Logger) *Handler {
	return &Handler{
		orderService:  orderService,
		config:        cfg,
		logger:        logger,
	}
}

func (h *Handler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := context.GetUserID(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cartCtx, err := context.GetCart(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Cart not found in context")
		return
	}

	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	v := validator.New()
	ValidateCreateOrderRequest(req, v)
	if !v.Valid() {
		response.JSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	paymentIntent, err := h.orderService.CreateOrder(r.Context(), userID, cartCtx.ID, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrInsufficientStock) {
			response.Error(w, http.StatusConflict, err.Error())
		} else {
			h.logger.Error("failed to create order", "user_id", userID, "error", err)
			response.Error(w, http.StatusInternalServerError, "Could not create order")
		}
		return
	}

	response.JSON(w, http.StatusCreated, paymentIntent)
}

func (h *Handler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Warn("stripe webhook payload read error", "error", err)
		response.Error(w, http.StatusServiceUnavailable, "Failed to read webhook payload")
		return
	}

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), h.config.WebhookSecret)
	h.logger.Info("webhook secret length", "len", len(h.config.WebhookSecret))

	if err != nil {
		h.logger.Warn("stripe webhook signature verification failed", "error", err)
		response.Error(w, http.StatusBadRequest, "Webhook signature verification failed")
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			h.logger.Error("failed to unmarshal payment_intent.succeeded data", "error", err)
			response.Error(w, http.StatusBadRequest, "Failed to parse webhook data")
			return
		}

		if err := h.orderService.HandlePaymentSucceeded(r.Context(), paymentIntent.ID); err != nil {
			h.logger.Error("failed to process successful payment webhook", "pi_id", paymentIntent.ID, "error", err)
			response.Error(w, http.StatusInternalServerError, "Error processing webhook")
			return
		}

	default:
		h.logger.Info("unhandled stripe event type", "type", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}