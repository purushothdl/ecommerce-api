package dto

// PaymentIntent represents a Stripe payment intent record
type PaymentIntent struct {
    ID           string    `json:"id"`
    OrderID      int64     `json:"order_id"`
    Amount       float64   `json:"amount"`
    Currency     string    `json:"currency"`
    Status       string    `json:"status"`
    ClientSecret string    `json:"client_secret"`
}
