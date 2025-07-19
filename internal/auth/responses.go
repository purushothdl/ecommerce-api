package auth

import (
	"time"
)

// LoginResponse represents a successful login response
type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshResponse represents a successful token refresh
type RefreshResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// SessionInfo represents user session information
type SessionInfo struct {
	ID        int64     `json:"id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	ExpiresAt time.Time `json:"expires_at" example:"2023-01-08T00:00:00Z"`
	IsActive  bool      `json:"is_active" example:"true"`
}

// SessionsResponse represents a list of user sessions
type SessionsResponse struct {
	Sessions []SessionInfo `json:"sessions"`
	Count    int           `json:"count" example:"3"`
}
