package usecase

import (
	"context"

	"github.com/ssklv/mixfood-order-service/internal/domain"
)

// OrderRepository — то, что мы потребуем от базы данных (в слое инфраструктуры)
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	UpdateStatus(ctx context.Context, orderID int64, status string) (int64, error)
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Order, error)
	GetAllOrders(ctx context.Context, limit, offset int) ([]domain.Order, error) // Для админа
}

// NotificationHub — то, что мы потребуем от вебсокетов для отправки пушей
type NotificationHub interface {
	NotifyUser(userID int64, message interface{})
}
