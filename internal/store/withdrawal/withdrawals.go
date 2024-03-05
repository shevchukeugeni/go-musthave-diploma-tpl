package withdrawal

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shevchukeugeni/gofermart/internal/store"
	"github.com/shevchukeugeni/gofermart/internal/types"
)

type repo struct {
	db     *sql.DB
	orders store.Order
}

func NewRepository(db *sql.DB, orders store.Order) store.Withdrawal {
	return &repo{db: db, orders: orders}
}

func (repo *repo) CreateWithdrawal(ctx context.Context, orderNum, userId string, sum float64) error {
	if userId == "" {
		return errors.New("repository: incorrect parameters")
	}

	balance, err := repo.GetBalance(ctx, userId)
	if err != nil {
		return err
	}

	if balance.Current < sum {
		return types.ErrInsufficientBalance
	}

	_, err = repo.db.ExecContext(ctx,
		"INSERT INTO withdrawals(user_id, number, sum) VALUES ($1, $2, $3)", userId, orderNum, sum)

	return err
}

func (repo *repo) GetBalance(ctx context.Context, userID string) (*types.UserBalance, error) {
	orders, err := repo.orders.GetProcessedOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	balance := new(types.UserBalance)
	for _, order := range orders {
		balance.Current += order.Accrual
	}

	wtdrwls, err := repo.GetWithdrawalsByUser(ctx, userID)

	for _, wtdrw := range wtdrwls {
		balance.Withdrawn += wtdrw.Sum
	}

	balance.Current = balance.Current - balance.Withdrawn

	return balance, nil
}

func (repo *repo) GetWithdrawalsByUser(ctx context.Context, userID string) ([]types.Withdrawal, error) {
	if userID == "" {
		return nil, errors.New("repository: incorrect parameters")
	}

	ret := []types.Withdrawal{}
	rows, err := repo.db.QueryContext(ctx,
		"SELECT number, sum, processed_at FROM withdrawals WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		wtdrw := types.Withdrawal{}
		err := rows.Scan(&wtdrw.Number, &wtdrw.Sum, &wtdrw.ProcessedAt)
		if err != nil {
			return nil, err
		}
		ret = append(ret, wtdrw)
	}

	return ret, nil
}
