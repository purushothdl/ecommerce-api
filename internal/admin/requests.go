// internal/admin/requests.go
package admin

import (
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

// CreateUserRequest represents the request for an admin to create a new user.
type CreateUserRequest struct {
	Name     string      `json:"name" example:"New User"`
	Email    string      `json:"email" example:"new.user@example.com"`
	Password string      `json:"password" example:"password123"`
	Role     models.Role `json:"role" example:"user"`
}

func (r CreateUserRequest) Validate(v *validator.Validator) {
	v.Check(validator.NotBlank(r.Name), "name", "must be provided")
	v.Check(len(r.Name) >= 2, "name", "must be at least 2 characters long")

	v.Check(validator.NotBlank(r.Email), "email", "must be provided")
	v.Check(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email address")

	v.Check(validator.NotBlank(r.Password), "password", "must be provided")
	v.Check(len(r.Password) >= 8, "password", "must be at least 8 characters long")

	v.Check(r.Role.IsValid(), "role", "is not a valid role")
}

// UpdateUserRequest represents the request for an admin to update a user's details.
type UpdateUserRequest struct {
	Name  *string      `json:"name,omitempty" example:"Updated Name"`
	Email *string      `json:"email,omitempty" example:"updated.user@example.com"`
	Role  *models.Role `json:"role,omitempty" example:"admin"`
}

func (r UpdateUserRequest) Validate(v *validator.Validator) {
	if r.Name != nil {
		v.Check(validator.NotBlank(*r.Name), "name", "must not be empty if provided")
	}
	if r.Email != nil {
		v.Check(validator.NotBlank(*r.Email), "email", "must not be empty if provided")
		v.Check(validator.Matches(*r.Email, validator.EmailRX), "email", "must be a valid email address")
	}
	if r.Role != nil {
		v.Check(r.Role.IsValid(), "role", "is not a valid role")
	}

	// Ensure at least one field is provided for an update
	v.Check(r.Name != nil || r.Email != nil || r.Role != nil, "request", "at least one field must be provided for an update")
}