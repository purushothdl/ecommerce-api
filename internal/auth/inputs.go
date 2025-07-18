// internal/auth/inputs.go
package auth

import (
	"github.com/purushothdl/ecommerce-api/pkg/validator"
)

type CreateTokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *CreateTokenRequest) Validate(v *validator.Validator) {
	v.Check(validator.NotBlank(r.Email), "email", "must be provided")
	v.Check(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email address")
	v.Check(validator.NotBlank(r.Password), "password", "must be provided")
}

// New struct for the refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *RefreshTokenRequest) Validate(v *validator.Validator) {
	v.Check(validator.NotBlank(r.RefreshToken), "refresh_token", "must be provided")
}