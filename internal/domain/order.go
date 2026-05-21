package domain

import "time"

type Order struct {
	ID         int64       `json:"id"`
	UserID     int64       `json:"user_id"`
	AddressID  int64       `json:"address_id"`
	TotalPrice float64     `json:"total_price"`
	Status     string      `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Items      []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CreateOrderInput struct {
	AddressID int64       `json:"address_id"`
	Items     []OrderItem `json:"items"`
}
type UpdateStatusInput struct {
	Status string `json:"status"`
}
