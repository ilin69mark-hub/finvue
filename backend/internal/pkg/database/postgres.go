package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"finvue/internal/pkg/config"
	"finvue/internal/pkg/logger"

	"go.uber.org/zap"
)

var Pool *pgxpool.Pool

func Connect(cfg *config.DatabaseConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return fmt.Errorf("ошибка парсинга DSN: %w", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("ошибка пинга БД: %w", err)
	}

	logger.Info("Подключение к PostgreSQL установлено")
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
		logger.Info("Подключение к PostgreSQL закрыто")
	}
}

func WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			logger.Error("Ошибка отката транзакции", zap.Error(rollbackErr))
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	return nil
}