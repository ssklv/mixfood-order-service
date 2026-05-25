package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-order-service/internal/usecase"
)

func NewAuthMiddleware(tp usecase.TokenParser, log Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		} else {
			token = c.Cookies(AccessCookie)
		}

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "Unauthorized: missing token"})
		}

		userID, role, err := tp.ParseToken(token)
		if err != nil {
			log.Warn("Middleware: invalid token parsing attempt", "error", err)
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "Invalid token"})
		}

		c.Locals("userID", userID)
		c.Locals("userRole", role)
		return c.Next()
	}
}
