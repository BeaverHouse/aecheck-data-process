package postgres

import (
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/types"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a new database connection pool
func NewPool(cfg types.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// InitFromEnv creates a connection pool from environment variables
func InitFromEnv() (*pgxpool.Pool, error) {
	cfg := types.PostgresConfig{
		Host:     logic.GetEnv("POSTGRES_HOST", "localhost"),
		Port:     logic.GetIntEnv("POSTGRES_PORT", 5432),
		User:     logic.GetEnv("POSTGRES_USER", "postgres"),
		Password: logic.GetEnv("POSTGRES_PASSWORD", ""),
		DBName:   logic.GetEnv("POSTGRES_DBNAME", "postgres"),
		SSLMode:  logic.GetEnv("POSTGRES_SSLMODE", "disable"),
	}

	return NewPool(cfg)
}
