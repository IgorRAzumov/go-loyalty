package memory

import (
	"context"
	"errors"
	"testing"

	"loyalty/internal/domain/auth/model"
)

func TestUserRepository_CreateAndFind(t *testing.T) {
	repo := NewRepository()

	u, err := repo.Create(context.Background(), "alice", []byte("hash"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if u.ID == 0 {
		t.Fatalf("expected id")
	}

	got, err := repo.FindByLogin(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Login != "alice" {
		t.Fatalf("want login alice, got %q", got.Login)
	}
}

func TestUserRepository_Create_Duplicate(t *testing.T) {
	repo := NewRepository()

	_, _ = repo.Create(context.Background(), "alice", []byte("hash"))
	_, err := repo.Create(context.Background(), "alice", []byte("hash2"))
	if err == nil || !errors.Is(err, model.ErrLoginTaken) {
		t.Fatalf("expected ErrLoginTaken, got %v", err)
	}
}

func TestUserRepository_Find_NotFound(t *testing.T) {
	repo := NewRepository()

	_, err := repo.FindByLogin(context.Background(), "missing")
	if err == nil || !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
