package order

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shevchukeugeni/gofermart/internal/store"
	"github.com/shevchukeugeni/gofermart/internal/types"
)

type repo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) store.Order {
	return &repo{db: db}
}

func (repo *repo) CreateOrder(ctx context.Context, orderNum, userId string) error {
	if userId == "" {
		return errors.New("repository: incorrect parameters")
	}

	ret := types.Order{Number: orderNum}
	err := repo.db.QueryRowContext(ctx, "SELECT user_id FROM orders WHERE number=$1", orderNum).Scan(
		&ret.UserID)
	if !errors.Is(err, sql.ErrNoRows) {
		if err == nil {
			if ret.UserID == userId {
				return types.ErrOrderAlreadyCreatedByUser
			} else {
				return types.ErrOrderAlreadyCreatedByAnother
			}
		} else {
			return err
		}
	}

	_, err = repo.db.ExecContext(ctx,
		"INSERT INTO orders(number, user_id, status) VALUES ($1, $2, $3)", orderNum, userId, types.New)
	if err != nil {
		return err
	}
	return nil
}

func (repo *repo) UpdateOrder(ctx context.Context, orderNum, status string, accrual float64) error {
	_, err := repo.db.ExecContext(ctx,
		"UPDATE orders SET status = $1, accrual = $2 WHERE number=$3", status, accrual, orderNum)
	if err != nil {
		return err
	}
	return nil
}

func (repo *repo) GetOrdersByUser(ctx context.Context, userId string) ([]types.Order, error) {
	if userId == "" {
		return nil, errors.New("repository: incorrect parameters")
	}

	ret := []types.Order{}
	rows, err := repo.db.QueryContext(ctx,
		"SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id=$1 ORDER BY uploaded_at DESC", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		order := types.Order{}
		var acc sql.NullFloat64
		err := rows.Scan(&order.Number, &order.Status, &acc, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		order.Accrual = acc.Float64
		ret = append(ret, order)
	}

	return ret, nil
}

func (repo *repo) GetProcessedOrdersByUser(ctx context.Context, userId string) ([]types.Order, error) {
	if userId == "" {
		return nil, errors.New("repository: incorrect parameters")
	}

	ret := []types.Order{}
	rows, err := repo.db.QueryContext(ctx,
		"SELECT number, accrual, uploaded_at FROM orders WHERE user_id=$1 and status=$2 ORDER BY uploaded_at DESC",
		userId, types.Processed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		order := types.Order{UserID: userId}
		var acc sql.NullFloat64
		err := rows.Scan(&order.Number, &acc, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		order.Accrual = acc.Float64
		ret = append(ret, order)
	}

	return ret, nil
}

func (repo *repo) GetPendingOrdersNumbers(ctx context.Context) ([]types.Order, error) {
	ret := []types.Order{}
	rows, err := repo.db.QueryContext(ctx,
		"SELECT number FROM orders WHERE status!=$1 and status != $2",
		types.Invalid, types.Processed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		order := types.Order{}
		err := rows.Scan(&order.Number)
		if err != nil {
			return nil, err
		}
		ret = append(ret, order)
	}

	return ret, nil
}
