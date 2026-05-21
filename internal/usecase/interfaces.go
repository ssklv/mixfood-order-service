package usecase

import (
	"context"

	"github.com/ssklv/mixfood-order-service/internal/domain"
)

//go:generate mockery --name OrderRepository --inpackage --case underscore
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	UpdateStatus(ctx context.Context, orderID int64, status string) (int64, error)
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Order, error)

	GetAllOrders(ctx context.Context, limit, offset int) ([]domain.Order, error)
}

//go:generate mockery --name NotificationHub --inpackage --case underscore
type NotificationHub interface {
	NotifyUser(userID int64, message interface{})
}

type TokenParser interface {
	ParseToken(tokenString string) (int64, string, error)
}
