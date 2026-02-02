package postgres

import (
	"context"
	"loyalty/internal/adapter/postgres/util"
	"testing"
	"time"
)

func TestSetQueryTimeout(t *testing.T) {
	timeout := 5 * time.Second
	util.SetQueryTimeout(timeout)
	// No panic = success (функция просто устанавливает глобальную переменную)
}

func TestPoolConfig(t *testing.T) {
	cfg := PoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}

	if cfg.MaxOpenConns != 10 {
		t.Errorf("MaxOpenConns = %v, want 10", cfg.MaxOpenConns)
	}
}

func TestDefaultPoolConfig(t *testing.T) {
	cfg := DefaultPoolConfig()
	if cfg.MaxOpenConns <= 0 {
		t.Error("MaxOpenConns should be > 0")
	}
	if cfg.MaxIdleConns <= 0 {
		t.Error("MaxIdleConns should be > 0")
	}
	if cfg.ConnMaxLifetime <= 0 {
		t.Error("ConnMaxLifetime should be > 0")
	}
	if cfg.ConnMaxIdleTime <= 0 {
		t.Error("ConnMaxIdleTime should be > 0")
	}
}

func TestOpen_EmptyDSN(t *testing.T) {
	_, err := Open(context.Background(), "", DefaultPoolConfig())
	if err == nil {
		t.Error("expected error for empty DSN")
	}
}

func TestOpen_InvalidDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := Open(ctx, "invalid://dsn", DefaultPoolConfig())
	if err == nil {
		t.Error("expected error for invalid DSN")
	}
}
