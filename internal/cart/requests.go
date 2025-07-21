package cart

import "github.com/purushothdl/ecommerce-api/pkg/validator"

// AddItemRequest defines the request body for adding an item to the cart.
type AddItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

// Validate checks the AddItemRequest for correctness.
func (r AddItemRequest) Validate(v *validator.Validator) {
	v.Check(r.ProductID > 0, "product_id", "must be a positive integer")
	v.Check(r.Quantity > 0, "quantity", "must be a positive integer")
}

// UpdateItemRequest defines the request body for updating an item's quantity.
type UpdateItemRequest struct {
	Quantity int `json:"quantity"`
}

// Validate checks the UpdateItemRequest for correctness.
func (r UpdateItemRequest) Validate(v *validator.Validator) {
	// Quantity can be 0, which means "remove the item".
	v.Check(r.Quantity >= 0, "quantity", "must be a non-negative integer")
}