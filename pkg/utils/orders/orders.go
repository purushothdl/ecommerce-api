package orders

import (
	"fmt"
	"time"

	"github.com/purushothdl/ecommerce-api/internal/models"
)

// Generate creates a unique order number (stub - use real seq in DB if needed)
func Generate() string {
	t := time.Now()
	return fmt.Sprintf("ORD-%s-%06d", t.Format("20060102"), t.UnixNano()%1_000_000)
}

// GenerateTrackingID creates a unique tracking number for shipments
func GenerateTrackingID() string {
    return fmt.Sprintf("TRK-%d-%06d", time.Now().Year(), time.Now().UnixNano()%1000000)
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