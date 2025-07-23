package address

import "github.com/purushothdl/ecommerce-api/internal/models"

// AddressResponse represents a single address response
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

// AddressListResponse represents a list of addresses
type AddressListResponse struct {
    Addresses []*models.UserAddress `json:"addresses"`
}
