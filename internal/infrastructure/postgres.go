package infrastructure

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(connectionString string) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), connectionString)
}
