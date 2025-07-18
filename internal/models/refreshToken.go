// internal/models/refreshToken.go
package models

import "time"

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	// Token is only populated when generating a new token to be sent to the user.
	// It is not stored in the database.
	Token string `json:"-"`
}