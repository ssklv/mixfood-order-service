package main

//swag init -g cmd/main.go --output docs
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
// @description                 API for order management and processing
// @host                        localhost:8083
// @BasePath                    /
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter token in format: Bearer <token>
func main() {
	logger.InitLogger()
	if logger.Logger != nil {
		defer logger.Logger.Sync()
	}

	if err := godotenv.Load(); err != nil && logger.Logger != nil {
		logger.Logger.Warn(".env file not found")
	}

	cfg := config.Load()
	logAdapter := &zapAdapter{}

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil {
		if logger.Logger != nil {
			logger.Logger.Fatal("Database connection error: " + err.Error())
		}
		panic("Database connection error: " + err.Error())
	}
	defer conn.Close()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Layer Initialization
	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret)
	wsHub := infrastructure.NewWsHub(logAdapter)
	orderRepo := infrastructure.NewOrderRepository(conn, psql)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, wsHub)

	app := fiber.New(fiber.Config{AppName: "MixFood Order Service"})

	handlers.ConfigureApp(app, orderUsecase, tokenProvider, wsHub, logAdapter)

	if logger.Logger != nil {
		logger.Logger.Info(fmt.Sprintf("Order service started on port :%s", cfg.ServerPort))
	}

	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		if logger.Logger != nil {
			logger.Logger.Fatal("Server failed to start: " + err.Error())
		}
		panic("Server failed to start: " + err.Error())
	}
}
