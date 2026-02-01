package usecase

import "context"

// AuthUsecase описывает бизнес-сценарии аутентификации/регистрации пользователя.
type AuthUsecase interface {
	Register(ctx context.Context, login, password string) (token string, err error)
	Login(ctx context.Context, login, password string) (token string, err error)
}
