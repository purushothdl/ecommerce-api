package domain

// ProductFilters defines options for filtering products
type ProductFilters struct {
	Category string
	Page     int
	PageSize int
}