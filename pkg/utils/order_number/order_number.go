package order_number

import (
	"fmt"
	"time"
)

// Generate creates a unique order number (stub - use real seq in DB if needed)
func Generate() string {
    return fmt.Sprintf("ORD-%d-%06d", time.Now().Year(), time.Now().UnixNano()%1000000)
}
