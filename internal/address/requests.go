package address

import (
    "github.com/purushothdl/ecommerce-api/internal/shared/dto"
    "github.com/purushothdl/ecommerce-api/pkg/validator"
)

// ValidateCreateAddressRequest validates the create request
func ValidateCreateAddressRequest(r dto.CreateAddressRequest, v *validator.Validator) {
    v.Check(validator.NotBlank(r.Name), "name", "must be provided")
    v.Check(len(r.Name) >= 2, "name", "must be at least 2 characters long")
    v.Check(len(r.Name) <= 100, "name", "must not exceed 100 characters")

    v.Check(validator.NotBlank(r.Phone), "phone", "must be provided")
    v.Check(len(r.Phone) >= 10, "phone", "must be at least 10 characters long")

    v.Check(validator.NotBlank(r.Street1), "street1", "must be provided")
    v.Check(validator.NotBlank(r.City), "city", "must be provided")
    v.Check(validator.NotBlank(r.State), "state", "must be provided")
    v.Check(validator.NotBlank(r.PostalCode), "postal_code", "must be provided")
    v.Check(validator.NotBlank(r.Country), "country", "must be provided")
}

// ValidateUpdateAddressRequest validates the update request
func ValidateUpdateAddressRequest(r dto.UpdateAddressRequest, v *validator.Validator) {
    // Ensure at least one field is provided
    v.Check(
        r.Name != nil || r.Phone != nil || r.Street1 != nil || r.Street2 != nil ||
            r.City != nil || r.State != nil || r.PostalCode != nil || r.Country != nil ||
            r.IsDefaultShipping != nil || r.IsDefaultBilling != nil,
        "", "at least one field must be provided for update",
    )

    // Validate provided fields
    if r.Name != nil {
        v.Check(len(*r.Name) >= 2, "name", "must be at least 2 characters long")
        v.Check(len(*r.Name) <= 100, "name", "must not exceed 100 characters")
    }
    if r.Phone != nil {
        v.Check(len(*r.Phone) >= 10, "phone", "must be at least 10 characters long")
    }
    if r.Street1 != nil {
        v.Check(validator.NotBlank(*r.Street1), "street1", "must not be empty if provided")
    }
    if r.City != nil {
        v.Check(validator.NotBlank(*r.City), "city", "must not be empty if provided")
    }
    if r.State != nil {
        v.Check(validator.NotBlank(*r.State), "state", "must not be empty if provided")
    }
    if r.PostalCode != nil {
        v.Check(validator.NotBlank(*r.PostalCode), "postal_code", "must not be empty if provided")
    }
    if r.Country != nil {
        v.Check(validator.NotBlank(*r.Country), "country", "must not be empty if provided")
    }
}

// ValidateSetDefaultAddressRequest validates the set default request
func ValidateSetDefaultAddressRequest(r dto.SetDefaultAddressRequest, v *validator.Validator) {
    v.Check(r.Type == "shipping" || r.Type == "billing", "type", "must be 'shipping' or 'billing'")
}
