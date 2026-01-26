package jwt

import (
	"errors"
	"testing"
	"time"

	"loyalty/internal/domain/auth/model"
)

func TestTokenService_IssueAndParse(t *testing.T) {
	t.Parallel()

	now := time.Now()
	svc := NewTokenService("secret", time.Hour)

	tok, err := svc.IssueToken(123, "alice", now)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tok == "" {
		t.Fatalf("expected non-empty token")
	}

	claim, err := svc.ParseToken(tok)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if claim.UserID != 123 {
		t.Fatalf("want userID %d, got %d", 123, claim.UserID)
	}
	if claim.Login != "alice" {
		t.Fatalf("want login %q, got %q", "alice", claim.Login)
	}
	if claim.ExpiresAt.IsZero() {
		t.Fatalf("expected exp set")
	}
}

func TestTokenService_IssueToken_InvalidUserID(t *testing.T) {
	t.Parallel()

	svc := NewTokenService("secret", time.Hour)
	_, err := svc.IssueToken(0, "alice", time.Now())
	if err == nil || !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
