package auth

import (
	"errors"
	"strings"
	"testing"

	"loyalty/internal/domain/auth/model"
)

func TestService_ValidateLogin(t *testing.T) {
	t.Parallel()

	svc := NewAuthService()

	got, err := svc.ValidateLogin("  alice  ")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "alice" {
		t.Fatalf("want %q, got %q", "alice", got)
	}

	if _, err := svc.ValidateLogin("   "); err == nil {
		t.Fatalf("expected error")
	}
}

func TestService_ValidatePassword(t *testing.T) {
	t.Parallel()

	svc := NewAuthService()

	if err := svc.ValidatePassword(""); err == nil || !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if err := svc.ValidatePassword("short"); err == nil || !errors.Is(err, model.ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
	if err := svc.ValidatePassword(strings.Repeat("a", 73)); err == nil || !errors.Is(err, model.ErrPasswordTooLong) {
		t.Fatalf("expected ErrPasswordTooLong, got %v", err)
	}
	if err := svc.ValidatePassword("longenough10"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestService_HashAndComparePassword(t *testing.T) {
	t.Parallel()

	svc := NewAuthService()

	hash, err := svc.HashPassword("longenough10")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(hash) == 0 {
		t.Fatalf("expected non-empty hash")
	}

	if err := svc.ComparePassword(hash, "longenough10"); err != nil {
		t.Fatalf("expected match, got err: %v", err)
	}
	if err := svc.ComparePassword(hash, "longenough11"); err == nil {
		t.Fatalf("expected mismatch error")
	}
}
