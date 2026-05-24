package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-order-service/internal/domain"
	"github.com/ssklv/mixfood-order-service/internal/infrastructure"
	"github.com/ssklv/mixfood-order-service/internal/usecase"
)

type Logger interface {
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
}

type OrderHandler struct {
	uc          *usecase.OrderUsecase
	tokenParser usecase.TokenParser
	wsHub       *infrastructure.WsHub
	logger      Logger
}

func NewOrderHandler(uc *usecase.OrderUsecase, tokenParser usecase.TokenParser, wsHub *infrastructure.WsHub, logger Logger) *OrderHandler {
	return &OrderHandler{
		uc:          uc,
		tokenParser: tokenParser,
		wsHub:       wsHub,
		logger:      logger,
	}
}

func (h *OrderHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	authRequired := func(c fiber.Ctx) error {
		tokenString := c.Cookies("access_token")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: missing token"})
		}

		userID, role, err := h.tokenParser.ParseToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: invalid token"})
		}

		c.Locals("userID", userID)
		c.Locals("role", role)
		return c.Next()
	}

	adminRequired := func(c fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: admin access required"})
		}
		return c.Next()
	}

	api.Get("/ws/orders",
		authRequired,
		websocket.New(func(c *websocket.Conn) {
			userID := c.Locals("userID").(int64)
			h.wsHub.Register(userID, c)
			for {
				if _, _, err := c.ReadMessage(); err != nil {
					break
				}
			}
			h.wsHub.Unregister(userID)
		}),
	)

	orders := api.Group("/orders", authRequired)
	orders.Post("", h.CreateOrder)
	orders.Get("", h.GetUserOrders)

	admin := api.Group("/admin/orders", authRequired, adminRequired)
	admin.Get("", h.GetAdminOrders)
	admin.Patch("/:id/status", h.UpdateStatus)
}

// CreateOrder обрабатывает POST /api/v1/orders
// @Summary      Создание нового заказа
// @Description  Принимает корзину товаров и адрес, высчитывает стоимость и создает заказ
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                  true  "Bearer JWT Token"
// @Param        input          body      domain.CreateOrderInput true  "Данные для создания заказа"
// @Success      201            {object}  domain.Order
// @Failure      400            {object}  map[string]string       "Невалидные входные данные"
// @Failure      401            {object}  map[string]string       "Пользователь не авторизован"
// @Failure      500            {object}  map[string]string       "Внутренняя ошибка сервера"
// @Router       /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var input domain.CreateOrderInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	order, err := h.uc.CreateOrder(c.Context(), userID, input)
	if err != nil {
		if errors.Is(err, usecase.ErrEmptyCart) ||
			errors.Is(err, usecase.ErrAddressRequired) ||
			errors.Is(err, usecase.ErrInvalidQuantity) ||
			errors.Is(err, usecase.ErrInvalidPrice) ||
			errors.Is(err, usecase.ErrInvalidProductID) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		h.logger.Error("failed to create order: " + err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

// GetUserOrders обрабатывает GET /api/v1/orders
// @Summary      Получение заказов пользователя
// @Description  Возвращает историю заказов текущего авторизованного пользователя с пагинацией
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true   "Bearer JWT Token"
// @Param        limit          query     int     false  "Количество заказов (default 10)"
// @Param        offset         query     int     false  "Смещение (default 0)"
// @Success      200            {array}   domain.Order
// @Failure      401            {object}  map[string]string "Неавторизован"
// @Failure      500            {object}  map[string]string "Внутренняя ошибка"
// @Router       /api/v1/orders [get]
func (h *OrderHandler) GetUserOrders(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	limitStr := c.Query("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offsetStr := c.Query("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	orders, err := h.uc.GetUserOrders(c.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get user orders: " + err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch orders"})
	}

	return c.JSON(orders)
}

// UpdateStatus обрабатывает PATCH /api/v1/admin/orders/:id/status
// @Summary      Обновление статуса заказа
// @Description  Только для администраторов
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id     path      int                    true  "ID заказа"
// @Param        input  body      domain.UpdateStatusInput true  "Новый статус"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      403    {object}  map[string]string
// @Router       /api/v1/admin/orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(c fiber.Ctx) error {
	orderIDStr := c.Params("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid order ID"})
	}

	var input domain.UpdateStatusInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err = h.uc.UpdateOrderStatus(c.Context(), orderID, input)
	if err != nil {
		if errors.Is(err, usecase.ErrStatusEmpty) || errors.Is(err, usecase.ErrInvalidStatus) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		h.logger.Error("failed to update status: " + err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update status"})
	}

	return c.JSON(fiber.Map{"message": "Order status updated successfully"})
}

// GetAdminOrders обрабатывает GET /api/v1/admin/orders
// @Summary      Получение всех заказов (для админа)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        limit  query     int  false  "Лимит"
// @Param        offset query     int  false  "Смещение"
// @Success      200    {array}   domain.Order
// @Router       /api/v1/admin/orders [get]
func (h *OrderHandler) GetAdminOrders(c fiber.Ctx) error {
	limitStr := c.Query("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offsetStr := c.Query("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	orders, err := h.uc.GetAdminOrders(c.Context(), limit, offset)
	if err != nil {
		h.logger.Error("failed to get admin orders: " + err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch admin orders"})
	}

	return c.JSON(orders)
}
