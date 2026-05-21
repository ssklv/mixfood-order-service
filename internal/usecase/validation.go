package usecase

import (
	"github.com/ssklv/mixfood-order-service/internal/domain"
)

func ValidateCreateOrder(input domain.CreateOrderInput) error {
	if input.AddressID <= 0 {
		return ErrAddressRequired
	}

	if len(input.Items) == 0 {
		return ErrEmptyCart
	}

	for _, item := range input.Items {
		if item.ProductID <= 0 {
			return ErrInvalidProductID
		}
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.Price < 0 {
			return ErrInvalidPrice
		}
	}

	return nil
}

func ValidateUpdateStatus(input domain.UpdateStatusInput) error {
	if input.Status == "" {
		return ErrStatusEmpty
	}

	validStatuses := map[string]bool{
		"new":        true,
		"cooking":    true,
		"delivering": true,
		"done":       true,
		"cancelled":  true,
	}

	if !validStatuses[input.Status] {
		return ErrInvalidStatus
	}

	return nil
}
