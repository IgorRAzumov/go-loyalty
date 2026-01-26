package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig_EnvAndDefaults(t *testing.T) {
	// Don't run in parallel: mutates os.Args.
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("PORT", "9999")
	t.Setenv("DATABASE_URI", "postgres://x")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_TTL_SECONDS", "")

	os.Args = []string{"cmd"} // no flags
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.RunAddress != ":9999" {
		t.Fatalf("want %q, got %q", ":9999", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://x" {
		t.Fatalf("unexpected DatabaseURI: %q", cfg.DatabaseURI)
	}
	if cfg.AccrualSystemAddress != "http://accrual" {
		t.Fatalf("unexpected AccrualSystemAddress: %q", cfg.AccrualSystemAddress)
	}
	if cfg.JWTSecret == "" {
		t.Fatalf("expected JWTSecret to be generated")
	}
	if cfg.JWTTTL != 24*time.Hour {
		t.Fatalf("expected default ttl 24h, got %v", cfg.JWTTTL)
	}
}

func TestLoadConfig_FlagsOverride(t *testing.T) {
	// Don't run in parallel: mutates os.Args.
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Setenv("RUN_ADDRESS", ":1111")
	t.Setenv("DATABASE_URI", "postgres://env")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://env")
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("JWT_TTL_SECONDS", "3600")

	os.Args = []string{"cmd", "-a", ":2222", "-d", "postgres://flag", "-r", "http://flag"}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.RunAddress != ":2222" {
		t.Fatalf("want %q, got %q", ":2222", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://flag" {
		t.Fatalf("want %q, got %q", "postgres://flag", cfg.DatabaseURI)
	}
	if cfg.AccrualSystemAddress != "http://flag" {
		t.Fatalf("want %q, got %q", "http://flag", cfg.AccrualSystemAddress)
	}
	if cfg.JWTSecret != "s" {
		t.Fatalf("unexpected JWTSecret: %q", cfg.JWTSecret)
	}
	if cfg.JWTTTL != time.Hour {
		t.Fatalf("expected 1h, got %v", cfg.JWTTTL)
	}
}
