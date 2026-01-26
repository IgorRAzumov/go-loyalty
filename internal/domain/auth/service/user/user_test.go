package user

import (
	"context"
	"testing"

	"loyalty/internal/domain/auth/model"
	"loyalty/internal/domain/auth/repository"
)

type mockRepo struct {
	createFn func(ctx context.Context, login string, passwordHash []byte) (model.User, error)
	findFn   func(ctx context.Context, login string) (model.User, error)
}

func (m *mockRepo) Create(ctx context.Context, login string, passwordHash []byte) (model.User, error) {
	return m.createFn(ctx, login, passwordHash)
}
func (m *mockRepo) FindByLogin(ctx context.Context, login string) (model.User, error) {
	return m.findFn(ctx, login)
}

var _ repository.UserRepository = (*mockRepo)(nil)

func TestUserService_Delegates(t *testing.T) {
	svc := NewUserService(&mockRepo{
		createFn: func(_ context.Context, login string, passwordHash []byte) (model.User, error) {
			if login != "alice" || string(passwordHash) != "h" {
				t.Fatalf("unexpected args")
			}
			return model.User{ID: 1, Login: login}, nil
		},
		findFn: func(_ context.Context, login string) (model.User, error) {
			if login != "alice" {
				t.Fatalf("unexpected login")
			}
			return model.User{ID: 1, Login: login}, nil
		},
	})

	_, _ = svc.CreateUser(context.Background(), " alice ", []byte("h"))
	_, _ = svc.FindUserByLogin(context.Background(), " alice ")
}
