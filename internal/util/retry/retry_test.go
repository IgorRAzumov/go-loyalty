package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo_SuccessFirstAttempt(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.MaxAttempts = 3

	n := 0
	got, err := Do(ctx, cfg, func() (int, error) {
		n++
		return 42, nil
	}, func(err error) bool { return true })
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 42 {
		t.Fatalf("want 42, got %d", got)
	}
	if n != 1 {
		t.Fatalf("want 1 attempt, got %d", n)
	}
}

func TestDo_RetryThenSuccess(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.MaxAttempts = 4
	cfg.InitialDelay = 10 * time.Millisecond
	cfg.MaxDelay = 50 * time.Millisecond

	n := 0
	got, err := Do(ctx, cfg, func() (int, error) {
		n++
		if n < 3 {
			return 0, ErrRetryable
		}
		return 100, nil
	}, IsRetryableNetwork)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 100 {
		t.Fatalf("want 100, got %d", got)
	}
	if n != 3 {
		t.Fatalf("want 3 attempts, got %d", n)
	}
}

func TestDo_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.MaxAttempts = 3

	n := 0
	_, err := Do(ctx, cfg, func() (int, error) {
		n++
		return 0, errors.New("fatal")
	}, func(err error) bool { return false })
	if err == nil {
		t.Fatal("expected error")
	}
	if n != 1 {
		t.Fatalf("want 1 attempt (no retry), got %d", n)
	}
}

func TestDo_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := DefaultConfig()
	cfg.MaxAttempts = 5
	cfg.InitialDelay = 100 * time.Millisecond

	n := 0
	_, err := Do(ctx, cfg, func() (int, error) {
		n++
		return 0, ErrRetryable
	}, func(err error) bool { return true })
	if err == nil {
		t.Fatal("expected error")
	}
	if n != 1 {
		t.Fatalf("want 1 attempt (ctx done), got %d", n)
	}
}
