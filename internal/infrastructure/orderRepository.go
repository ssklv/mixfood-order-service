package infrastructure

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssklv/mixfood-order-service/internal/domain"
)

var orderCols = []string{
	"id",
	"user_id",
	"address_id",
	"total_price",
	"status",
	"delivery_time",
	"created_at",
	"updated_at",
}

var orderItemCols = []string{
	"id",
	"order_id",
	"product_id",
	"quantity",
	"price",
}

type orderRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewOrderRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *orderRepository {
	return &orderRepository{
		db:   db,
		psql: psql,
	}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ErrFailedToBeginTx
	}
	defer tx.Rollback(ctx)

	orderSql, orderArgs, err := r.psql.
		Insert("orders").
		Columns(orderCols[1:6]...).
		Values(
			order.UserID,
			order.AddressID,
			order.TotalPrice,
			order.Status,
			order.DeliveryTime,
		).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()

	if err != nil {
		return err
	}

	err = tx.QueryRow(ctx, orderSql, orderArgs...).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range order.Items {
		itemSql, itemArgs, err := r.psql.
			Insert("order_items").
			Columns(orderItemCols[1:]...).
			Values(
				order.ID,
				order.Items[i].ProductID,
				order.Items[i].Quantity,
				order.Items[i].Price,
			).
			Suffix("RETURNING id").
			ToSql()

		if err != nil {
			return err
		}

		err = tx.QueryRow(ctx, itemSql, itemArgs...).Scan(&order.Items[i].ID)
		if err != nil {
			return err
		}
		order.Items[i].OrderID = order.ID
	}

	return tx.Commit(ctx)
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) (int64, error) {
	sql, args, err := r.psql.
		Update("orders").
		Set("status", status).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": orderID}).
		Suffix("RETURNING user_id").
		ToSql()

	if err != nil {
		return 0, err
	}

	var userID int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrOrderNotFound
		}
		return 0, err
	}

	return userID, nil
}

func (r *orderRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Order, error) {
	sql, args, err := r.psql.
		Select(orderCols...).
		From("orders").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := scanOrder(rows, &o); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func (r *orderRepository) GetAllOrders(ctx context.Context, limit, offset int) ([]domain.Order, error) {
	sql, args, err := r.psql.
		Select(orderCols...).
		From("orders").
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := scanOrder(rows, &o); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func scanOrder(row pgx.Row, o *domain.Order) error {
	return row.Scan(
		&o.ID,
		&o.UserID,
		&o.AddressID,
		&o.TotalPrice,
		&o.Status,
		&o.DeliveryTime,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
}
