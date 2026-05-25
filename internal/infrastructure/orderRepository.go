package infrastructure

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssklv/mixfood-order-service/internal/domain"
)

type orderRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewOrderRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *orderRepository {
	return &orderRepository{db: db, psql: psql}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql, args, err := r.psql.
		Insert("orders").
		Columns("user_id", "address_id", "total_price", "status", "delivery_time").
		Values(order.UserID, order.AddressID, order.TotalPrice, order.Status, order.DeliveryTime).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return err
	}

	err = tx.QueryRow(ctx, sql, args...).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range order.Items {
		sql, args, err := r.psql.
			Insert("order_items").
			Columns("order_id", "product_id", "quantity", "price").
			Values(order.ID, order.Items[i].ProductID, order.Items[i].Quantity, order.Items[i].Price).
			Suffix("RETURNING id").
			ToSql()
		if err != nil {
			return err
		}

		err = tx.QueryRow(ctx, sql, args...).Scan(&order.Items[i].ID)
		if err != nil {
			return err
		}
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

func (r *orderRepository) GetAllOrders(ctx context.Context, limit, offset int) ([]domain.Order, error) {
	// Используем squirrel для построения запроса
	query, args, err := r.psql.Select("id", "user_id", "address_id", "total_price", "status", "created_at").
		From("orders").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		OrderBy("created_at DESC").
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.AddressID, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func (r *orderRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Order, error) {
	sql, args, err := r.psql.
		Select("id", "user_id", "address_id", "total_price", "status", "delivery_time", "created_at", "updated_at").
		From("orders").
		Where(sq.Eq{"user_id": userID}).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		OrderBy("created_at DESC").
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
		err := rows.Scan(&o.ID, &o.UserID, &o.AddressID, &o.TotalPrice, &o.Status, &o.DeliveryTime, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *orderRepository) GetOrder(ctx context.Context, id int64) (*domain.Order, error) {
	sql, args, err := r.psql.
		Select("id", "user_id", "address_id", "total_price", "status", "delivery_time", "created_at", "updated_at").
		From("orders").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var o domain.Order
	err = r.db.QueryRow(ctx, sql, args...).Scan(&o.ID, &o.UserID, &o.AddressID, &o.TotalPrice, &o.Status, &o.DeliveryTime, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
