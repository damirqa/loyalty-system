package service

import (
	"context"
	"damirqa/loyalty-system/internal/domain"
	"damirqa/loyalty-system/internal/repository/pg"
	"errors"
	"time"
)

type OrderService interface {
	SubmitOrder(ctx context.Context, userID int64, orderNumber string) error
	ListOrders(ctx context.Context, userID int64) ([]*domain.Order, error)
}

type orderService struct {
	orderRepo pg.OrderRepository
}

func (o orderService) SubmitOrder(ctx context.Context, userID int64, orderNumber string) error {
	if !validateLuhn(orderNumber) {
		return errors.New("invalid order number")
	}

	existing, err := o.orderRepo.GetOrderByID(ctx, orderNumber)
	if err != nil {
		return err
	}

	if existing != nil {
		if existing.UserID == userID {
			return errors.New("order already submitted by user")
		}

		return errors.New("order already submitted by another user")
	}

	order := &domain.Order{
		UserID:     userID,
		Number:     orderNumber,
		Status:     domain.OrderStatusProcessing,
		UploadedAt: time.Now(),
	}

	return o.orderRepo.CreateOrder(ctx, order)
}

func (o orderService) ListOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return o.orderRepo.GetOrdersByUserID(ctx, userID)
}

func NewOrderService(orderRepo pg.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

func validateLuhn(number string) bool {
	var sum int
	alt := false
	nDigits := len(number)
	for i := nDigits - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if alt {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alt = !alt
	}
	return sum%10 == 0
}
