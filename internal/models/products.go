// internal/models/product.go
package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq" 
)

// Dimensions represents the product's size.
type Dimensions struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Depth  float64 `json:"depth"`
}

type Product struct {
	ID                  int64           `json:"id"`
	Name                string          `json:"name"`
	Description         string          `json:"description"`
	Price               float64         `json:"price"`
	StockQuantity       int             `json:"stock_quantity"`
	CategoryID          int64           `json:"-"` 					// Foreign key
	Category            *Category       `json:"category,omitempty"` // For joining data
	Brand               string          `json:"brand,omitempty"`
	SKU                 string          `json:"sku,omitempty"`
	Images              pq.StringArray  `json:"images" gorm:"type:text[]"`
	Thumbnail           string          `json:"thumbnail,omitempty"`
	Dimensions          json.RawMessage `json:"dimensions,omitempty" gorm:"type:jsonb"`
	WarrantyInformation string          `json:"warranty_information,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	Version             int             `json:"version"`
}