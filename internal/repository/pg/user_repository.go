package pg

import (
	"context"
	"damirqa/loyalty-system/internal/domain"
	"database/sql"
	"errors"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
}

type userRepository struct {
	db *sql.DB
}

func (u userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := "INSERT INTO users (login, pssword_hash) VALUES ($1, $2)"
	_, err := u.db.ExecContext(ctx, query, user.Login, user.PasswordHash)
	return err
}

func (u userRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	query := "SELECT login, password_hash FROM users WHERE login = $1"
	row := u.db.QueryRowContext(ctx, query, login)

	var user domain.User
	err := row.Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}
