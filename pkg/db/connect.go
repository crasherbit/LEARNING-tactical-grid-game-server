package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect establishes a connection to PostgreSQL database
func Connect(databaseURL string) (*pgxpool.Pool, error) {
	slog.Info("Connecting to database")

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	slog.Info("Database connection established")
	return pool, nil
}
