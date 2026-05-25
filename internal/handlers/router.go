package handlers

import (
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/ssklv/mixfood-order-service/internal/infrastructure"
	"github.com/ssklv/mixfood-order-service/internal/usecase"
)

func ConfigureApp(
	app *fiber.App,
	orderUC *usecase.OrderUsecase,
	tokenProvider usecase.TokenParser,
	wsHub *infrastructure.WsHub,
	log Logger,
) {
	app.Use(recover.New())
	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:8080", "http://localhost:8082"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	app.Get("/swagger/*", swaggo.HandlerDefault)

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		userID, _, _ := tokenProvider.ParseToken(c.Query("token"))
		wsHub.Register(userID, c)
		defer wsHub.Unregister(userID)
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))

	authMiddleware := NewAuthMiddleware(tokenProvider, log)
	apiGroup := app.Group("/api")

	// ВАЖНО: передаем все 4 аргумента
	handler := NewOrderHandler(orderUC, tokenProvider, wsHub, log)
	handler.RegisterRoutes(apiGroup, authMiddleware)
}
