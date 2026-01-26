package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"loyalty/internal/domain/auth/model"
	"loyalty/internal/domain/auth/service"
)

type mockUserService struct {
	createFn func(ctx context.Context, login string, passwordHash []byte) (model.User, error)
	findFn   func(ctx context.Context, login string) (model.User, error)
}

func (m *mockUserService) CreateUser(ctx context.Context, login string, passwordHash []byte) (model.User, error) {
	return m.createFn(ctx, login, passwordHash)
}
func (m *mockUserService) FindUserByLogin(ctx context.Context, login string) (model.User, error) {
	return m.findFn(ctx, login)
}

type mockAuthService struct {
	hashPasswordFn    func(password string) ([]byte, error)
	comparePasswordFn func(hash []byte, password string) error
}

func (m *mockAuthService) ValidateLogin(string) (string, error)   { panic("not used") }
func (m *mockAuthService) ValidatePassword(password string) error { panic("not used") }
func (m *mockAuthService) HashPassword(password string) ([]byte, error) {
	return m.hashPasswordFn(password)
}
func (m *mockAuthService) ComparePassword(hash []byte, password string) error {
	return m.comparePasswordFn(hash, password)
}

var _ service.AuthService = (*mockAuthService)(nil)

type mockTokenService struct {
	issueFn func(userID int64, login string, now time.Time) (string, error)
}

func (m *mockTokenService) IssueToken(userID int64, login string, now time.Time) (string, error) {
	return m.issueFn(userID, login, now)
}
func (m *mockTokenService) ParseToken(string) (*model.Claim, error) { panic("not used") }

var _ service.TokenService = (*mockTokenService)(nil)

func TestUsecase_Register_HappyPath(t *testing.T) {
	t.Parallel()

	u := &mockUserService{
		createFn: func(_ context.Context, login string, passwordHash []byte) (model.User, error) {
			if login != " alice " {
				t.Fatalf("want raw login %q, got %q", " alice ", login)
			}
			if len(passwordHash) == 0 {
				t.Fatalf("expected password hash")
			}
			return model.User{ID: 7, Login: login}, nil
		},
		findFn: func(context.Context, string) (model.User, error) { panic("not used") },
	}

	a := &mockAuthService{
		hashPasswordFn: func(password string) ([]byte, error) {
			if password != "longenough10" {
				t.Fatalf("unexpected password: %q", password)
			}
			return []byte("hash"), nil
		},
		comparePasswordFn: func([]byte, string) error { panic("not used") },
	}

	ts := &mockTokenService{
		issueFn: func(userID int64, login string, _ time.Time) (string, error) {
			if userID != 7 || login != " alice " {
				t.Fatalf("unexpected args: %d %q", userID, login)
			}
			return "token", nil
		},
	}

	uc := NewUsecase(u, a, ts)
	got, err := uc.Register(context.Background(), " alice ", "longenough10")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "token" {
		t.Fatalf("want %q, got %q", "token", got)
	}
}

func TestUsecase_Login_InvalidPasswordIsBadRequest(t *testing.T) {
	t.Parallel()

	u := &mockUserService{
		createFn: func(context.Context, string, []byte) (model.User, error) { panic("not used") },
		findFn: func(_ context.Context, login string) (model.User, error) {
			if login != "alice" {
				t.Fatalf("unexpected login %q", login)
			}
			return model.User{ID: 1, Login: login, PasswordHash: []byte("hash")}, nil
		},
	}

	a := &mockAuthService{
		hashPasswordFn: func(string) ([]byte, error) { panic("not used") },
		comparePasswordFn: func([]byte, string) error {
			return model.ErrPasswordTooShort
		},
	}

	uc := NewUsecase(u, a, &mockTokenService{issueFn: func(int64, string, time.Time) (string, error) { panic("not used") }})
	_, err := uc.Login(context.Background(), "alice", "short")
	if !errors.Is(err, model.ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestUsecase_Login_WrongPasswordBecomesInvalidCreds(t *testing.T) {
	t.Parallel()

	u := &mockUserService{
		createFn: func(context.Context, string, []byte) (model.User, error) { panic("not used") },
		findFn: func(_ context.Context, login string) (model.User, error) {
			if login != "alice" {
				t.Fatalf("unexpected login %q", login)
			}
			return model.User{ID: 1, Login: login, PasswordHash: []byte("hash")}, nil
		},
	}

	a := &mockAuthService{
		hashPasswordFn: func(string) ([]byte, error) { panic("not used") },
		comparePasswordFn: func([]byte, string) error {
			return errors.New("bcrypt: mismatch")
		},
	}

	uc := NewUsecase(u, a, &mockTokenService{issueFn: func(int64, string, time.Time) (string, error) { panic("not used") }})
	_, err := uc.Login(context.Background(), "alice", "longenough11")
	if !errors.Is(err, model.ErrInvalidCreds) {
		t.Fatalf("expected ErrInvalidCreds, got %v", err)
	}
}
