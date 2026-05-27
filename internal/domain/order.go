package domain

import "time"

type Order struct {
	ID           int64       `json:"id"`
	UserID       int64       `json:"userId"`
	AddressID    int64       `json:"addressId"`
	TotalPrice   float64     `json:"totalPrice"`
	Status       string      `json:"status"`
	DeliveryTime string      `json:"deliveryTime"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	Items        []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"orderId"`
	ProductID int64   `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CreateOrderInput struct {
	AddressID    int64                  `json:"addressId"`
	DeliveryTime string                 `json:"deliveryTime"`
	TotalPrice   float64                `json:"totalPrice"`
	Items        []CreateOrderItemInput `json:"items"`
}

type CreateOrderItemInput struct {
	ProductID int64   `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type UpdateStatusInput struct {
	Status string `json:"status"`
}
