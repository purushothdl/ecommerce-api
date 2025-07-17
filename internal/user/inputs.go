// internal/user/inputs.go
package user

import (
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *CreateUserRequest) Validate(v *validator.Validator) {
	v.Check(validator.NotBlank(r.Name), "name", "must be provided")
	v.Check(validator.NotBlank(r.Email), "email", "must be provided")
	v.Check(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email address")
	v.Check(validator.NotBlank(r.Password), "password", "must be provided")
	v.Check(validator.MinChars(r.Password, 8), "password", "must be at least 8 characters long")
}
