// internal/user/responses.go
package user

import "github.com/purushothdl/ecommerce-api/internal/shared/dto"

// LoginResponse combines user data with an access token for authentication.
type LoginResponse struct {
	User        *dto.UserResponse `json:"user"`
	AccessToken string            `json:"access_token"`
}
