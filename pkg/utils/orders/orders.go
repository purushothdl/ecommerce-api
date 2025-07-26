package orders

import (
	"fmt"
	"time"

	"github.com/purushothdl/ecommerce-api/internal/models"
)

// Generate creates a unique order number (stub - use real seq in DB if needed)
func Generate() string {
    return fmt.Sprintf("ORD-%d-%06d", time.Now().Year(), time.Now().UnixNano()%1000000)
}

// Helper function to convert a models.UserAddress to a models.OrderAddress (JSONB snapshot)
func ToOrderAddress(addr *models.UserAddress) models.OrderAddress {
	return models.OrderAddress{
		Name:       addr.Name,
		Phone:      addr.Phone,
		Street1:    addr.Street1,
		Street2:    addr.Street2,
		City:       addr.City,
		State:      addr.State,
		PostalCode: addr.PostalCode,
		Country:    addr.Country,
	}
}