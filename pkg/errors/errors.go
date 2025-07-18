//pkg/errors/errors.go
package apperrors

import "errors"

// User-related errors
var (
	ErrUserNotFound   = errors.New("user not found")
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrEditConflict   = errors.New("edit conflict")
)

// Auth-related errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid refresh token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrUnexpectedMethod   = errors.New("unexpected signing method")
	ErrSessionNotFound 	  = errors.New("session not found")
	ErrWeakPassword       = errors.New("password too weak")
)
