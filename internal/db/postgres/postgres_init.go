package postgres

import (
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/types"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a new database connection pool
func NewPool(cfg types.PostgresConfig) (*pgxpool.Pool, error) {
	// Add statement_timeout to DSN for global query timeout
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s statement_timeout=30000",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Parse config and set timeouts
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Set connection pool timeouts
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute
	poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
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
