package pg

import (
	"context"
	"damirqa/loyalty-system/internal/domain"
	"database/sql"
	"errors"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error)
	GetOrderByID(ctx context.Context, number string) (*domain.Order, error)
}

type orderRepository struct {
	db *sql.DB
}

func (o orderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	query := "INSERT INTO orders (number, user_id, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)"
	_, err := o.db.ExecContext(ctx, query, order.Number, order.UserID, order.Status, order.Accrual, order.UploadedAt)
	return err
}

func (o orderRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	query := "SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC"
	rows, err := o.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var order domain.Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

func (o orderRepository) GetOrderByID(ctx context.Context, number string) (*domain.Order, error) {
	query := "SELECT number, status, accrual, uploaded_at FROM orders WHERE number = $1"
	row := o.db.QueryRowContext(ctx, query, number)

	var order domain.Order
	err := row.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &order, nil
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}
