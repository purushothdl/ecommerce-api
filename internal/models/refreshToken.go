// internal/models/refresh_token.go
package models

import "time"

type RefreshToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Token     string    `json:"-"` // Never serialize this
}

// SessionInfo represents a sanitized view of a refresh token for API responses
type SessionInfo struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}

// ToSessionInfo converts a RefreshToken to SessionInfo for safe API responses
func (rt *RefreshToken) ToSessionInfo() *SessionInfo {
	return &SessionInfo{
		ID:        rt.ID,
		CreatedAt: rt.CreatedAt,
		ExpiresAt: rt.ExpiresAt,
		IsActive:  rt.ExpiresAt.After(time.Now()),
	}
}
