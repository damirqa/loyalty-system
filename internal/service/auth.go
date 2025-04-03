package service

import (
	"context"
	"damirqa/loyalty-system/internal/domain"
	"damirqa/loyalty-system/internal/repository/pg"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, login, password string) (*domain.User, error)
	Login(ctx context.Context, login, password string) (*domain.User, error)
}

type authService struct {
	userRepo pg.UserRepository
}

func NewAuthService(userRepo pg.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (a authService) Register(ctx context.Context, login, password string) (*domain.User, error) {
	existing, err := a.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("user already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Login:        login,
		PasswordHash: string(passwordHash),
	}

	err = a.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a authService) Login(ctx context.Context, login, password string) (*domain.User, error) {
	user, err := a.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}
