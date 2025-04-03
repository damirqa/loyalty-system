package service

import (
	"context"
	"damirqa/loyalty-system/internal/domain"
	"damirqa/loyalty-system/internal/repository/pg"
	"errors"
	"time"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID int64) (*domain.Balance, error)
	Withdraw(ctx context.Context, userID int64, order string, sum float64) error
	ListWithdrawals(ctx context.Context, userID int64) ([]*domain.Withdrawal, error)
}

type balanceService struct {
	balanceRepo pg.BalanceRepository
	orderRepo   pg.OrderRepository
}

func (b balanceService) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	return b.balanceRepo.GetBalanceByUserID(ctx, userID)
}

func (b balanceService) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	balance, err := b.balanceRepo.GetBalanceByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if balance.Current < sum {
		return errors.New("insufficient funds")
	}

	err = b.balanceRepo.UpdateBalance(ctx, userID, -sum)
	if err != nil {
		return err
	}

	withdrawal := &domain.Withdrawal{
		UserID:      userID,
		Order:       order,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}

	return b.balanceRepo.CreateWithdrawal(ctx, withdrawal)
}

func (b balanceService) ListWithdrawals(ctx context.Context, userID int64) ([]*domain.Withdrawal, error) {
	return b.balanceRepo.GetWithdrawalsByUserID(ctx, userID)
}

func NewBalanceService(balanceRepo pg.BalanceRepository, orderRepo pg.OrderRepository) BalanceService {
	return &balanceService{
		balanceRepo: balanceRepo,
		orderRepo:   orderRepo,
	}
}
