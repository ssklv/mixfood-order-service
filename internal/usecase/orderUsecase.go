package usecase

import (
	"context"

	"github.com/ssklv/mixfood-order-service/internal/domain"
)

type OrderUsecase struct {
	repo OrderRepository
	hub  NotificationHub
}

func NewOrderUsecase(repo OrderRepository, hub NotificationHub) *OrderUsecase {
	return &OrderUsecase{
		repo: repo,
		hub:  hub,
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, userID int64, input domain.CreateOrderInput) (*domain.Order, error) {
	if err := ValidateCreateOrder(input); err != nil {
		return nil, err
	}

	var totalPrice float64
	for i := range input.Items {
		totalPrice += input.Items[i].Price * float64(input.Items[i].Quantity)
	}

	order := &domain.Order{
		UserID:     userID,
		AddressID:  input.AddressID,
		TotalPrice: totalPrice,
		Status:     "new",
		Items:      input.Items,
	}

	if err := u.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (u *OrderUsecase) UpdateOrderStatus(ctx context.Context, orderID int64, input domain.UpdateStatusInput) error {
	if err := ValidateUpdateStatus(input); err != nil {
		return err
	}
	userID, err := u.repo.UpdateStatus(ctx, orderID, input.Status)
	if err != nil {
		return err
	}
	u.hub.NotifyUser(userID, map[string]interface{}{
		"type":     "ORDER_STATUS_CHANGED",
		"order_id": orderID,
		"status":   input.Status,
		"message":  "Your order status has been updated to: " + input.Status,
	})

	return nil
}

func (u *OrderUsecase) GetUserOrders(ctx context.Context, userID int64, limit, offset int) ([]domain.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return u.repo.GetByUserID(ctx, userID, limit, offset)
}

func (u *OrderUsecase) GetAdminOrders(ctx context.Context, limit, offset int) ([]domain.Order, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.repo.GetAllOrders(ctx, limit, offset)
}
