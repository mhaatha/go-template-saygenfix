package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/go-template-saygenfix/internal/config"
)

func ConnectDB(cfg *config.Config) (*pgxpool.Pool, error) {
	dbPool, err := pgxpool.New(context.Background(), cfg.DBURL)
	if err != nil {
		return nil, err
	}

	// Check DB connection
	err = dbPool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	slog.Info("connected to the database")
	return dbPool, nil
}
