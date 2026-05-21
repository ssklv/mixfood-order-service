package handlers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-order-service/internal/infrastructure"
	"github.com/ssklv/mixfood-order-service/internal/usecase"
)

func SetupRoutes(app *fiber.App, orderHandler *OrderHandler, tokenParser usecase.TokenParser, wsHub *infrastructure.WsHub) {
	api := app.Group("/api/v1")

	authRequired := func(c fiber.Ctx) error {
		tokenString := c.Cookies("jwt")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: missing token"})
		}

		userID, role, err := tokenParser.ParseToken(tokenString)
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

			wsHub.Register(userID, c)

			for {
				if _, _, err := c.ReadMessage(); err != nil {
					break
				}
			}

			wsHub.Unregister(userID)
		}),
	)

	orders := api.Group("/orders", authRequired)
	orders.Post("", orderHandler.CreateOrder)
	orders.Get("", orderHandler.GetUserOrders)

	admin := api.Group("/admin/orders", authRequired, adminRequired)
	admin.Get("", orderHandler.GetAdminOrders)
	admin.Patch("/:id/status", orderHandler.UpdateStatus)
}
