package dto

import (
    "time"
    "github.com/purushothdl/ecommerce-api/internal/models"
)

// UserResponse is a shared DTO for representing a user in API responses.
// It's used by both the user and admin domains.
type UserResponse struct {
    ID        int64       `json:"id" example:"1"`
    Name      string      `json:"name" example:"John Doe"`
    Email     string      `json:"email" example:"john@example.com"`
    Role      models.Role `json:"role" example:"user"`
    CreatedAt time.Time   `json:"created_at" example:"2023-01-01T00:00:00Z"`
    Version   int         `json:"version" example:"1"`
}

// NewUserResponse creates a UserResponse DTO from a User model.
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