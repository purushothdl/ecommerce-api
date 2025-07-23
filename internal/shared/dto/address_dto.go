package dto

import "github.com/purushothdl/ecommerce-api/internal/models"

// CreateAddressRequest is the input for creating an address
type CreateAddressRequest struct {
    Name               string `json:"name" example:"John Doe"`
    Phone              string `json:"phone" example:"+1234567890"`
    Street1            string `json:"street1" example:"123 Main St"`
    Street2            string `json:"street2" example:"Apt 4B"`
    City               string `json:"city" example:"New York"`
    State              string `json:"state" example:"NY"`
    PostalCode         string `json:"postal_code" example:"10001"`
    Country            string `json:"country" example:"USA"`
    IsDefaultShipping  bool   `json:"is_default_shipping" example:"true"`
    IsDefaultBilling   bool   `json:"is_default_billing" example:"false"`
}

// UpdateAddressRequest is the input for updating an address
type UpdateAddressRequest struct {
    Name               *string `json:"name" example:"John Doe"`
    Phone              *string `json:"phone" example:"+1234567890"`
    Street1            *string `json:"street1" example:"123 Main St"`
    Street2            *string `json:"street2" example:"Apt 4B"`
    City               *string `json:"city" example:"New York"`
    State              *string `json:"state" example:"NY"`
    PostalCode         *string `json:"postal_code" example:"10001"`
    Country            *string `json:"country" example:"USA"`
    IsDefaultShipping  *bool   `json:"is_default_shipping" example:"true"`
    IsDefaultBilling   *bool   `json:"is_default_billing" example:"false"`
}

// SetDefaultAddressRequest is the input for setting default type
type SetDefaultAddressRequest struct {
    Type string `json:"type" example:"shipping"` // "shipping" or "billing"
}

// AddressResponse is the output for a single address
type AddressResponse struct {
    ID                 int64  `json:"id"`
    Name               string `json:"name"`
    Phone              string `json:"phone"`
    Street1            string `json:"street1"`
    Street2            string `json:"street2,omitempty"`
    City               string `json:"city"`
    State              string `json:"state"`
    PostalCode         string `json:"postal_code"`
    Country            string `json:"country"`
    IsDefaultShipping  bool   `json:"is_default_shipping"`
    IsDefaultBilling   bool   `json:"is_default_billing"`
}

// AddressListResponse is the output for listing addresses
type AddressListResponse struct {
    Addresses []*models.UserAddress `json:"addresses"`
}
