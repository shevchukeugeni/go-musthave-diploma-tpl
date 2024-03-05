package store

import (
	"context"

	"github.com/shevchukeugeni/gofermart/internal/types"
)

type User interface {
	CreateUser(ctx context.Context, user *types.User) error
	GetByLogin(ctx context.Context, login string) (*types.User, error)
}

type Order interface {
	CreateOrder(ctx context.Context, orderNum, userID string) error
	UpdateOrder(ctx context.Context, orderNum, status string, accrual float64) error
	GetOrdersByUser(ctx context.Context, userID string) ([]types.Order, error)
	GetProcessedOrdersByUser(ctx context.Context, userID string) ([]types.Order, error)
	GetPendingOrdersNumbers(ctx context.Context) ([]types.Order, error)
}

type Withdrawal interface {
	CreateWithdrawal(ctx context.Context, orderNum, userID string, sum float64) error
	GetBalance(ctx context.Context, userID string) (*types.UserBalance, error)
	GetWithdrawalsByUser(ctx context.Context, userID string) ([]types.Withdrawal, error)
}
