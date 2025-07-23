package order

import (
	"fmt"

	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

// ValidateCreateOrderRequest validates the create order request
func ValidateCreateOrderRequest(r dto.CreateOrderRequest, v *validator.Validator) {
    v.Check(len(r.Items) > 0, "items", "must provide at least one item")
    for i, item := range r.Items {
        v.Check(item.ProductID > 0, fmt.Sprintf("items[%d].product_id", i), "must be a valid product ID")
        v.Check(item.Quantity > 0, fmt.Sprintf("items[%d].quantity", i), "must be at least 1")
    }

    v.Check(r.ShippingAddressID != nil || r.ShippingAddress != nil, "shipping_address", "must provide shipping address ID or details")
    v.Check(r.BillingAddressID != nil || r.BillingAddress != nil, "billing_address", "must provide billing address ID or details")
    v.Check(validator.NotBlank(r.PaymentMethod), "payment_method", "must be provided")
}

// ValidateConfirmPaymentRequest validates the confirm payment request
func ValidateConfirmPaymentRequest(r dto.ConfirmPaymentRequest, v *validator.Validator) {
    v.Check(validator.NotBlank(r.PaymentIntentID), "payment_intent_id", "must be provided")
}

// ValidateCancelOrderRequest validates the cancel order request
func ValidateCancelOrderRequest(r dto.CancelOrderRequest, v *validator.Validator) {
    // Optional reason - no strict validation needed, but can add if required
    if r.Reason != "" {
        v.Check(len(r.Reason) <= 500, "reason", "must not exceed 500 characters")
    }
}
