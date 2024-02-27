package worker

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

type Worker struct {
	logger            *zap.Logger
	db                *sql.DB
	accrualSystemAddr string
}

func NewWorker(logger *zap.Logger, db *sql.DB, accrualSystemAddr string) *Worker {
	return &Worker{
		logger:            logger,
		db:                db,
		accrualSystemAddr: accrualSystemAddr,
	}
}

func (w *Worker) Run(ctx context.Context) {
	//TODO:
	// Implement worker logic (periodic update request to Accrual System)
}
