package main

import (
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/ssklv/mixfood-order-service/internal/config"
	"github.com/ssklv/mixfood-order-service/internal/handlers"
	"github.com/ssklv/mixfood-order-service/internal/infrastructure"
	"github.com/ssklv/mixfood-order-service/internal/usecase"
	"github.com/ssklv/pizza-shared/pkg/logger"
)

type zapAdapter struct{}

func (za *zapAdapter) Error(msg string, fields ...any) { logger.Logger.Error(msg) }
func (za *zapAdapter) Warn(msg string, fields ...any)  { logger.Logger.Warn(msg) }

// @title MixFood Order Service API
// @version 1.0
// @description Сервис заказов для доставки еды
// @host localhost:8083
// @BasePath /
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

	app.Get("/docs/*", adaptor.HTTPHandler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs")))))

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

	// Исправлено: используем твой реальный TokenProvider (принимает только секрет)
	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret)
	wsHub := infrastructure.NewWsHub()
	// Строку go wsHub.Run() убрали, так как твоему хабу фоновый цикл не нужен

	orderRepo := infrastructure.NewOrderRepository(conn, psql)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, wsHub)

	logAdapter := &zapAdapter{}

	// Передаем правильный OrderHandler (с большой буквы)
	orderHandler := handlers.NewOrderHandler(orderUsecase, tokenProvider, wsHub, logAdapter)
	orderHandler.RegisterRoutes(app)

	logger.Logger.Info(fmt.Sprintf("Сервер стартовал на :%s", cfg.ServerPort))
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		logger.Logger.Fatal("Сервер упал: " + err.Error())
	}
}
