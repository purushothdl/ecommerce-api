// internal/admin/responses.go
package admin

import "github.com/purushothdl/ecommerce-api/internal/shared/dto"

// UserListResponse represents the response for listing all users.
type UserListResponse struct {
	Users []*dto.UserResponse `json:"users"`
	Count int                 `json:"count"`
}