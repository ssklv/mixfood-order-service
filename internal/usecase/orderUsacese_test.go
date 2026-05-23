package usecase

import (
	"context"
	"testing"

	"github.com/ssklv/mixfood-order-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOrder_Success(t *testing.T) {
	repo := NewMockOrderRepository(t)
	hub := NewMockNotificationHub(t)

	repo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		order := args.Get(1).(*domain.Order)
		order.ID = 100
	})

	uc := NewOrderUsecase(repo, hub)

	input := domain.CreateOrderInput{
		AddressID:    10,
		DeliveryTime: "18:30",
		Items: []domain.OrderItem{
			{ProductID: 1, Quantity: 2, Price: 250.00},
			{ProductID: 2, Quantity: 1, Price: 150.00},
		},
	}

	order, err := uc.CreateOrder(context.Background(), 42, input)

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, int64(100), order.ID)
	assert.Equal(t, 650.00, order.TotalPrice)
	assert.Equal(t, "new", order.Status)
	assert.Equal(t, "18:30", order.DeliveryTime)
	repo.AssertExpectations(t)
}

func TestCreateOrder_InvalidTime(t *testing.T) {
	repo := NewMockOrderRepository(t)
	hub := NewMockNotificationHub(t)

	uc := NewOrderUsecase(repo, hub)

	input := domain.CreateOrderInput{
		AddressID:    10,
		DeliveryTime: "asap",
		Items: []domain.OrderItem{
			{ProductID: 1, Quantity: 2, Price: 250.00},
		},
	}

	_, err := uc.CreateOrder(context.Background(), 42, input)

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestGetUserOrders_Success(t *testing.T) {
	repo := NewMockOrderRepository(t)
	hub := NewMockNotificationHub(t)
	uc := NewOrderUsecase(repo, hub)

	expectedOrders := []domain.Order{
		{ID: 1, UserID: 42, TotalPrice: 500.0, DeliveryTime: "18:30"},
	}

	repo.On("GetByUserID", mock.Anything, int64(42), 10, 0).Return(expectedOrders, nil)

	orders, err := uc.GetUserOrders(context.Background(), 42, 0, -1)

	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	repo.AssertExpectations(t)
}

func TestGetAdminOrders_Success(t *testing.T) {
	repo := NewMockOrderRepository(t)
	hub := NewMockNotificationHub(t)
	uc := NewOrderUsecase(repo, hub)

	expectedOrders := []domain.Order{
		{ID: 1, TotalPrice: 500.0, DeliveryTime: "18:30"},
	}

	repo.On("GetAllOrders", mock.Anything, 20, 0).Return(expectedOrders, nil)

	orders, err := uc.GetAdminOrders(context.Background(), 0, 0)

	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	repo.AssertExpectations(t)
}

func TestUpdateOrderStatus_Success(t *testing.T) {
	repo := NewMockOrderRepository(t)
	hub := NewMockNotificationHub(t)
	uc := NewOrderUsecase(repo, hub)

	input := domain.UpdateStatusInput{Status: "cooking"}

	repo.On("UpdateStatus", mock.Anything, int64(100), "cooking").Return(int64(42), nil)

	hub.On("NotifyUser", int64(42), mock.Anything).Return()

	err := uc.UpdateOrderStatus(context.Background(), 100, input)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	hub.AssertExpectations(t)
}

func TestCreateOrder_EmptyCart(t *testing.T) {
	repo := NewMockOrderRepository(t)
	hub := NewMockNotificationHub(t)

	uc := NewOrderUsecase(repo, hub)

	input := domain.CreateOrderInput{
		AddressID:    10,
		DeliveryTime: "12:00",
		Items:        []domain.OrderItem{},
	}

	_, err := uc.CreateOrder(context.Background(), 42, input)

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}
