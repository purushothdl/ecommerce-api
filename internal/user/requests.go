// internal/user/requests.go
package user

import "github.com/purushothdl/ecommerce-api/pkg/validator"

// CreateUserRequest represents user registration request
type CreateUserRequest struct {
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"password123"`
}

func (r CreateUserRequest) Validate(v *validator.Validator) {
	v.Check(validator.NotBlank(r.Name), "name", "must be provided")
	v.Check(len(r.Name) >= 2, "name", "must be at least 2 characters long")
	v.Check(len(r.Name) <= 100, "name", "must not exceed 100 characters")
	
	v.Check(validator.NotBlank(r.Email), "email", "must be provided")
	v.Check(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email address")
	
	v.Check(validator.NotBlank(r.Password), "password", "must be provided")
	v.Check(len(r.Password) >= 8, "password", "must be at least 8 characters long")
	v.Check(len(r.Password) <= 72, "password", "must not exceed 72 characters")
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"john@example.com"`
}

func (r UpdateProfileRequest) Validate(v *validator.Validator) {
	v.Check(validator.NotBlank(r.Name), "name", "must be provided")
	v.Check(len(r.Name) >= 2, "name", "must be at least 2 characters long")
	v.Check(len(r.Name) <= 100, "name", "must not exceed 100 characters")
	
	v.Check(validator.NotBlank(r.Email), "email", "must be provided")
	v.Check(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email address")
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" example:"oldpassword123"`
	NewPassword     string `json:"new_password" example:"newpassword123"`
}

func (r ChangePasswordRequest) Validate(v *validator.Validator) {
	v.Check(validator.NotBlank(r.CurrentPassword), "current_password", "must be provided")
	v.Check(validator.NotBlank(r.NewPassword), "new_password", "must be provided")
	v.Check(len(r.NewPassword) >= 8, "new_password", "must be at least 8 characters long")
	v.Check(len(r.NewPassword) <= 72, "new_password", "must not exceed 72 characters")
	v.Check(r.CurrentPassword != r.NewPassword, "new_password", "must be different from current password")
}

// DeleteAccountRequest represents account deletion request
type DeleteAccountRequest struct {
	Password string `json:"password" example:"password123"`
}

func (r DeleteAccountRequest) Validate(v *validator.Validator) {
	v.Check(validator.NotBlank(r.Password), "password", "must be provided")
}
