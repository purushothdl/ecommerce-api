// internal/user/responses.go
package user

import (
	"time"
	"github.com/purushothdl/ecommerce-api/internal/models"
)

// UserResponse represents user data in API responses
type UserResponse struct {
	ID        int64     `json:"id" example:"1"`
	Name      string    `json:"name" example:"John Doe"`
	Email     string    `json:"email" example:"john@example.com"`
	Role      models.Role    `json:"role" example:"user"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	Version   int       `json:"version" example:"1"`
}

// NewUserResponse creates a UserResponse from User model
func NewUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		Version:   user.Version,
	}
}

// ProfileResponse represents profile data
type ProfileResponse struct {
	*UserResponse
	Message string `json:"message" example:"Welcome to your profile!"`
}

// UpdateProfileResponse represents profile update response
type UpdateProfileResponse struct {
	*UserResponse
	Message string `json:"message" example:"Profile updated successfully"`
}
