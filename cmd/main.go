package main

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/ssklv/mixfood-order-service/docs"
	"github.com/ssklv/mixfood-order-service/internal/config"
	"github.com/ssklv/mixfood-order-service/internal/handlers"
	"github.com/ssklv/mixfood-order-service/internal/infrastructure"
	"github.com/ssklv/mixfood-order-service/internal/usecase"
	"github.com/ssklv/pizza-shared/pkg/logger"
)

type zapAdapter struct{}

func (za *zapAdapter) Error(msg string, fields ...any) { logger.Logger.Error(msg) }
func (za *zapAdapter) Warn(msg string, fields ...any)  { logger.Logger.Warn(msg) }

// @title       MixFood Order Service API
// @version     1.0
// @description Микросервис для управления заказами в системе MixFood.
// @host        localhost:8083
// @BasePath    /
func main() {
	logger.InitLogger()
	defer logger.Logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Logger.Warn("Файл .env не найден")
	}

	cfg := config.Load()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	app := fiber.New(fiber.Config{
		AppName: "MixFood Order Service v1.0",
	})

	app.Get("/swagger/*", adaptor.HTTPHandler(httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"),
	)))

	app.Get("/docs/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
	}))

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Logger.Fatal("Ошибка БД: " + err.Error())
	}
	defer conn.Close()

	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret)
	wsHub := infrastructure.NewWsHub()

	orderRepo := infrastructure.NewOrderRepository(conn, psql)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, wsHub)

	logAdapter := &zapAdapter{}

	orderHandler := handlers.NewOrderHandler(orderUsecase, tokenProvider, wsHub, logAdapter)
	orderHandler.RegisterRoutes(app)

	logger.Logger.Info(fmt.Sprintf("Сервер стартовал на :%s", cfg.ServerPort))
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		logger.Logger.Fatal("Сервер упал: " + err.Error())
	}
}
