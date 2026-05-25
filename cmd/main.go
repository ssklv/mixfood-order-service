package main

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"

	"github.com/ssklv/mixfood-order-service/internal/config"
	"github.com/ssklv/mixfood-order-service/internal/handlers"
	"github.com/ssklv/mixfood-order-service/internal/infrastructure"
	"github.com/ssklv/mixfood-order-service/internal/usecase"
	"github.com/ssklv/pizza-shared/pkg/logger"

	_ "github.com/ssklv/mixfood-order-service/docs"
)

type zapAdapter struct{}

func (za *zapAdapter) Info(msg string, fields ...any) {
	if logger.Logger != nil {
		logger.Logger.Sugar().Infow(msg, fields...)
	}
}

func (za *zapAdapter) Error(msg string, fields ...any) {
	if logger.Logger != nil {
		logger.Logger.Sugar().Errorw(msg, fields...)
	}
}

func (za *zapAdapter) Warn(msg string, fields ...any) {
	if logger.Logger != nil {
		logger.Logger.Sugar().Warnw(msg, fields...)
	}
}

// @title                       Mixfood Order Service API
// @version                     1.0
// @description                 API для управления заказами
// @host                        localhost:8083
// @BasePath                    /
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Введите токен в формате: Bearer <token>
func main() {
	logger.InitLogger()
	if logger.Logger != nil {
		defer logger.Logger.Sync()
	}

	if err := godotenv.Load(); err != nil && logger.Logger != nil {
		logger.Logger.Warn("Файл .env не найден")
	}

	cfg := config.Load()
	logAdapter := &zapAdapter{}

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil && logger.Logger != nil {
		logger.Logger.Fatal("Ошибка БД: " + err.Error())
	}
	defer conn.Close()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Инициализация слоев
	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret)
	wsHub := infrastructure.NewWsHub(logAdapter)
	orderRepo := infrastructure.NewOrderRepository(conn, psql)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, wsHub)

	app := fiber.New(fiber.Config{AppName: "MixFood Order Service"})

	handlers.ConfigureApp(app, orderUsecase, tokenProvider, wsHub, logAdapter)

	if logger.Logger != nil {
		logger.Logger.Info(fmt.Sprintf("Сервер заказа запущен на порту :%s", cfg.ServerPort))
	}

	if err := app.Listen(":" + cfg.ServerPort); err != nil && logger.Logger != nil {
		logger.Logger.Fatal("Сервер упал: " + err.Error())
	}
}
