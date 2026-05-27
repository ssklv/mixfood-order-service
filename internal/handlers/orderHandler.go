package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-order-service/internal/domain"
	"github.com/ssklv/mixfood-order-service/internal/infrastructure"
	"github.com/ssklv/mixfood-order-service/internal/usecase"
)

type OrderHandler struct {
	uc          *usecase.OrderUsecase
	tokenParser usecase.TokenParser
	wsHub       *infrastructure.WsHub
	logger      Logger
}

func NewOrderHandler(uc *usecase.OrderUsecase, tp usecase.TokenParser, wsHub *infrastructure.WsHub, log Logger) *OrderHandler {
	return &OrderHandler{uc: uc, tokenParser: tp, wsHub: wsHub, logger: log}
}

func (h *OrderHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	orders := router.Group("/orders", authMiddleware)
	orders.Post("", h.CreateOrder)
	orders.Get("", h.GetUserOrders)

	admin := router.Group("/admin/orders", authMiddleware, h.adminRequired)
	admin.Get("", h.GetAdminOrders)
	admin.Patch("/:id/status", h.UpdateStatus)
}

func (h *OrderHandler) adminRequired(c fiber.Ctx) error {
	role, ok := c.Locals("userRole").(string)
	if !ok || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Forbidden: admin access required"})
	}
	return c.Next()
}

// @Summary Create a new order
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body domain.CreateOrderInput true "Order details"
// @Success 201 {object} domain.Order
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/orders [post]
func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "Unauthorized"})
	}

	var input domain.CreateOrderInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	order, err := h.uc.CreateOrder(c.Context(), userID, input)
	if err != nil {
		if errors.Is(err, usecase.ErrEmptyCart) || errors.Is(err, usecase.ErrAddressRequired) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
		}
		h.logger.Error("failed to create order", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

// @Summary Get user's orders
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit of records"
// @Param offset query int false "Offset of records"
// @Success 200 {array} domain.Order
// @Failure 500 {object} ErrorResponse
// @Router /api/orders [get]
func (h *OrderHandler) GetUserOrders(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "Unauthorized"})
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		offset = 0
	}

	orders, err := h.uc.GetUserOrders(c.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get user orders", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to fetch orders"})
	}
	return c.JSON(orders)
}

// @Summary Update order status (Admin only)
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param input body domain.UpdateStatusInput true "New status"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/admin/orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(c fiber.Ctx) error {
	orderID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid order ID"})
	}

	var input domain.UpdateStatusInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	if err := h.uc.UpdateOrderStatus(c.Context(), orderID, input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
	}
	return c.JSON(MessageResponse{Message: "Status updated"})
}

// @Summary Get all orders (Admin only)
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit of records"
// @Param offset query int false "Offset of records"
// @Success 200 {array} domain.Order
// @Failure 500 {object} ErrorResponse
// @Router /api/admin/orders [get]
func (h *OrderHandler) GetAdminOrders(c fiber.Ctx) error {
	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil {
		limit = 20
	}
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		offset = 0
	}

	orders, err := h.uc.GetAdminOrders(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to fetch admin orders"})
	}
	return c.JSON(orders)
}
