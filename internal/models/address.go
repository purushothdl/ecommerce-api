package models

import "time"

// UserAddress represents a saved address in the user's address book
type UserAddress struct {
    ID                 int64     `json:"id"`
    UserID             int64     `json:"user_id"`
    Name               string    `json:"name"`
    Phone              string    `json:"phone"`
    Street1            string    `json:"street1"`
    Street2            string    `json:"street2,omitempty"`
    City               string    `json:"city"`
    State              string    `json:"state"`
    PostalCode         string    `json:"postal_code"`
    Country            string    `json:"country"`
    IsDefaultShipping  bool      `json:"is_default_shipping"`
    IsDefaultBilling   bool      `json:"is_default_billing"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}
