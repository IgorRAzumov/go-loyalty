package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"loyalty/internal/logger"

	// Регистрируем драйвер pgx для database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	connectTimeout = 5 * time.Second
)

// PoolConfig содержит настройки connection pool для sql.DB.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultPoolConfig возвращает рекомендуемые настройки для production (10K RPS target).
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    100,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
}

// Open открывает соединение с PostgreSQL по DSN, применяет настройки pool и проверяет доступность.
func Open(ctx context.Context, dsn string, poolCfg PoolConfig, log logger.Logger) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("DATABASE_URI is empty")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(poolCfg.MaxOpenConns)
	db.SetMaxIdleConns(poolCfg.MaxIdleConns)
	db.SetConnMaxLifetime(poolCfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(poolCfg.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if log != nil {
		log.Info().
			Int("max_open_conns", poolCfg.MaxOpenConns).
			Int("max_idle_conns", poolCfg.MaxIdleConns).
			Duration("conn_max_lifetime", poolCfg.ConnMaxLifetime).
			Duration("conn_max_idle_time", poolCfg.ConnMaxIdleTime).
			Message("database connection pool configured")
	}
	return db, nil
}

// OpenWithMigrations открывает соединение с PostgreSQL, применяет миграции и настраивает pool.
//
// Важно: golang-migrate может закрывать underlying DB driver при закрытии migrate-инстанса,
// поэтому миграции выполняются на отдельном *sql.DB, после чего соединение закрывается,
// и для работы приложения открывается новое *sql.DB с настроенным connection pool.
func OpenWithMigrations(ctx context.Context, dsn string, poolCfg PoolConfig, log logger.Logger) (*sql.DB, error) {
	migrationPoolCfg := PoolConfig{
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
	migrationsDB, err := Open(ctx, dsn, migrationPoolCfg, nil)
	if err != nil {
		return nil, err
	}
	if err := ApplyMigrations(migrationsDB); err != nil {
		_ = migrationsDB.Close()
		return nil, err
	}
	_ = migrationsDB.Close()

	return Open(ctx, dsn, poolCfg, log)
}
