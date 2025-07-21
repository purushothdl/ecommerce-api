// internal/models/base.go
package models

import "time"

// BaseModel contains common fields for all models.
type BaseModel struct {
    ID        int64     `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}