package app

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRun_GracefulShutdown(t *testing.T) {
	// config.LoadConfig parses os.Args; tests run with -test.* flags.
	// Override to avoid unknown flag parsing.
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	os.Args = []string{"cmd"}

	t.Setenv("RUN_ADDRESS", ":0")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("LOG_LEVEL", "")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	if err := Run(ctx); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
