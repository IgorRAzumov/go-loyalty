package auth

import (
	"context"
	"errors"
	"testing"

	"loyalty/internal/domain/auth/model"
	"loyalty/internal/mocks"

	"go.uber.org/mock/gomock"
)

func TestUsecase_Register_HappyPath(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	u := mocks.NewMockUserService(ctrl)
	u.EXPECT().
		CreateUser(gomock.Any(), " alice ", gomock.Any()).
		DoAndReturn(func(_ context.Context, login string, passwordHash []byte) (model.User, error) {
			if login != " alice " {
				t.Fatalf("want raw login %q, got %q", " alice ", login)
			}
			if len(passwordHash) == 0 {
				t.Fatalf("expected password hash")
			}
			return model.User{ID: 7, Login: login}, nil
		})
	u.EXPECT().FindUserByLogin(gomock.Any(), gomock.Any()).MaxTimes(0)

	a := mocks.NewMockAuthService(ctrl)
	a.EXPECT().
		HashPassword("longenough10").
		Return([]byte("hash"), nil)
	a.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).MaxTimes(0)

	ts := mocks.NewMockTokenService(ctrl)
	ts.EXPECT().
		IssueToken(int64(7), " alice ", gomock.Any()).
		Return("token", nil)

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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	u := mocks.NewMockUserService(ctrl)
	u.EXPECT().
		FindUserByLogin(gomock.Any(), "alice").
		Return(model.User{ID: 1, Login: "alice", PasswordHash: []byte("hash")}, nil)
	u.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

	a := mocks.NewMockAuthService(ctrl)
	a.EXPECT().
		ComparePassword([]byte("hash"), "short").
		Return(model.ErrPasswordTooShort)
	a.EXPECT().HashPassword(gomock.Any()).MaxTimes(0)

	ts := mocks.NewMockTokenService(ctrl)
	ts.EXPECT().IssueToken(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

	uc := NewUsecase(u, a, ts)
	_, err := uc.Login(context.Background(), "alice", "short")
	if !errors.Is(err, model.ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestUsecase_Login_WrongPasswordBecomesInvalidCreds(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	u := mocks.NewMockUserService(ctrl)
	u.EXPECT().
		FindUserByLogin(gomock.Any(), "alice").
		Return(model.User{ID: 1, Login: "alice", PasswordHash: []byte("hash")}, nil)
	u.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

	a := mocks.NewMockAuthService(ctrl)
	a.EXPECT().
		ComparePassword([]byte("hash"), "longenough11").
		Return(errors.New("bcrypt: mismatch"))
	a.EXPECT().HashPassword(gomock.Any()).MaxTimes(0)

	ts := mocks.NewMockTokenService(ctrl)
	ts.EXPECT().IssueToken(gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(0)

	uc := NewUsecase(u, a, ts)
	_, err := uc.Login(context.Background(), "alice", "longenough11")
	if !errors.Is(err, model.ErrInvalidCreds) {
		t.Fatalf("expected ErrInvalidCreds, got %v", err)
	}
}
