// internal/context/context.go
package context

import (
	"context"
	"errors"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// UserContext represents the authenticated user in request context
type UserContext struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// SetUser adds user to context
func SetUser(ctx context.Context, user UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// GetUser retrieves user from context
func GetUser(ctx context.Context) (UserContext, error) {
	user, ok := ctx.Value(UserContextKey).(UserContext)
	if !ok {
		return UserContext{}, errors.New("user not found in context")
	}
	return user, nil
}

// GetUserID is a convenience method to get just the user ID
func GetUserID(ctx context.Context) (int64, error) {
	user, err := GetUser(ctx)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}
