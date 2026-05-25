package handlers

import "github.com/gofiber/fiber/v3"

const (
	AccessCookie  = "access_token"
	RefreshCookie = "refresh_token"
)

type Handler interface {
	RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler)
}

type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
