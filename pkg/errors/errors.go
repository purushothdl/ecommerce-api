package apperrors

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrDuplicateEmail = errors.New("duplicate email")
	// Add other shared errors here
)
