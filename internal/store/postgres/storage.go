package postgres

import (
	"database/sql"

	"go.uber.org/zap"
)

type DBStore struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewStore(logger *zap.Logger, db *sql.DB) *DBStore {
	return &DBStore{
		logger: logger,
		db:     db,
	}
}
