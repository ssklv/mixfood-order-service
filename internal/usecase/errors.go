package usecase

import "errors"

var (
	ErrEmptyCart        = errors.New("cart is empty, cannot place order")
	ErrAddressRequired  = errors.New("order cannot be placed: delivery address is required in profile")
	ErrInvalidQuantity  = errors.New("product quantity must be greater than zero")
	ErrInvalidPrice     = errors.New("product price cannot be negative")
	ErrInvalidProductID = errors.New("invalid product ID")

	ErrInvalidStatus = errors.New("invalid order status")
	ErrStatusEmpty   = errors.New("order status cannot be empty")

	ErrOrderNotFound = errors.New("order not found")
)
