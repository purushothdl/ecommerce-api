package models

// Define a custom type for Role for type safety
type Role string

// Define the valid roles as constants
const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// IsValid is a helper method to check if a role is valid.
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser:
		return true
	}
	return false
}

// User represents the core user entity in our domain.
type User struct {
    BaseModel    
    Name         string    `json:"name"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`
    Role         Role      `json:"role"`
    Version      int       `json:"version"`
}