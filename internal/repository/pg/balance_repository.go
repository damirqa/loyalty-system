package pg

import (
	"context"
	"damirqa/loyalty-system/internal/domain"
	"database/sql"
)

type BalanceRepository interface {
	GetBalanceByUserID(ctx context.Context, userID int64) (*domain.Balance, error)
	UpdateBalance(ctx context.Context, userID int64, delta float64) error
	CreateWithdrawal(ctx context.Context, withdrawal *domain.Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]*domain.Withdrawal, error)
}

type balanceRepository struct {
	db *sql.DB
}

func (b balanceRepository) GetBalanceByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	query := "SELECT current, withdrawn FROM balances WHERE user_id = $1"
	row := b.db.QueryRowContext(ctx, query, userID)

	var balance domain.Balance
	err := row.Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (b balanceRepository) UpdateBalance(ctx context.Context, userID int64, delta float64) error {
	query := "UPDATE balances SET current = current + $1 WHERE user_id = $2"
	_, err := b.db.ExecContext(ctx, query, delta, userID)
	return err
}

func (b balanceRepository) CreateWithdrawal(ctx context.Context, withdrawal *domain.Withdrawal) error {
	query := "INSERT INTO withdrawals (user_id, order, sum, processed_at) VALUES ($1, $2, $3, $4)"
	_, err := b.db.ExecContext(ctx, query, withdrawal.UserID, withdrawal.Order, withdrawal.Sum, withdrawal.ProcessedAt)
	return err
}

func (b balanceRepository) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]*domain.Withdrawal, error) {
	query := "SELECT order, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC"
	rows, err := b.db.QueryContext(ctx, query, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*domain.Withdrawal

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	for rows.Next() {
		var w domain.Withdrawal
		err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, &w)
	}
	return withdrawals, nil
}

func NewBalanceRepository(db *sql.DB) BalanceRepository {
	return &balanceRepository{db: db}
}
