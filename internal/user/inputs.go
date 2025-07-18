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


type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (r UpdateProfileRequest) Validate(v *validator.Validator) {
	v.Check(r.Name != "", "name", "must be provided")
	v.Check(len(r.Name) >= 2, "name", "must be at least 2 characters long")
	v.Check(len(r.Name) <= 100, "name", "must not be more than 100 characters long")
	
	v.Check(r.Email != "", "email", "must be provided")
	v.Check(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email address")
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (r ChangePasswordRequest) Validate(v *validator.Validator) {
	v.Check(r.CurrentPassword != "", "current_password", "must be provided")
	v.Check(r.NewPassword != "", "new_password", "must be provided")
	v.Check(len(r.NewPassword) >= 8, "new_password", "must be at least 8 characters long")
	v.Check(len(r.NewPassword) <= 72, "new_password", "must not be more than 72 characters long")
}

type DeleteAccountRequest struct {
	Password string `json:"password"`
}

func (r DeleteAccountRequest) Validate(v *validator.Validator) {
	v.Check(r.Password != "", "password", "must be provided")
}