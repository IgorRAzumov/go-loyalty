package auth

import "testing"

func TestHashAndComparePassword(t *testing.T) {
	hash, err := HashPassword("longenough10")
	if err != nil {
		t.Fatalf("HashPassword err: %v", err)
	}
	if len(hash) == 0 {
		t.Fatalf("expected non-empty hash")
	}

	if err := ComparePassword(hash, "longenough10"); err != nil {
		t.Fatalf("expected match, got err: %v", err)
	}
	if err := ComparePassword(hash, "longenough11"); err == nil {
		t.Fatalf("expected mismatch error")
	}
}
