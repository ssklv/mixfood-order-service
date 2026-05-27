package usecase

//go test -coverprofile=coverage.out ./internal/usecase/...
import (
	"context"
	"testing"

	"github.com/ssklv/mixfood-order-service/internal/domain"
	"github.com/ssklv/mixfood-order-service/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOrder_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewOrderRepository(t)
	mockHub := mocks.NewNotificationHub(t)
	uc := NewOrderUsecase(mockRepo, mockHub)

	input := domain.CreateOrderInput{
		AddressID: 1,
		Items: []domain.CreateOrderItemInput{
			{ProductID: 1, Price: 100.0, Quantity: 2},
		},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)

	order, err := uc.CreateOrder(ctx, 42, input)

	assert.NoError(t, err)
	assert.Equal(t, 200.0, order.TotalPrice)
	mockRepo.AssertExpectations(t)
}

func TestUpdateOrderStatus_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewOrderRepository(t)
	mockHub := mocks.NewNotificationHub(t)
	uc := NewOrderUsecase(mockRepo, mockHub)

	orderID := int64(123)
	userID := int64(42)
	newStatus := "cooking"

	mockRepo.On("UpdateStatus", ctx, orderID, newStatus).Return(userID, nil)

	mockHub.On("NotifyUser", userID, mock.MatchedBy(func(msg map[string]interface{}) bool {
		return msg["status"] == newStatus && msg["orderId"] == orderID
	})).Return()

	err := uc.UpdateOrderStatus(ctx, orderID, domain.UpdateStatusInput{Status: newStatus})

	assert.NoError(t, err)
	mockHub.AssertExpectations(t)
}

func TestUpdateOrderStatus_RepoError(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewOrderRepository(t)
	mockHub := mocks.NewNotificationHub(t)
	uc := NewOrderUsecase(mockRepo, mockHub)

	mockRepo.On("UpdateStatus", ctx, int64(123), "cooking").Return(int64(0), assert.AnError)

	err := uc.UpdateOrderStatus(ctx, 123, domain.UpdateStatusInput{Status: "cooking"})

	assert.Error(t, err)
	mockHub.AssertNotCalled(t, "NotifyUser", mock.Anything, mock.Anything)
}

func TestGetUserOrders_LimitOffset(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewOrderRepository(t)
	uc := NewOrderUsecase(mockRepo, nil)

	mockRepo.On("GetByUserID", ctx, int64(1), 10, 0).Return([]domain.Order{}, nil)
	_, _ = uc.GetUserOrders(ctx, 1, 0, -5)

	mockRepo.AssertExpectations(t)
}
