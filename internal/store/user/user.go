package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/shevchukeugeni/gofermart/internal/store"
	"github.com/shevchukeugeni/gofermart/internal/types"
)

type repo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) store.User {
	return &repo{db: db}
}

func (repo *repo) CreateUser(ctx context.Context, user *types.User) error {
	if user == nil {
		return errors.New("repository: incorrect parameters")
	}

	_, err := repo.db.ExecContext(ctx,
		"INSERT INTO users(id, login, password) VALUES ($1, $2, $3)", user.ID, user.Login, user.Password)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return types.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (repo *repo) GetByLogin(ctx context.Context, login string) (*types.User, error) {
	if login == "" {
		return nil, errors.New("repository: incorrect parameters")
	}

	ret := types.User{Login: login}

	err := repo.db.QueryRowContext(ctx, "SELECT id, password, created_at FROM users WHERE login=$1", login).Scan(
		&ret.ID, &ret.Password, &ret.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
