// internal/models/user.go
package models

import "time"

// User represents the core user entity in our domain.
type User struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never send this in JSON responses
	Role         string    `json:"role"`
	Version      int       `json:"version"`
}