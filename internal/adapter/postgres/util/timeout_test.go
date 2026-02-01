package util

import (
	"context"
	"testing"
	"time"
)

func TestWithQueryTimeout_CreatesDeadline(t *testing.T) {
	SetQueryTimeout(100 * time.Millisecond)

	ctx := context.Background()
	queryCtx, cancel := WithQueryTimeout(ctx)
	defer cancel()

	deadline, ok := queryCtx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}

	expectedDeadline := time.Now().Add(100 * time.Millisecond)
	if deadline.After(expectedDeadline.Add(10 * time.Millisecond)) {
		t.Fatalf("deadline too far in future: %v", deadline)
	}
	if deadline.Before(expectedDeadline.Add(-10 * time.Millisecond)) {
		t.Fatalf("deadline too soon: %v", deadline)
	}
}

func TestWithQueryTimeout_RespectsParentDeadline(t *testing.T) {
	SetQueryTimeout(1 * time.Second)

	parentDeadline := time.Now().Add(50 * time.Millisecond)
	parentCtx, parentCancel := context.WithDeadline(context.Background(), parentDeadline)
	defer parentCancel()

	queryCtx, cancel := WithQueryTimeout(parentCtx)
	defer cancel()

	deadline, ok := queryCtx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}

	if deadline.After(parentDeadline.Add(10 * time.Millisecond)) {
		t.Fatalf("expected parent deadline to be respected, got %v, want ~%v", deadline, parentDeadline)
	}
}

func TestSetQueryTimeout_IgnoresZero(t *testing.T) {
	original := queryTimeout
	defer func() { queryTimeout = original }()

	SetQueryTimeout(100 * time.Millisecond)
	if queryTimeout != 100*time.Millisecond {
		t.Fatalf("expected timeout to be set to 100ms, got %v", queryTimeout)
	}

	SetQueryTimeout(0)
	if queryTimeout != 100*time.Millisecond {
		t.Fatalf("expected timeout to remain 100ms when setting to 0, got %v", queryTimeout)
	}

	SetQueryTimeout(-1 * time.Second)
	if queryTimeout != 100*time.Millisecond {
		t.Fatalf("expected timeout to remain 100ms when setting to negative, got %v", queryTimeout)
	}
}
