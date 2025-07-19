package auth

import (
	"time"
)

// BaseResponse contains common fields for API responses
type BaseResponse struct {
	Message string `json:"message,omitempty" example:"Operation successful"`
}

// LoginResponse represents the response for a successful login
type LoginResponse struct {
	BaseResponse
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshTokenResponse represents the response for a successful token refresh
type RefreshTokenResponse struct {
	BaseResponse
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// SessionInfo represents information about a user session
type SessionInfo struct {
	ID        int64     `json:"id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	ExpiresAt time.Time `json:"expires_at" example:"2023-01-08T00:00:00Z"`
	IsActive  bool      `json:"is_active" example:"true"`
}

// GetSessionsResponse represents the response for fetching user sessions
type GetSessionsResponse struct {
	BaseResponse
	Sessions []SessionInfo `json:"sessions"`
	Count    int           `json:"count" example:"3"`
}
