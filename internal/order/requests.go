package order

import (

	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

// ValidateCreateOrderRequest validates the create order request
func ValidateCreateOrderRequest(r dto.CreateOrderRequest, v *validator.Validator) {
    v.Check(r.ShippingAddressID > 0, "shipping_address_id", "must be a valid address ID")
    v.Check(r.BillingAddressID > 0, "billing_address_id", "must be a valid address ID")
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
